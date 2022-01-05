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
	"github.com/Code-Hex/vz"
	vfdriver "github.com/code-ready/machine/drivers/vf"
	"github.com/code-ready/machine/libmachine/drivers"
	"github.com/code-ready/machine/libmachine/state"
)

type Driver vfdriver.Driver

func NewDriver() *Driver {
	// checks that vfdriver.Driver implements the libmachine.Driver interface
	var _ drivers.Driver = &Driver{}
	return &Driver{
		VMDriver: &drivers.VMDriver{
			BaseDriver: &drivers.BaseDriver{},
			CPU:        DefaultCPUs,
			Memory:     DefaultMemory,
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

// Create a host using the driver's config
func (d *Driver) Create() error {
	bootLoader := vz.NewLinuxBootLoader(
		d.VmlinuzPath,
		vz.WithCommandLine(d.KernelCmdLine),
		vz.WithInitrd(d.InitrdPath),
	)

	config := vz.NewVirtualMachineConfiguration(
		bootLoader,
		uint(d.CPU),
		uint64(d.Memory),
	)

	// add console for serial output

	natAttachment := vz.NewNATNetworkDeviceAttachment()
	networkConfig := vz.NewVirtioNetworkDeviceConfiguration(natAttachment)
	config.SetNetworkDevicesVirtualMachineConfiguration([]*vz.VirtioNetworkDeviceConfiguration{
		networkConfig,
	})

	entropyConfig := vz.NewVirtioEntropyDeviceConfiguration()
	config.SetEntropyDevicesVirtualMachineConfiguration([]*vz.VirtioEntropyDeviceConfiguration{
		entropyConfig,
	})

	// add disk
}

// GetState returns the state that the host is in (running, stopped, etc)
func (d *Driver) GetState() (state.State, error) {
	return state.Error, nil
}

// Kill stops a host forcefully
func (d *Driver) Kill() error {
	return nil
}

// Remove a host
func (d *Driver) Remove() error {
	return nil
}

// UpdateConfigRaw allows to change the state (memory, ...) of an already created machine
func (d *Driver) UpdateConfigRaw(rawDriver []byte) error {
	return nil
}

// Start a host
func (d *Driver) Start() error {
	return nil
}

// Stop a host gracefully
func (d *Driver) Stop() error {
	return nil
}
