package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/scanner"
)

// reads the set of kconfig variables and their values into a map.
func readAnswers() map[string]string {
	out := make(map[string]string, 256)
	conf, err := os.Open(os.Getenv("KERNEL_CONFIG"))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer conf.Close()

	var s scanner.Scanner
	s.Init(conf)
	s.Filename = os.Getenv("KERNEL_CONFIG")

	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		configName := s.TokenText()

		tok = s.Scan()
		configVal := s.TokenText()

		// fmt.Printf("%s: %s = %s\n", s.Position, configName, configVal)
		out[configName] = configVal
	}

	return out
}

type mapping struct {
	Choice int
	Name   string
	ID     string
}

// query describes parsed information from a single prompt,
// emitted by `make config`.
type query struct {
	ID       string // corresponds to a Kconfig ID, eg: VIRTIO_MEM
	Question string // prompt text

	Default        string // The default choice if no response was provided
	Multichoice    bool
	ChoiceMappings []mapping
	Choices        []string // choices, if any
}

// TODO: Holy shit this is actually worse than regex somehow. Fix!
func parseQuery(output string, idx int) (query, error) {
	var out query

	// Parse out any options
	optStart := strings.LastIndex(output[:idx], "[")
	if optStart < 0 {
		return query{}, errors.New("no options presented")
	}
	optEnd := strings.LastIndex(output[:idx], "]")

	s := output[optStart+1 : optEnd]
	switch {
	case s == "" || s == "(none)":
		// Settable string value
		out.Choices = nil

	case strings.Contains(s, "-") && len(s) < 7:
		// Its one of those multiple-choice options
		out.Multichoice = true
		start, err := strconv.ParseUint(s[:strings.Index(s, "-")], 10, 64)
		if err != nil {
			return query{}, fmt.Errorf("cannot parse start integer for multichoice: %v", err)
		}
		end, err := strconv.ParseUint(s[strings.Index(s, "-")+1:len(s)-1], 10, 64) // one less for the question mark
		if err != nil {
			return query{}, fmt.Errorf("cannot parse end integer for multichoice: %v", err)
		}

		for i := start; i <= end; i++ {
			out.Choices = append(out.Choices, fmt.Sprint(i))
		}

	default:
		// Something to be enabled/disabled/module.
		out.Choices = strings.Split(s, "/")
	}

	if out.Multichoice {
		// Identify the multichoice mappings
		i := strings.LastIndex(output[:optStart], "\n")
		for i > 0 {
			o := strings.LastIndex(output[:i], "\n")
			s := strings.TrimSuffix(output[o+1:i], " (NEW)")

			if (strings.HasPrefix(s, "> ") || strings.HasPrefix(s, "  ") && strings.Contains(s, ". ")) {
				fmt.Println(s)
				s2 := strings.TrimLeft(s[0:strings.Index(s, ".")], " >")
				c, err := strconv.ParseUint(s2, 10, 64)
				if err != nil {
					return query{}, fmt.Errorf("cannot parse choice for %q: %v", s, err)
				}
				out.ChoiceMappings = append(out.ChoiceMappings, mapping{
					Choice: int(c),
					ID:     s[strings.LastIndex(s, "(")+1 : strings.LastIndex(s, ")")],
					Name:   s[strings.Index(s, ".")+2 : strings.LastIndex(s, " (")],
				})

				if strings.HasPrefix(strings.TrimSpace(s), "> ") {
					out.Default = fmt.Sprint(c)
				}
			} else {
				if strings.TrimSpace(s) == "*" {
					break
				}
				out.Question = strings.TrimSpace(s)
			}
			i = o
		}

	} else {
		// Identify a default if there was one
		for i := range out.Choices {
			if c := out.Choices[i]; c == "Y" || c == "N" || c == "M" {
				out.Choices[i] = strings.ToLower(c)
				out.Default = out.Choices[i]
			}
		}

		// Parse out the Kconfig identifer
		idStart := strings.LastIndex(output[:optStart], "(")
		idEnd := strings.LastIndex(output[:optStart], ")")
		if idStart < 0 || idEnd < 0 {
			return query{}, errors.New("could not find Kconfig identifier")
		}
		out.ID = output[idStart+1 : idEnd]

		if idx := strings.LastIndex(output[:idStart], "\n"); idx >= 0 {
			out.Question = output[idx+1 : idStart-1]
		}
	}

	return out, nil
}

type answerOracle struct {
	want   map[string]string
	answer *io.PipeWriter

	temp bytes.Buffer
}

func (a *answerOracle) answerMultichoice(q query) {
	answered := false
	for _, choice := range q.ChoiceMappings {
		if _, knowAnswer := a.want[choice.ID]; knowAnswer {
			fmt.Fprintf(os.Stderr, "Answering %s (%d)\n", choice.ID, choice.Choice)
			a.answer.Write([]byte(fmt.Sprint(choice.Choice) + "\n"))
			answered = true
			break
		}
	}
	if !answered && q.Default != "" {
		fmt.Fprintf(os.Stderr, "Answering default %s\n", q.Default)
		a.answer.Write([]byte("\n"))
	}
}

func (a *answerOracle) answerScalar(q query) {
	if answer, knowAnswer := a.want[q.ID]; knowAnswer {
		a.answer.Write([]byte(answer + "\n"))
		fmt.Fprintf(os.Stderr, "%s: Answering %s\n", q.ID, answer)
	} else {
		// answer unknown

		if len(q.Choices) == 0 {
			fmt.Fprintf(os.Stderr, "Question %s has no choices: presuming string option with acceptable empty value\n", q.ID)
			a.answer.Write([]byte("\n"))
		} else if q.Default != "" {
			fmt.Fprintf(os.Stderr, "Answering default %q for %s\n", q.Default, q.ID)
			a.answer.Write([]byte("\n"))
		} else {
			fmt.Fprintf(os.Stderr, "Answering default for %s\n", q.ID)
			a.answer.Write([]byte("\n"))
		}
	}
}

func (a *answerOracle) Write(b []byte) (int, error) {
	n, err := a.temp.Write(b)
	if err != nil {
		return n, err
	}

	spl := strings.Split(string(b), "\n")
	for _, line := range spl {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fmt.Fprintf(os.Stderr, "Got: %s\n", line)
	}

	s := a.temp.String()
	if idx := strings.LastIndex(s, " ###"); idx > 0 {
		q, err := parseQuery(s, idx)
		if err != nil {
			panic(err)
		}

		//fmt.Fprintf(os.Stderr, "%+v\n", q)

		if q.Multichoice {
			fmt.Fprintf(os.Stderr, "\n(multichoice) %s?\n", q.Question)
			for _, c := range q.ChoiceMappings {
				fmt.Fprintf(os.Stderr, "%d. %s (%s)\n", c.Choice, c.Name, c.ID)
			}

			a.answerMultichoice(q)
		} else {
			fmt.Fprintf(os.Stderr, "\n(%s) %s?\n", q.ID, q.Question)
			a.answerScalar(q)
		}
		a.temp.Truncate(0)
	} else {
		// TODO: Printout here
	}

	return n, nil
}

func main() {
	var (
		want        = readAnswers()
		read, write = io.Pipe()
	)

	makeConfig := exec.Command("make", "-C", os.Getenv("SRC"), "O="+os.Getenv("BUILD_ROOT"), "config", "SHELL=bash",
		"ARCH="+os.Getenv("ARCH"), "CC="+os.Getenv("CC"), "HOSTCC="+os.Getenv("HOSTCC"), "HOSTCXX="+os.Getenv("HOSTCXX"))
	if mf := os.Getenv("MAKE_FLAGS"); mf != "" {
		makeConfig.Args = append(makeConfig.Args, mf)
	}

	makeConfig.Stdin = read
	makeConfig.Stderr = os.Stderr
	makeConfig.Stdout = &answerOracle{
		want:   want,
		answer: write,
	}

	if err := makeConfig.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "%Err: %v\n", err)
		os.Exit(1)
	}

	p, _ := os.FindProcess(makeConfig.Process.Pid)
	p.Wait()
}
