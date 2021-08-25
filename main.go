package main

import (
	"io"
	l "log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Code-Hex/vz"
	"github.com/code-ready/machine/drivers/hyperkit"
	"github.com/code-ready/machine/libmachine/drivers"
	"github.com/kr/pty"
	"github.com/pkg/term/termios"
	"golang.org/x/sys/unix"
)

var vmConfig = hyperkit.Driver{
	VMDriver: &drivers.VMDriver{
		ImageSourcePath: "/Users/teuf/.crc/cache/crc_hyperkit_4.8.4/crc.raw.img",
		ImageFormat:     "raw", // must be 'raw'
		Memory:          1 * 1024 * 1024 * 1024,
		CPU:             4,
	},

	VmlinuzPath:   "/Users/teuf/.crc/cache/crc_hyperkit_4.8.4/vmlinuz-4.18.0-305.10.2.el8_4.x86_64",
	InitrdPath:    "/Users/teuf/.crc/cache/crc_hyperkit_4.8.4/initramfs-4.18.0-305.10.2.el8_4.x86_64.img",
	KernelCmdLine: "console=hvc0 rd.udev.debug rd.debug irqfixup " + "BOOT_IMAGE=(hd0,gpt3)/ostree/rhcos-0f2014cf018bafd35ec93f5b8813b2d105c002f6d998c42f8ec7792e5f2b933b/vmlinuz-4.18.0-305.10.2.el8_4.x86_64 random.trust_cpu=on  ignition.platform.id=qemu ostree=/ostree/boot.1/rhcos/0f2014cf018bafd35ec93f5b8813b2d105c002f6d998c42f8ec7792e5f2b933b/0 root=UUID=d74a2195-33c4-440e-bbbe-9e3fa50953e6 rw rootflags=prjquota",

	// Need to be supported?
	UUID:       "",
	VpnKitSock: "",
	VpnKitUUID: "",
	VSockPorts: []string{},
	VMNet:      false,
}

var log *l.Logger

func setNonCanonicalMode(f *os.File) {
	var attr unix.Termios

	// Get settings for terminal
	termios.Tcgetattr(f.Fd(), &attr)

	// Disable cannonical mode （&^ AND NOT)
	attr.Lflag &^= syscall.ICANON

	// Set minimum characters when reading = 1 char
	attr.Cc[syscall.VMIN] = 1

	// set timeout when reading as non-canonical mode
	attr.Cc[syscall.VTIME] = 0

	// reflects the changed settings
	termios.Tcsetattr(f.Fd(), termios.TCSANOW, &attr)
}

func main() {

	// 238 57
	// width, height, err := terminal.GetSize(int(os.Stdout.Fd()))
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// fmt.Println(width, height)
	// return

	file, err := os.Create("./log.log")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	log = l.New(file, "", l.LstdFlags)

	bootLoader := vz.NewLinuxBootLoader(
		vmConfig.VmlinuzPath,
		vz.WithCommandLine(vmConfig.KernelCmdLine),
		vz.WithInitrd(vmConfig.InitrdPath),
	)
	log.Println("bootLoader:", bootLoader)

	config := vz.NewVirtualMachineConfiguration(
		bootLoader,
		uint(vmConfig.CPU),
		uint64(vmConfig.Memory),
	)

	setNonCanonicalMode(os.Stdin)

	ptmx, tty, err := pty.Open()
	if err != nil {
		panic(err)
	}
	defer ptmx.Close()
	defer tty.Close()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if err := pty.InheritSize(os.Stdout, ptmx); err != nil {
				log.Printf("error resizing pty: %s", err)
			}
		}
	}()
	ch <- syscall.SIGWINCH // Initial resize.

	go func() {
		_, err := io.Copy(os.Stdout, ptmx)
		if err != nil {
			log.Println("pty stdout err", err)
		}
	}()

	log.Println("pty: ", tty.Name())

	// console
	serialPortAttachment := vz.NewFileHandleSerialPortAttachment(os.Stdin, tty)
	consoleConfig := vz.NewVirtioConsoleDeviceSerialPortConfiguration(serialPortAttachment)
	config.SetSerialPortsVirtualMachineConfiguration([]*vz.VirtioConsoleDeviceSerialPortConfiguration{
		consoleConfig,
	})

	// network
	natAttachment := vz.NewNATNetworkDeviceAttachment()
	networkConfig := vz.NewVirtioNetworkDeviceConfiguration(natAttachment)
	config.SetNetworkDevicesVirtualMachineConfiguration([]*vz.VirtioNetworkDeviceConfiguration{
		networkConfig,
	})

	// entropy
	entropyConfig := vz.NewVirtioEntropyDeviceConfiguration()
	config.SetEntropyDevicesVirtualMachineConfiguration([]*vz.VirtioEntropyDeviceConfiguration{
		entropyConfig,
	})

	// disk
	diskImageAttachment, err := vz.NewDiskImageStorageDeviceAttachment(
		vmConfig.ImageSourcePath,
		false,
	)
	if err != nil {
		log.Fatal(err)
	}
	storageDeviceConfig := vz.NewVirtioBlockDeviceConfiguration(diskImageAttachment)
	config.SetStorageDevicesVirtualMachineConfiguration([]vz.StorageDeviceConfiguration{
		storageDeviceConfig,
	})

	// traditional memory balloon device which allows for managing guest memory. (optional)
	config.SetMemoryBalloonDevicesVirtualMachineConfiguration([]vz.MemoryBalloonDeviceConfiguration{
		vz.NewVirtioTraditionalMemoryBalloonDeviceConfiguration(),
	})

	// socket device (optional)
	config.SetSocketDevicesVirtualMachineConfiguration([]vz.SocketDeviceConfiguration{
		vz.NewVirtioSocketDeviceConfiguration(),
	})
	log.Println(config.Validate())

	vm := vz.NewVirtualMachine(config)
	_ = vm
	go func(vm *vz.VirtualMachine) {
		t := time.NewTicker(time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
			case newState := <-vm.StateChangedNotify():
				log.Println(
					"newState:", newState,
					"state:", vm.State(),
					"canStart:", vm.CanStart(),
					"canResume:", vm.CanResume(),
					"canPause:", vm.CanPause(),
					"canStopRequest:", vm.CanRequestStop(),
				)
			}
		}
	}(vm)

	vm.Start(func(err error) {
		log.Println("in start:", err)
	})

	<-time.After(3 * time.Minute)

	// vm.Resume(func(err error) {
	// 	fmt.Println("in resume:", err)
	// })
}
