cmdline="console=hvc0 BOOT_IMAGE=(hd0,gpt3)/ostree/rhcos-36fd944867b0e491991a65f6f3b7209c937fe3bd7cdbd855c7c5d5a7070ce570/vmlinuz-4.18.0-305.28.1.el8_4.x86_64 random.trust_cpu=on console=tty0 console=ttyS0,115200n8 ignition.platform.id=qemu ostree=/ostree/boot.1/rhcos/36fd944867b0e491991a65f6f3b7209c937fe3bd7cdbd855c7c5d5a7070ce570/0 root=UUID=91ba4914-fd2b-4a7c-b498-28585a80a40e rw rootflags=prjquota"

./vfkit -c 2 -m 2048 \
	-d virtio-blk,path=$HOME/.crc/machines/crc/crc.img \
	-i ~/.crc/cache/crc_hyperkit_4.9.10/initramfs-4.18.0-305.28.1.el8_4.x86_64.img -k ~/.crc/cache/crc_hyperkit_4.9.10/vmlinuz-4.18.0-305.28.1.el8_4.x86_64 -C "$cmdline" \
	-d virtio-net,nat \
	-d virtio-rng \
	-d virtio-vsock,port=1234,socketURL=$HOME/.crc/machines/crc/virtio-vsock-1234.sock
