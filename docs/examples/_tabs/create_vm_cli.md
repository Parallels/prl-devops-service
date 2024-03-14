---
layout: page
---

for creating the machine

```sh
prlctl create "test-vm" -d ubuntu
prlctl set "test-vm" --cpus 2
prlctl set "test-vm" --memsize 2048
prlctl set "test-vm" --device-set hdd0 --size 64G
```

Now we need to set the ISO file as the boot device and start the virtual machine and start it.

```sh
prlctl set "test-vm" --device-set cdrom0 --image /path/to/ubuntu-22.04.4-live-server-amd64.iso --connect
prlctl set "test-vm" --device-bootorder "cdrom0 hdd0"
prlctl start "test-vm"
```