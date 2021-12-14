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

package vz

import (
	vzdriver "github.com/code-ready/machine/drivers/vz"
	"github.com/code-ready/machine/libmachine/state"
	"github.com/code-ready/machine/libmachine/drivers"
)

type Driver vzdriver.Driver

func NewDriver() *Driver {
	return &Driver{
		VMDriver: &drivers.VMDriver{
			BaseDriver: &drivers.BaseDriver{},
			CPU:        DefaultCPUs,
			Memory:     DefaultMemory,
		},
	}
}

// Create a host using the driver's config
func (d *Driver) Create() error {
	return nil
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return ""
}

// GetIP returns an IP or hostname that this host is available at
// e.g. 1.2.3.4 or docker-host-d60b70a14d3a.cloudapp.net
func (d *Driver) GetIP() (string, error) {
	return "", nil
}

// GetMachineName returns the name of the machine
func (d *Driver) GetMachineName() string {
	return ""
}

// GetBundleName() Returns the name of the unpacked bundle which was used to create this machine
func (d *Driver) GetBundleName() (string, error) {
	return "", nil
}

// GetState returns the state that the host is in (running, stopped, etc)
func (d *Driver) GetState() (state.State, error) {
	return state.Error, nil
}

// Kill stops a host forcefully
func (d *Driver) Kill() error {
	return nil
}

// PreCreateCheck allows for pre-create operations to make sure a driver is ready for creation
func (d *Driver) PreCreateCheck() error {
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

// Get Version information
func (d *Driver) DriverVersion() string {
	return ""
}
