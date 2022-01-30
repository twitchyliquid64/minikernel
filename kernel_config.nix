{pkgs}:
with pkgs.lib.kernel; {
      # Base subsystems
      SCSI = yes;
      PM = yes;
      PCI = yes;
      ATA = yes;
      MD = yes;
      SERIO = yes;

      TTY = yes;
      VT = yes;

      SYSVIPC = yes;
      SYSVIPC_SYSCTL = yes;
      POSIX_MQUEUE = yes;
      POSIX_MQUEUE_SYSCTL = yes;
      CROSS_MEMORY_ATTACH = yes;
      BLK_DEV_INITRD = yes;

      RELOCATABLE = yes;
      LEGACY_VSYSCALL_EMULATE = yes;
      UEVENT_HELPER = yes;
      DEVTMPFS = yes;
      DEVTMPFS_MOUNT = yes;
      CONNECTOR = yes;
      PROC_EVENTS = yes;

      # Containment / audit subsystems
      AUDIT = yes;
      HAVE_ARCH_AUDITSYSCALL = yes;
      AUDITSYSCALL = yes;
      BPF_SYSCALL = yes;
      CGROUPS = yes;
      MEMCG = yes;
      BLK_CGROUP = yes;
      CGROUP_SCHED = yes;
      CGROUP_DEVICE = yes;
      CGROUP_BPF = yes;
      UTS_NS = yes;
      TIME_NS = yes;
      IPC_NS = yes;
      USER_NS = yes;
      PID_NS = yes;
      SECCOMP = yes;
      SECCOMP_FILTER = yes;
      VETH = yes;
      TUN = yes;
      VXLAN = yes;
      IPVLAN = yes;

      GENERIC_IRQ_PROBE = yes;
      GENERIC_IRQ_SHOW = yes;
      GENERIC_IRQ_EFFECTIVE_AFF_MASK = yes;
      GENERIC_PENDING_IRQ = yes;
      GENERIC_IRQ_MIGRATION = yes;
      HARDIRQS_SW_RESEND = yes;
      IRQ_DOMAIN = yes;
      IRQ_DOMAIN_HIERARCHY = yes;
      GENERIC_IRQ_MATRIX_ALLOCATOR = yes;
      GENERIC_IRQ_RESERVATION_MODE = yes;
      IRQ_FORCED_THREADING = yes;
      SPARSE_IRQ = yes;

      TICK_ONESHOT = yes;
      NO_HZ_COMMON = yes;
      NO_HZ_IDLE = yes;
      NO_HZ = yes;
      HIGH_RES_TIMERS = yes;
      TICK_CPU_ACCOUNTING = yes;

      SLUB = yes;
      SLAB_MERGE_DEFAULT = yes;
      SLAB_FREELIST_HARDENED = yes;
      SLUB_CPU_PARTIAL = yes;
      DMADEVICES = yes;
      SYNC_FILE = yes;

      SMP = yes;
      RETPOLINE = yes;
      X86_MSR = yes;
      X86_CPUID = yes;
      GENERIC_CPU = yes;
      HOTPLUG_CPU = yes;

      ACPI = yes;
      ACPI_LEGACY_TABLES_LOOKUP = yes;

      BLOCK = yes;
      BLK_DEV = yes;
      BLK_DEV_SD = yes;
      BLK_DEV_SR = yes;
      CHR_DEV_SG = yes;
      ISCSI_TCP = yes;
      SCSI_LOWLEVEL = yes;

      PARTITION_ADVANCED = yes;
      MSDOS_PARTITION = yes;

      INPUT = yes;
      INPUT_KEYBOARD = yes;
      KEYBOARD_ATKBD = yes;
      INPUT_MOUSE = yes;
      INPUT_MOUSEDEV = yes;
      INPUT_EVDEV = yes;

      # Network
      NET = yes;
      NET_INGRESS = yes;
      PACKET = yes;
      UNIX = yes;
      UNIX_SCM = yes;
      INET = yes;
      IP_MULTICAST = yes;
      IP_ADVANCED_ROUTER = yes;
      IP_MULTIPLE_TABLES = yes;
      IP_ROUTE_MULTIPATH = yes;
      IP_ROUTE_VERBOSE = yes;
      IP_PNP = yes;
      IP_PNP_DHCP = yes;
      IP_PNP_BOOTP = yes;
      IP_PNP_RARP = yes;
      SYN_COOKIES = yes;
      TCP_CONG_ADVANCED = yes;
      TCP_CONG_CUBIC = yes;
      TCP_MD5SIG = yes;
      IPV6 = yes;
      IPV6_ROUTER_PREF = yes;
      IPV6_ROUTE_INFO = yes;
      IPV6_OPTIMISTIC_DAD = yes;
      IPV6_MULTIPLE_TABLES = yes;
      IPV6_SUBTREES = yes;
      IPV6_MROUTE = yes;
      IPV6_MROUTE_MULTIPLE_TABLES = yes;
      NETLABEL = yes;
      NETWORK_SECMARK = yes;
      NET_PTP_CLASSIFY = yes;
      NETWORK_PHY_TIMESTAMPING = yes;
      IP_SET = yes;
      WIREGUARD = yes;
      BLK_DEV_NBD = yes;
      NET_9P = yes;

      NF_TABLES = yes;
      NETFILTER = yes;
      NETFILTER_ADVANCED = yes;
      NETFILTER_INGRESS = yes;
      NETFILTER_EGRESS = yes;
      NETFILTER_NETLINK_LOG = yes;
      NETFILTER_XTABLES = yes;
      NF_CONNTRACK = yes;
      NF_LOG_SYSLOG = yes;
      NF_SOCKET_IPV4 = yes;
      NF_LOG_IPV4 = yes;
      NF_REJECT_IPV4 = yes;
      NF_SOCKET_IPV6 = yes;
      NF_REJECT_IPV6 = yes;
      NF_LOG_IPV6 = yes;
      IP_NF_IPTABLES = yes;
      IP6_NF_IPTABLES = yes;
      BPFILTER = yes;
      # TODO: Rest of netfilter


      NETDEVICES = yes;
      NET_CORE = yes;

      SECURITY = yes;
      SECURITYFS = yes;
      SECURITY_NETWORK = yes;
      SECURITY_PATH = yes;
      SECURITY_LANDLOCK = yes;

      PRINTK = yes;
      SERIAL_8250 = yes;
      SERIAL_8250_CONSOLE = yes;
      SERIAL_DEV_BUS = yes;
      SERIAL_DEV_CTRL_TTYPORT = yes;
      VT_CONSOLE = yes;
      HW_CONSOLE = yes;
      VT_HW_CONSOLE_BINDING = yes;

      # Virtual devices
      KVM_GUEST = yes;
      HYPERVISOR_GUEST = yes;
      PARAVIRT = yes;
      PARAVIRT_SPINLOCKS = yes;
      PARAVIRT_TIME_ACCOUNTING = yes;
      PARAVIRT_CLOCK = yes;

      VIRTIO = yes;
      VIRT_DRIVERS = yes;
      STAGING = yes; # Enable 'staging' drivers
      VIRTIO_MENU = yes;
      X86_PLATFORM_DEVICES = yes;
      VIRTIO_FS = yes;
      VIRTIO_PCI = yes;
      VIRTIO_PCI_LEGACY = yes;
      VIRTIO_MMIO = yes;
      VIRTIO_MMIO_CMDLINE_DEVICES = yes;
      VIRTIO_BLK = yes;
      VIRTIO_CONSOLE = yes;
      BLK_MQ_VIRTIO = yes;
      SCSI_VIRTIO = yes;
      VSOCKETS = yes;
      VIRTIO_VSOCKETS = yes;
      VIRTIO_VSOCKETS_COMMON = yes;
      VIRTIO_BALLOON = yes;

      VIRTIO_NET = yes;

      PTP_1588_CLOCK = yes;
      PTP_1588_CLOCK_KVM = yes;

      # Useful kernel features
      BLK_DEV_LOOP = yes;

      # Enable correctness checks
      COMPILE_TEST = no;
      WERROR = yes;

      # Prevent explosions
      #MODULES = yes;

      # Filesystems
      EXT2_FS = yes;
      EXT3_FS = yes;
      EXT4_FS = yes;
      EXT4_FS_POSIX_ACL = yes;
      EXT4_FS_SECURITY = yes;
      FUSE_FS = yes;
      CUSE = yes;
      OVERLAY_FS = yes;
      FS_ENCRYPTION = yes;
      FSNOTIFY = yes;
      DNOTIFY = yes;
      INOTIFY_USER = yes;
      FANOTIFY = yes;
      MSDOS_FS = yes;
      VFAT_FS = yes;
      EXFAT_FS = yes;

      PROC_FS = yes;
      PROC_CHILDREN = yes;
      KERNFS = yes;
      SYSFS = yes;
      TMPFS = yes;
      TMPFS_POSIX_ACL = yes;
      TMPFS_XATTR = yes;
      HUGETLBFS = yes;
      HUGETLB_PAGE = yes;
      MEMFD_CREATE = yes;

      MISC_FILESYSTEMS = yes;
      SQUASHFS = yes;
      SQUASHFS_FILE_CACHE = yes;
      SQUASHFS_DECOMP_SINGLE = yes;
      SQUASHFS_XATTR = yes;
      SQUASHFS_ZLIB = yes;
      SQUASHFS_LZ4 = yes;
      SQUASHFS_XZ = yes;
      SQUASHFS_ZSTD = yes;


      # Performance / hardening
      JUMP_LABEL = yes;
    }