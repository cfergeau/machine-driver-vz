package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const vfkitVersion = "0.0.1"

var opts = &cmdlineOptions{}

var rootCmd = &cobra.Command{
	Use:   "vfkit",
	Short: "vfkit is a simple hypervisor using Apple's virtualization framework",
	Long: `A hypervisor written in Go using Apple's virtualization framework to run linux virtual machines.
                Complete documentation is available at https://github.com/code-ready/vfkit`,
	RunE: func(cmd *cobra.Command, args []string) error {
		vm, err := newVirtualMachine(opts)
		if err != nil {
			return err
		}
		return runVirtualMachine(vm)
	},
	Version: vfkitVersion,
}

func init() {
	rootCmd.Flags().StringVarP(&opts.vmlinuzPath, "kernel", "k", "", "path to the virtual machine linux kernel")
	rootCmd.Flags().StringVarP(&opts.kernelCmdline, "kernel-cmdline", "C", "", "linux kernel command line")
	rootCmd.Flags().StringVarP(&opts.initrdPath, "initrd", "i", "", "path to the virtual machine initrd")
	rootCmd.MarkFlagRequired("kernel")
	rootCmd.MarkFlagRequired("kernel-cmdline")
	rootCmd.MarkFlagRequired("initrd")

	rootCmd.Flags().UintVarP(&opts.vcpus, "cpus", "c", 1, "number of virtual CPUs")
	// FIXME: use go-units for parsing
	rootCmd.Flags().UintVarP(&opts.memoryMiB, "memory", "m", 512, "virtual machine RAM size in mibibytes")

	// should all the options below become -s virtio-blk,options -s virtio-rng,options -s virtio-sock,options ? limits the amount of options, and should be more flexible when additional parameters are needed (virtio-vsock port, macaddress, ...)
	rootCmd.Flags().StringVarP(&opts.diskPath, "disk", "d", "", "path to the virtual machine raw disk image")
	// FIXME: missing port number
	rootCmd.Flags().StringVarP(&opts.vsockSocketPath, "virtio-vsock", "V", "", "path to the unix socket for virtio-vsock communication")

	// FIXME: move this to -n nat,macaddress?
	rootCmd.Flags().StringVarP(&opts.macAddress, "mac-address", "M", "", "virtual machine MAC address")
	rootCmd.Flags().BoolVarP(&opts.natNetworking, "nat", "n", false, "use NAT networking")
	rootCmd.Flags().BoolVarP(&opts.rngDevice, "rng", "r", false, "add RNG device")
	rootCmd.Flags().StringVarP(&opts.logFilePath, "log-file", "l", "", "path to log file for virtual machine console output")

	rootCmd.Flags().StringArrayVarP(&opts.devices, "device", "d", []string{}, "devices")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	Execute()
}
