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

package vf

import (
	"errors"
	"fmt"
	"os/exec"
	"time"

	"github.com/Code-Hex/vz"
	crcos "github.com/code-ready/crc/pkg/os"
	vfdriver "github.com/code-ready/machine/drivers/vf"
	"github.com/code-ready/machine/libmachine/drivers"
	"github.com/code-ready/machine/libmachine/state"
	log "github.com/sirupsen/logrus"
)

type Driver struct {
	vfdriver.Driver
	vzVirtualMachine *vz.VirtualMachine
}

func NewDriver() *Driver {
	// checks that vfdriver.Driver implements the libmachine.Driver interface
	var _ drivers.Driver = &Driver{}
	return &Driver{
		Driver: vfdriver.Driver{
			VMDriver: &drivers.VMDriver{
				BaseDriver: &drivers.BaseDriver{},
				CPU:        DefaultCPUs,
				Memory:     DefaultMemory,
			},
		},
	}
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return DriverName
}

// Get Version information
func (d *Driver) DriverVersion() string {
	return DriverVersion
}

// GetIP returns an IP or hostname that this host is available at
// inherited from  libmachine.BaseDriver
//func (d *Driver) GetIP() (string, error)

// GetMachineName returns the name of the machine
// inherited from  libmachine.BaseDriver
//func (d *Driver) GetMachineName() string

// GetBundleName() Returns the name of the unpacked bundle which was used to create this machine
// inherited from  libmachine.BaseDriver
//func (d *Driver) GetBundleName() (string, error)

// PreCreateCheck allows for pre-create operations to make sure a driver is ready for creation
func (d *Driver) PreCreateCheck() error {
	return nil
}

func (d *Driver) getDiskPath() string {
	return d.ResolveStorePath(fmt.Sprintf("%s.img", d.MachineName))
}

// Create a host using the driver's config
func (d *Driver) Create() error {
	if err := d.PreCreateCheck(); err != nil {
		return err
	}

	// use qemu-img for now for the conversion, but we need to remove this dependency
	qemuImgPath, err := exec.LookPath("qemu-img")
	if err != nil {
		log.Println("Could not find the qemu-img execurable in $PATH, please install it using 'brew install qemu'")
		return err
	}

	log.Println("Converting disk image")
	stdout, stderr, err := crcos.RunWithDefaultLocale(qemuImgPath, "convert", "-f", d.ImageFormat, "-O", "raw", d.ImageSourcePath, d.getDiskPath())
	if err != nil {
		log.Println("RunWithDefaultLocale error: %s %s\n", stdout, stderr)
		return err
	}

	// TODO: resize disk
	return nil
}

// Start a host
func (d *Driver) Start() error {
	bootLoader := vz.NewLinuxBootLoader(
		d.VmlinuzPath,
		vz.WithCommandLine("console=hvc0 "+d.Cmdline),
		vz.WithInitrd(d.InitrdPath),
	)
	log.Println("bootloader:", bootLoader)

	config := vz.NewVirtualMachineConfiguration(
		bootLoader,
		uint(d.CPU),
		uint64(d.Memory*1024*1024),
	)

	// console
	logFile := d.ResolveStorePath(fmt.Sprintf("%s.log", d.MachineName))
	serialPortAttachment, err := vz.NewFileSerialPortAttachment(logFile, false)
	if err != nil {
		return err
	}
	//serialPortAttachment := vz.NewFileHandleSerialPortAttachment(os.Stdin, tty)
	consoleConfig := vz.NewVirtioConsoleDeviceSerialPortConfiguration(serialPortAttachment)
	config.SetSerialPortsVirtualMachineConfiguration([]*vz.VirtioConsoleDeviceSerialPortConfiguration{
		consoleConfig,
	})

	log.Println("d.VMNet: ", d.VMNet)
	// network

	var mac *vz.MACAddress
	if d.VMNet {
		natAttachment := vz.NewNATNetworkDeviceAttachment()
		networkConfig := vz.NewVirtioNetworkDeviceConfiguration(natAttachment)
		mac = vz.NewRandomLocallyAdministeredMACAddress()
		networkConfig.SetMacAddress(mac)
		config.SetNetworkDevicesVirtualMachineConfiguration([]*vz.VirtioNetworkDeviceConfiguration{
			networkConfig,
		})
	}

	// entropy
	entropyConfig := vz.NewVirtioEntropyDeviceConfiguration()
	config.SetEntropyDevicesVirtualMachineConfiguration([]*vz.VirtioEntropyDeviceConfiguration{
		entropyConfig,
	})

	// disk
	diskPath := d.getDiskPath()

	diskImageAttachment, err := vz.NewDiskImageStorageDeviceAttachment(
		diskPath,
		false,
	)
	if err != nil {
		return err
	}
	storageDeviceConfig := vz.NewVirtioBlockDeviceConfiguration(diskImageAttachment)
	config.SetStorageDevicesVirtualMachineConfiguration([]vz.StorageDeviceConfiguration{
		storageDeviceConfig,
	})

	// virtio-vsock device
	config.SetSocketDevicesVirtualMachineConfiguration([]vz.SocketDeviceConfiguration{
		vz.NewVirtioSocketDeviceConfiguration(),
	})

	valid, err := config.Validate()
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("Invalid virtual machine configuration")
	}

	vm := vz.NewVirtualMachine(config)
	d.vzVirtualMachine = vm
	/*
		go func(vm *vz.VirtualMachine) {
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()
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
	*/

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
	if err := d.exposeVsock(); err != nil {
		log.Warnf("Error listening on vsock: %v", err)
	}

	if !d.VMNet {
		return nil
	}

	getIP := func() error {
		d.IPAddress, err = GetIPAddressByMACAddress(mac.String())
		if err != nil {
			return &RetriableError{Err: err}
		}
		return nil
	}

	if err := RetryAfter(60, getIP, 2*time.Second); err != nil {
		return fmt.Errorf("IP address never found in dhcp leases file %v", err)
	}
	log.Debugf("IP: %s", d.IPAddress)

	return nil
}

func vzStateToState(vzState vz.VirtualMachineState) state.State {
	switch vzState {
	case vz.VirtualMachineStateStopped:
		return state.Stopped

	case vz.VirtualMachineStateRunning:
		return state.Running

	case vz.VirtualMachineStateStarting:
		// not sure what the proper state is
		return state.Stopped

	case vz.VirtualMachineStatePaused:
	case vz.VirtualMachineStateError:
	case vz.VirtualMachineStatePausing:
	case vz.VirtualMachineStateResuming:
		return state.Error
	default:
		log.Warnf("Unhandled stated: %v", vzState)
	}
	return state.Error
}

// GetState returns the state that the host is in (running, stopped, etc)
func (d *Driver) GetState() (state.State, error) {
	if d.vzVirtualMachine == nil {
		return state.Stopped, nil
	}
	return vzStateToState(d.vzVirtualMachine.State()), nil
}

// Kill stops a host forcefully
func (d *Driver) Kill() error {
	return errors.New("Kill() is not implemented")
}

// Remove a host
func (d *Driver) Remove() error {
	return errors.New("Remove() is not implemented")
}

// UpdateConfigRaw allows to change the state (memory, ...) of an already created machine
func (d *Driver) UpdateConfigRaw(rawDriver []byte) error {
	return errors.New("UpdateConfigRaw() is not implemented")
}

// Stop a host gracefully
func (d *Driver) Stop() error {
	st, err := d.GetState()
	if err != nil {
		return err
	}
	if st == state.Stopped {
		return nil
	}
	stopped, err := d.vzVirtualMachine.RequestStop()
	if err != nil {
		log.Warnf("Failed to stop VM")
		return err
	}
	st, _ = d.GetState()
	log.Warnf("Stop(): stopped: %v current state: %v", stopped, st)

	/*
		if !stopped {
	*/
	for i := 0; i < 120; i++ {
		st, _ := d.GetState()
		log.Debugf("VM state: %s", st)
		if st == state.Stopped {
			return nil
		}
		time.Sleep(time.Second)
	}
	return errors.New("VM Failed to gracefully shutdown, try the kill command")
	/*
		}
		return nil
	*/
}
