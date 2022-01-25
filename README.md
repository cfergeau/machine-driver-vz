CodeReady Containers Virtual Machine using macOS virtualization.framework
====

This implements a machine driver for CodeReady Containers using macOS virtualization framework.
This generates 2 binaries:
- vfkit which offers a command-line interface to start virtual machines using virtualization framework
- crc-driver-vf which is the machine driver implementation itself

The binaries are separate as crc-driver-vf is only running for a short time to execute commands, but the binary which starts the virtual machine must keep running for the whole lifetime of the VM.

In order to test this, you need to copy the crc-driver-vf and vfkit executables to ~/.crc/bin, and then use this crc branch: https://github.com/cfergeau/crc/tree/macos-vf
The vfkit executable must be signed after being copied to ~/.crc/bin: `codesign --force  --entitlements vf.entitlements -s - ~/.crc/bin/vfkit`

The machine driver currently depend on qemu-img to convert crc disk images from qcow2 to raw. qemu-img can be obtained from `brew install qemu`.

The work in this repository makes use of https://github.com/Code-Hex/vz to create a Linux virtual machine with virtualization.framework using go.
