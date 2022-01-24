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
	"os"
	"os/exec"
	"strings"
	"time"

	crcos "github.com/code-ready/crc/pkg/os"
	"github.com/code-ready/machine-driver-vf/pkg/client"
	vfdriver "github.com/code-ready/machine/drivers/vf"
	"github.com/code-ready/machine/libmachine/drivers"
	"github.com/code-ready/machine/libmachine/state"
	log "github.com/sirupsen/logrus"
)

type Driver struct {
	vfdriver.Driver
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

func convertToRaw(source, sourceFormat string, dest string) error {
	// use qemu-img for now for the conversion, but we need to remove this dependency
	qemuImgPath, err := exec.LookPath("qemu-img")
	if err != nil {
		log.Println("Could not find the qemu-img execurable in $PATH, please install it using 'brew install qemu'")
		return err
	}

	log.Println("Converting disk image")
	stdout, stderr, err := crcos.RunWithDefaultLocale(qemuImgPath, "convert", "-f", sourceFormat, "-O", "raw", source, dest)
	if err != nil {
		log.Println("RunWithDefaultLocale error: %s %s\n", stdout, stderr)
		return err
	}

	return nil
}

// Create a host using the driver's config
func (d *Driver) Create() error {
	if err := d.PreCreateCheck(); err != nil {
		return err
	}

	switch d.ImageFormat {
	case "raw":
		break
	case "qcow2", "vmdk", "vhdx":
		if err := convertToRaw(d.ImageSourcePath, d.ImageFormat, d.getDiskPath()); err != nil {
			return err
		}
	}

	// TODO: resize disk
	return nil
}

func startVfkit(args []string) error {
	vfkitPath, err := exec.LookPath("vfkit")
	if err != nil {
		return err
	}
	cmd := exec.Command(vfkitPath, args...)
	cmd.Env = os.Environ()
	if err := cmd.Start(); err != nil {
		return err
	}
	// cmd.Process.Pid
	// cmd.Process
	// cmd.Process.Kill()

	return nil
}

// Start a host
func (d *Driver) Start() error {
	bootLoader := client.NewBootloader(
		d.VmlinuzPath,
		"console=hvc0 "+d.Cmdline,
		d.InitrdPath,
	)
	log.Println("bootloader:", bootLoader)

	vm := client.NewVirtualMachine(
		uint(d.CPU),
		uint64(d.Memory*1024*1024),
		bootLoader,
	)

	// console
	logFile := d.ResolveStorePath(fmt.Sprintf("%s.log", d.MachineName))
	dev, err := client.VirtioSerialNew(logFile)
	if err != nil {
		return err
	}
	err = vm.AddDevice(dev)
	if err != nil {
		return err
	}

	// network
	log.Println("d.VMNet: ", d.VMNet)
	// 52:54:00 is the OUI used by QEMU
	const mac = "52:54:00:70:2b:79"
	if d.VMNet {
		dev, err = client.VirtioNetNew("")
		if err != nil {
			return err
		}
		err = vm.AddDevice(dev)
		if err != nil {
			return err
		}
	}

	// entropy
	dev, err = client.VirtioRNGNew()
	if err != nil {
		return err
	}
	err = vm.AddDevice(dev)
	if err != nil {
		return err
	}

	// disk
	diskPath := d.getDiskPath()
	dev, err = client.VirtioBlkNew(diskPath)
	if err != nil {
		return err
	}
	err = vm.AddDevice(dev)
	if err != nil {
		return err
	}

	// virtio-vsock device
	const vsockPort = 1024
	dev, err = client.VirtioVsockNew(1024, d.VsockPath)
	err = vm.AddDevice(dev)
	if err != nil {
		return err
	}

	if !d.VMNet {
		return nil
	}

	args, err := vm.ToCmdLine()
	if err != nil {
		return err
	}
	log.Infof("commandline: %s", strings.Join(args, " "))
	if err := startVfkit(args); err != nil {
		return err
	}
	//return fmt.Errorf("starting the VM is not implemented yet!!")
	getIP := func() error {
		d.IPAddress, err = GetIPAddressByMACAddress(mac)
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

// GetState returns the state that the host is in (running, stopped, etc)
func (d *Driver) GetState() (state.State, error) {
	/*
		if d.vzVirtualMachine == nil {
			return state.Stopped, nil
		}
		return vzStateToState(d.vzVirtualMachine.State()), nil
	*/
	return state.Stopped, nil
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
	return fmt.Errorf("Stop() is not fully implemented")
	/*
		stopped, err := d.vzVirtualMachine.RequestStop()
		if err != nil {
			log.Warnf("Failed to stop VM")
			return err
		}
		st, _ = d.GetState()
		log.Warnf("Stop(): stopped: %v current state: %v", stopped, st)
	*/

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
