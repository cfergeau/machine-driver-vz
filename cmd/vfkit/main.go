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
	"time"

	"github.com/Code-Hex/vz"
	"github.com/code-ready/machine-driver-vf/pkg/config"
	"github.com/code-ready/machine-driver-vf/pkg/vf"
	"github.com/docker/go-units"
	log "github.com/sirupsen/logrus"
)

type cmdlineOptions struct {
	vcpus     uint
	memoryMiB uint

	vmlinuzPath   string
	kernelCmdline string
	initrdPath    string

	devices []string
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

	if err := config.AddDevicesFromCmdLine(opts.devices, vmConfig); err != nil {
		return nil, err
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

func newVirtualMachine(opts *cmdlineOptions) (*vz.VirtualMachine, error) {
	vmConfig, err := createVMConfiguration(opts)
	if err != nil {
		return nil, err
	}

	vm := vz.NewVirtualMachine(vmConfig)
	return vm, nil
}

func waitForVMState(vm *vz.VirtualMachine, state vz.VirtualMachineState) error {
	for {
		select {
		case newState := <-vm.StateChangedNotify():
			if newState == state {
				return nil
			}
		case <-time.After(5 * time.Second):
			return fmt.Errorf("timeout waiting for VM state %v", state)
		}
	}
}

func runVirtualMachine(vm *vz.VirtualMachine) error {
	errCh := make(chan error, 1)
	vm.Start(func(err error) {
		if err != nil {
			errCh <- err
		}
		errCh <- waitForVMState(vm, vz.VirtualMachineStateRunning)
	})

	err := <-errCh
	if err != nil {
		return err
	}
	log.Infof("virtual machine is running")
	if err := vf.ExposeVsock(vm, opts.vsockSocketPath); err != nil {
		log.Warnf("error listening on vsock: %v", err)
	}
	log.Infof("waiting for VM to stop")
	for {
		err := waitForVMState(vm, vz.VirtualMachineStateStopped)
		if err == nil {
			log.Infof("VM is stopped")
			return nil
		}
	}

}
