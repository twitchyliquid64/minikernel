 {config, pkgs, boot, ...}:
{
  imports = [
    <nixpkgs/nixos/modules/installer/scan/detected.nix>
    <nixpkgs/nixos/modules/installer/scan/not-detected.nix>
    <nixpkgs/nixos/modules/profiles/all-hardware.nix>
    <nixpkgs/nixos/modules/profiles/base.nix>
  ];

  # Make assertions pass
  boot.loader.grub.enable = false;
  boot.loader.systemd-boot.enable = true;
  fileSystems = {
    # Mounts whatever device has the NIXOS_ROOT label on it as /
    "/".label = "NIXOS_ROOT";
  };

  # Basic configuration which is not related to an installation mode.
  users.users.root.password = "xxx";
  networking.dhcpcd.enable = false;

  # Basic localization
  console = {
	keyMap = "us";
  };
  i18n = {
	defaultLocale = "en_US.UTF-8";
	supportedLocales = [ (config.i18n.defaultLocale + "/UTF-8") ];
  };
  time.timeZone = "America/Los_Angeles";

  environment.extraInit = ''
    unset GOROOT
    unset GOPATH
    unset GOBIN
  '';

  boot.initrd.supportedFilesystems = [ "squashfs" "overlay" "vfat" "ext4" "ntfs" "f2fs" ];
  boot.initrd.availableKernelModules = [ "squashfs" "overlay" ] ++
        [
          "virtio_net" "virtio_pci" "virtio_mmio"
          "virtio_blk" "virtio_scsi" "9p" "9pnet_virtio"
          "virtio_rng" "crc32c_generic"
        ]; # VM guests

  boot.initrd.kernelModules = [ "loop" "squashfs" "overlay" ];
}
