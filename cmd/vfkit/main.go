// +build darwin

/*
Copyright 2021, Red Hat, Inc - All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"net"

	"github.com/Code-Hex/vz"
	"github.com/docker/go-units"
	log "github.com/sirupsen/logrus"
)

type cmdlineOptions struct {
	vcpus     uint
	memoryMiB uint

	vmlinuzPath   string
	kernelCmdline string
	initrdPath    string

	natNetworking bool
	macAddress    string
	logFilePath   string

	rngDevice bool

	diskPath string

	vsockSocketPath string
}

type virtualMachine *vz.VirtualMachine

func addLogFile(vmConfig *vz.VirtualMachineConfiguration, logFile string) error {
	//serialPortAttachment := vz.NewFileHandleSerialPortAttachment(os.Stdin, tty)
	serialPortAttachment, err := vz.NewFileSerialPortAttachment(logFile, false)
	if err != nil {
		return err
	}
	consoleConfig := vz.NewVirtioConsoleDeviceSerialPortConfiguration(serialPortAttachment)
	vmConfig.SetSerialPortsVirtualMachineConfiguration([]*vz.VirtioConsoleDeviceSerialPortConfiguration{
		consoleConfig,
	})

	return nil
}

func addNetworkNAT(vmConfig *vz.VirtualMachineConfiguration, macAddress string) error {
	var mac *vz.MACAddress
	if macAddress == "" {
		mac = vz.NewRandomLocallyAdministeredMACAddress()
	} else {
		hwAddr, err := net.ParseMAC(macAddress)
		if err != nil {
			return err
		}
		mac = vz.NewMACAddress(hwAddr)
	}
	natAttachment := vz.NewNATNetworkDeviceAttachment()
	networkConfig := vz.NewVirtioNetworkDeviceConfiguration(natAttachment)
	networkConfig.SetMacAddress(mac)
	vmConfig.SetNetworkDevicesVirtualMachineConfiguration([]*vz.VirtioNetworkDeviceConfiguration{
		networkConfig,
	})

	return nil
}

func addEntropy(vmConfig *vz.VirtualMachineConfiguration) error {
	entropyConfig := vz.NewVirtioEntropyDeviceConfiguration()
	vmConfig.SetEntropyDevicesVirtualMachineConfiguration([]*vz.VirtioEntropyDeviceConfiguration{
		entropyConfig,
	})

	return nil
}

func addDisk(vmConfig *vz.VirtualMachineConfiguration, diskPath string) error {
	diskImageAttachment, err := vz.NewDiskImageStorageDeviceAttachment(
		diskPath,
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

func addVirtioVsock(vmConfig *vz.VirtualMachineConfiguration) error {
	vmConfig.SetSocketDevicesVirtualMachineConfiguration([]vz.SocketDeviceConfiguration{
		vz.NewVirtioSocketDeviceConfiguration(),
	})

	return nil
}

func createVMConfiguration(opts *cmdlineOptions) (*vz.VirtualMachineConfiguration, error) {
	log.Info(opts)
	bootLoader := vz.NewLinuxBootLoader(
		opts.vmlinuzPath,
		vz.WithCommandLine(opts.kernelCmdline),
		vz.WithInitrd(opts.initrdPath),
	)
	log.Info("boot parameters:")
	log.Infof("\tkernel: %s", opts.vmlinuzPath)
	log.Infof("\tkernel command line:%s", opts.kernelCmdline)
	log.Infof("\tinitrd: %s", opts.initrdPath)
	log.Info()

	vmConfig := vz.NewVirtualMachineConfiguration(
		bootLoader,
		opts.vcpus,
		uint64(opts.memoryMiB*units.MiB),
	)
	log.Info("virtual machine parameters:")
	log.Infof("\tvCPUs: %d", opts.vcpus)
	log.Infof("\tmemory: %d MiB", opts.memoryMiB)
	log.Info()

	if opts.logFilePath != "" {
		log.Infof("sending VM output to %s", opts.logFilePath)
		if err := addLogFile(vmConfig, opts.logFilePath); err != nil {
			return nil, err
		}
	}

	if opts.natNetworking {
		log.Infof("adding virtio-net device (mac: %s)", opts.macAddress)
		if err := addNetworkNAT(vmConfig, opts.macAddress); err != nil {
			return nil, err
		}
	}

	if opts.rngDevice {
		log.Infof("adding virtio-rng device")
		if err := addEntropy(vmConfig); err != nil {
			return nil, err
		}
	}

	if opts.diskPath != "" {
		log.Infof("adding disk image %s", opts.diskPath)
		if err := addDisk(vmConfig, opts.diskPath); err != nil {
			return nil, err
		}
	}

	if opts.vsockSocketPath != "" {
		log.Infof("adding virtio-vsock device at %s", opts.vsockSocketPath)
		if err := addVirtioVsock(vmConfig); err != nil {
			return nil, err
		}
	}

	valid, err := vmConfig.Validate()
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, fmt.Errorf("Invalid virtual machine configuration")
	}

	return vmConfig, nil
}

func newVirtualMachine(opts *cmdlineOptions) (virtualMachine, error) {
	vmConfig, err := createVMConfiguration(opts)
	if err != nil {
		return nil, err
	}

	vm := vz.NewVirtualMachine(vmConfig)
	return vm, nil
	/*
		d.vzVirtualMachine = vm

		errCh := make(chan error, 1)
		vm.Start(func(err error) {
			log.Println("in start:", err)
			if err != nil {
				errCh <- err
			}
		loop:
			for {
				select {
				case newState := <-vm.StateChangedNotify():
					if newState == vz.VirtualMachineStateRunning {
						errCh <- nil
						break loop
					}
				case <-time.After(5 * time.Second):
					errCh <- errors.New("virtual machine failed to start")
					break loop
				}
			}
		})

		err = <-errCh
		if err != nil {
			return err
		}
		if err := exposeVsock(vm, opts.vsockSocketPath); err != nil {
			log.Warnf("Error listening on vsock: %v", err)
		}
	*/
}
