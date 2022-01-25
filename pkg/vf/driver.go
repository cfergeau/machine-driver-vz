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
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	crcos "github.com/code-ready/crc/pkg/os"
	"github.com/code-ready/machine-driver-vf/pkg/client"
	vfdriver "github.com/code-ready/machine/drivers/vf"
	"github.com/code-ready/machine/libmachine/drivers"
	"github.com/code-ready/machine/libmachine/state"
	"github.com/mitchellh/go-ps"
	"github.com/pkg/errors"
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

func startVfkit(args []string) (*os.Process, error) {
	vfkitPath, err := exec.LookPath("vfkit")
	if err != nil {
		return nil, err
	}
	/*
		// for debug logs of vfkit startup
		logFile, err := os.Create("/tmp/vfkit.log")
		if err != nil {
			return nil, err
		}
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	*/

	cmd := exec.Command(vfkitPath, args...)
	cmd.Env = os.Environ()
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	// cmd.Process.Pid
	// cmd.Process
	// cmd.Process.Kill()

	errCh := make(chan error, 1)
	go func() {
		errCh <- cmd.Wait()
	}()

	// catch vfkit early startup failures
	select {
	case err := <-errCh:
		return nil, err
	case <-time.After(time.Second):
		break
	}

	return cmd.Process, nil
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
		uint64(d.Memory),
		bootLoader,
	)

	// console
	//logFile := d.ResolveStorePath(fmt.Sprintf("%s.log", d.MachineName))
	logFile := d.ResolveStorePath("vfkit.log")
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
		dev, err = client.VirtioNetNew(mac)
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

	args, err := vm.ToCmdLine()
	if err != nil {
		return err
	}
	log.Infof("commandline: %s", strings.Join(args, " "))
	process, err := startVfkit(args)
	if err != nil {
		return err
	}

	_ = os.WriteFile(d.getPidFilePath(), []byte(strconv.Itoa(process.Pid)), 0600)

	if !d.VMNet {
		return nil
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
	p, err := d.findVfkitProcess()
	if err != nil {
		return state.Error, err
	}
	if p == nil {
		return state.Stopped, nil
	}
	return state.Running, nil
	/*
			if d.vzVirtualMachine == nil {
				return state.Stopped, nil
			}
			return vzStateToState(d.vzVirtualMachine.State()), nil
		return state.Stopped, nil
	*/
}

// Kill stops a host forcefully
func (d *Driver) Kill() error {
	return d.sendSignal(syscall.SIGKILL)
}

// Remove a host
func (d *Driver) Remove() error {
	s, err := d.GetState()
	if err != nil || s == state.Error {
		log.Debugf("Error checking machine status: %v, assuming it has been removed already", err)
	}
	if s == state.Running {
		if err := d.Kill(); err != nil {
			return err
		}
	}
	return nil
}

// UpdateConfigRaw allows to change the state (memory, ...) of an already created machine
func (d *Driver) UpdateConfigRaw(rawDriver []byte) error {
	return errors.New("UpdateConfigRaw() is not implemented")
}

// Stop a host gracefully
func (d *Driver) Stop() error {
	s, err := d.GetState()
	if err != nil {
		return err
	}

	if s != state.Stopped {
		err := d.sendSignal(syscall.SIGTERM)
		if err != nil {
			return errors.Wrap(err, "hyperkit sigterm failed")
		}
		// wait 120s for graceful shutdown
		for i := 0; i < 60; i++ {
			time.Sleep(2 * time.Second)
			s, _ := d.GetState()
			log.Debugf("VM state: %s", s)
			if s == state.Stopped {
				return nil
			}
		}
		return errors.New("VM Failed to gracefully shutdown, try the kill command")
	}
	return nil
}

func (d *Driver) getPidFilePath() string {
	const pidFileName = "vfkit.pid"
	return d.ResolveStorePath(pidFileName)
}

/*
 * Returns a ps.Process instance if it could find a vfkit process with the pid
 * stored in $pidFileName
 *
 * Returns nil, nil if:
 * - if the $pidFileName file does not exist,
 * - if a process with the pid from this file cannot be found,
 * - if a process was found, but its name is not 'vfkit'
 */
func (d *Driver) findVfkitProcess() (ps.Process, error) {
	pidFile := d.getPidFilePath()
	pid, err := readPidFromFile(pidFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "error reading pidfile %s", pidFile)
	}

	p, err := ps.FindProcess(pid)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("cannot find pid %d", pid))
	}
	if p == nil {
		log.Debugf("vfkit pid %d missing from process table", pid)
		// return PidNotExist error?
		return nil, nil
	}

	// match both hyperkit and com.docker.hyper
	if p.Executable() != "vfkit" {
		// return InvalidExecutable error?
		log.Debugf("pid %d is stale, and is being used by %s", pid, p.Executable())
		return nil, nil
	}

	return p, nil
}

func readPidFromFile(filename string) (int, error) {
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		return 0, err
	}
	content := strings.TrimSpace(string(bs))
	pid, err := strconv.Atoi(content)
	if err != nil {
		return 0, errors.Wrapf(err, "parsing %s", filename)
	}

	return pid, nil
}

// recoverFromUncleanShutdown searches for an existing vfkit.pid file in
// the machine directory. If it can't find it, a clean shutdown is assumed.
// If it finds the pid file, it checks for a running vfkit process with that pid
// as the existence of a file might not indicate an unclean shutdown but an actual running
// vfkit server. This is an error situation - we shouldn't start minikube as there is likely
// an instance running already. If the PID in the pidfile does not belong to a running vfkit
// process, we can safely delete it, and there is a good chance the machine will recover when restarted.
func (d *Driver) recoverFromUncleanShutdown() error {
	proc, err := d.findVfkitProcess()
	if err == nil && proc != nil {
		/* hyperkit is running, pid file can't be stale */
		return nil
	}
	pidFile := d.getPidFilePath()
	/* There might be a stale pid file, try to remove it */
	if err := os.Remove(pidFile); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return errors.Wrap(err, fmt.Sprintf("removing pidFile %s", pidFile))
		}
	} else {
		log.Debugf("Removed stale pid file %s...", pidFile)
	}
	return nil
}

func (d *Driver) sendSignal(s os.Signal) error {
	psProc, err := d.findVfkitProcess()
	if err != nil {
		return err
	}
	proc, err := os.FindProcess(psProc.Pid())
	if err != nil {
		return err
	}

	return proc.Signal(s)
}
