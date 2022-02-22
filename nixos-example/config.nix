 {config, pkgs, boot, ...}:
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


  # Random crap
  environment.systemPackages = with pkgs; [
    htop sudo
  ];
}