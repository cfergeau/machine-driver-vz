CodeReady Containers Virtual Machine using macOS virtualization.framework
====

The work in this repository makes use of https://github.com/Code-Hex/vz to create a Linux virtual machine with virtualization.framework using go.
After building it with `make`, the `machine-driver-vz` executable must be run with the name of a bundle (eg `crc_hyperkit_4.8.4`), and a virtual machine for this bunudle will be started.
The bundle must be unpacked in `~/.crc/cache`. The `qcow2` image it contains is first converted to raw using `qemu-img` (which must be installed).

The current code stops the virtual machine after 3 minutes. This behaviour is unchanged from https://github.com/Code-Hex/vz/blob/master/example/main.go
