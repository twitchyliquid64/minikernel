 {config, pkgs, boot, lib, ...}:
{
  # Make assertions pass
  boot.loader.grub.enable = false;
  boot.loader.systemd-boot.enable = true;
  fileSystems = {
    "/".label = "NIXOS_ROOT";
  };
  networking.firewall.enable = false;

  services.getty = {
    greetingLine = ''<<< Booting into the minikernel nixos example! (\m) - \l >>>'';
    autologinUser = "xxx";
  };
  users.users.xxx = {
    isNormalUser = true;
    extraGroups = [ "wheel" "networkmanager" "video" ];
    initialHashedPassword = "";
  };
  security.sudo = {
    enable = true;
    wheelNeedsPassword = false;
  };


  networking.dhcpcd.enable = false;
  console.keyMap = "us";
  i18n = {
		defaultLocale = "en_US.UTF-8";
    supportedLocales = [ (config.i18n.defaultLocale + "/UTF-8") ];
	};
	time.timeZone = "America/Los_Angeles";


  # Fonts
  fonts.fontDir.enable = true;
  fonts.fonts = with pkgs; [
    freefont_ttf
    liberation_ttf
  ];

  # Random crap
  environment.etc.swayconf = {
    mode = "0644";
    text = ''
      exec wayvnc 0.0.0.0 8080
      bindsym o+Return exec foot
    '';
  };

  system.activationScripts.etc = lib.stringAfter [ "users" "groups" ]
    ''
    echo 'if [[ $(tty) == "/dev/ttyS0" ]]; then' >> /home/xxx/.bash_login
    echo '  sleep 2 && headless' >> /home/xxx/.bash_login
    echo 'fi' >> /home/xxx/.bash_login
    '';

  environment.systemPackages = with pkgs; [
    htop sudo sway-unwrapped wayvnc foot chromium

    (
      pkgs.writeShellScriptBin "headless" ''
      export WLR_BACKENDS=headless
      export WLR_RENDERER=pixman
      export WLR_LIBINPUT_NO_DEVICES=1
      export WAYLAND_DISPLAY=wayland-1
      export XDG_RUNTIME_DIR=/tmp
      export XDG_SESSION_TYPE=wayland
      export WLR_RENDERER_ALLOW_SOFTWARE=1

      exec ${sway-unwrapped}/bin/sway -c /etc/swayconf
      ''
    )
  ];
}