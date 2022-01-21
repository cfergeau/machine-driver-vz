package config

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/Code-Hex/vz"
	log "github.com/sirupsen/logrus"
)

type VirtioDevice interface {
	FromOptions([]option) error
	AddToVirtualMachineConfig(*vz.VirtualMachineConfiguration) error
}

type virtioVsock struct {
	port      uint
	socketURL string
}

type virtioBlk struct {
	imagePath string
}

type virtioRng struct {
}

type virtioNet struct {
	nat        bool
	macAddress net.HardwareAddr
}

type virtioSerial struct {
	logFile string
}

type option struct {
	key   string
	value string
}

func strToOption(str string) option {
	splitStr := strings.SplitN(str, "=", 2)

	opt := option{
		key: splitStr[0],
	}
	if len(splitStr) > 1 {
		opt.value = splitStr[1]
	}

	return opt
}

func DevicesFromCmdLine(cmdlineOpts []string) ([]VirtioDevice, error) {
	devs := []VirtioDevice{}
	for _, deviceOpts := range cmdlineOpts {
		dev, err := deviceFromCmdLine(deviceOpts)
		if err != nil {
			return nil, err
		}
		devs = append(devs, dev)
	}
	return devs, nil
}

func AddDevicesFromCmdLine(cmdlineOpts []string, vmConfig *vz.VirtualMachineConfiguration) error {
	for _, deviceOpts := range cmdlineOpts {
		if err := addDeviceFromCmdLine(deviceOpts, vmConfig); err != nil {
			return err
		}
	}
	return nil
}

func deviceFromCmdLine(deviceOpts string) (VirtioDevice, error) {
	opts := strings.Split(deviceOpts, ",")
	if len(opts) == 0 {
		return nil, fmt.Errorf("empty option list in command line argument")
	}
	var dev VirtioDevice
	switch opts[0] {
	case "virtio-blk":
		dev = &virtioBlk{}
	case "virtio-net":
		dev = &virtioNet{}
	case "virtio-rng":
		dev = &virtioRng{}
	case "virtio-serial":
		dev = &virtioSerial{}
	case "virtio-vsock":
		dev = &virtioVsock{}
	default:
		return nil, fmt.Errorf("unknown device type: %s", opts[0])
	}
	parsedOpts := []option{}
	for _, opt := range opts[1:] {
		if len(opt) == 0 {
			continue
		}
		parsedOpts = append(parsedOpts, strToOption(opt))
	}

	if err := dev.FromOptions(parsedOpts); err != nil {
		return nil, err
	}

	return dev, nil
}

func addDeviceFromCmdLine(deviceOpts string, vmConfig *vz.VirtualMachineConfiguration) error {
	dev, err := deviceFromCmdLine(deviceOpts)
	if err != nil {
		return err
	}
	return dev.AddToVirtualMachineConfig(vmConfig)
}

func (dev *virtioSerial) FromOptions(options []option) error {
	for _, option := range options {
		switch option.key {
		case "logFilePath":
			dev.logFile = option.value
		default:
			return fmt.Errorf("Unknown option for virtio-serial devices: %s", option.key)
		}
	}
	return nil
}

func (dev *virtioSerial) AddToVirtualMachineConfig(vmConfig *vz.VirtualMachineConfiguration) error {
	if dev.logFile == "" {
		return fmt.Errorf("missing mandatory 'logFile' option for virtio-serial device")
	}
	log.Infof("Adding virtio-serial device (logFile: %s)", dev.logFile)

	//serialPortAttachment := vz.NewFileHandleSerialPortAttachment(os.Stdin, tty)
	serialPortAttachment, err := vz.NewFileSerialPortAttachment(dev.logFile, false)
	if err != nil {
		return err
	}
	consoleConfig := vz.NewVirtioConsoleDeviceSerialPortConfiguration(serialPortAttachment)
	vmConfig.SetSerialPortsVirtualMachineConfiguration([]*vz.VirtioConsoleDeviceSerialPortConfiguration{
		consoleConfig,
	})

	return nil
}

func (dev *virtioNet) FromOptions(options []option) error {
	for _, option := range options {
		switch option.key {
		case "nat":
			if option.value != "" {
				return fmt.Errorf("Unexpected value for virtio-net 'nat' option: %s", option.value)
			}
			dev.nat = true
		case "mac":
			macAddress, err := net.ParseMAC(option.value)
			if err != nil {
				return err
			}
			dev.macAddress = macAddress
		default:
			return fmt.Errorf("Unknown option for virtio-net devices: %s", option.key)
		}
	}
	return nil
}

func (dev *virtioNet) AddToVirtualMachineConfig(vmConfig *vz.VirtualMachineConfiguration) error {
	var mac *vz.MACAddress

	if !dev.nat {
		return fmt.Errorf("NAT is the only supported networking mode")
	}

	log.Infof("Adding virtio-net device (nat: %t macAddress: [%s])", dev.nat, dev.macAddress)

	if len(dev.macAddress) == 0 {
		mac = vz.NewRandomLocallyAdministeredMACAddress()
	} else {
		mac = vz.NewMACAddress(dev.macAddress)
	}
	natAttachment := vz.NewNATNetworkDeviceAttachment()
	networkConfig := vz.NewVirtioNetworkDeviceConfiguration(natAttachment)
	networkConfig.SetMacAddress(mac)
	vmConfig.SetNetworkDevicesVirtualMachineConfiguration([]*vz.VirtioNetworkDeviceConfiguration{
		networkConfig,
	})

	return nil
}

func (dev *virtioRng) FromOptions(options []option) error {
	if len(options) != 0 {
		return fmt.Errorf("Unknown options for virtio-rng devices: %s", options)
	}
	return nil
}

func (dev *virtioRng) AddToVirtualMachineConfig(vmConfig *vz.VirtualMachineConfiguration) error {
	log.Infof("Adding virtio-rng device")
	entropyConfig := vz.NewVirtioEntropyDeviceConfiguration()
	vmConfig.SetEntropyDevicesVirtualMachineConfiguration([]*vz.VirtioEntropyDeviceConfiguration{
		entropyConfig,
	})

	return nil
}

func (dev *virtioBlk) FromOptions(options []option) error {
	for _, option := range options {
		switch option.key {
		case "path":
			dev.imagePath = option.value
		default:
			return fmt.Errorf("Unknown option for virtio-blk devices: %s", option.key)
		}
	}
	return nil
}

func (dev *virtioBlk) AddToVirtualMachineConfig(vmConfig *vz.VirtualMachineConfiguration) error {
	if dev.imagePath == "" {
		return fmt.Errorf("missing mandatory 'path' option for virtio-blk device")
	}
	log.Infof("Adding virtio-blk device (imagePath: %s)", dev.imagePath)
	diskImageAttachment, err := vz.NewDiskImageStorageDeviceAttachment(
		dev.imagePath,
		false,
	)
	if err != nil {
		return err
	}
	storageDeviceConfig := vz.NewVirtioBlockDeviceConfiguration(diskImageAttachment)
	vmConfig.SetStorageDevicesVirtualMachineConfiguration([]vz.StorageDeviceConfiguration{
		storageDeviceConfig,
	})
	return nil
}

func (dev *virtioVsock) FromOptions(options []option) error {
	for _, option := range options {
		switch option.key {
		case "socketURL":
			dev.socketURL = option.value
		case "port":
			port, err := strconv.Atoi(option.value)
			if err != nil {
				return err
			}
			dev.port = uint(port)
		default:
			return fmt.Errorf("Unknown option for virtio-vsock devices: %s", option.key)
		}
	}
	return nil
}

func (dev *virtioVsock) AddToVirtualMachineConfig(vmConfig *vz.VirtualMachineConfiguration) error {
	log.Infof("Adding virtio-vsock device")
	vmConfig.SetSocketDevicesVirtualMachineConfiguration([]vz.SocketDeviceConfiguration{
		vz.NewVirtioSocketDeviceConfiguration(),
	})

	return nil
}
