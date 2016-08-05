# overlay

[![Build Status](https://travis-ci.org/vbatts/overlay.svg?branch=master)](https://travis-ci.org/vbatts/overlay)

A helper for setting up overlay filesystem mounts in Linux.
This helper is aimed at being usable by limited users.

## Installing

While this application is `go get`'able, for the `mount()` command, the binary
needs to be owned by root and setuid.

```bash
$ go get github.com/vbatts/overlay
$ sudo dd if=$GOPATH/bin/overlay of=/usr/local/bin/overlay
$ sudo chmod +x /usr/local/bin/overlay
```

Since this binary is doing mount() and umount(), it needs privileges
(unfortunately those privileges are within CAP_SYS_ADMIN).

So either run:
```bash
sudo setcap cap_sys_admin+ep /usr/local/bin/overlay
```
or setuid of:
```bash
sudo chmod s+u /usr/local/bin/overlay
```

## Usage

Here is making a directory, mounting it in an overlay, and showing the
differences after modifications.  All as a limited user.

```bash
$ mkdir x
$ touch x/file
$ overlay -src ./x
/home/vbatts/.local/share/overlay/mounts/43b2d008-706e-4378-ae18-b86066ecd0d6/rootfs
$ cd /home/vbatts/.local/share/overlay/mounts/43b2d008-706e-4378-ae18-b86066ecd0d6/rootfs
$ ls
file
$ touch file2
$ rm file
$ ls
file2
$ cd -
/home/vbatts
$ ls ./x
file
```

Here is unmounting the overlay as a limited user.

```bash
$ findmnt | grep -w overlay
│ └─/home/vbatts/.local/share/overlay/mounts/43b2d008-706e-4378-ae18-b86066ecd0d6/rootfs /home/vbatts/.local/share/overlay/mounts/43b2d008-706e-4378-ae18-b86066ecd0d6/merge overlay  rw,relatime,lowerdir=/home/vbatts/x,upperdir=/home/vbatts/.local/share/overlay/mounts/43b2d008-706e-4378-ae18-b86066ecd0d6/upper,workdir=/home/vbatts/.local/share/overlay/mounts/43b2d008-706e-4378-ae18-b86066ecd0d6/work
$ overlay -unmount /home/vbatts/.local/share/overlay/mounts/43b2d008-706e-4378-ae18-b86066ecd0d6/rootfs
$ echo $?
0
$ overlay -list
TARGET          SOURCE          UUID
/home/vbatts/.local/share/overlay/mounts/43b2d008-706e-4378-ae18-b86066ecd0d6/rootfs            /home/vbatts/x          43b2d008-706e-4378-ae18-b86066ecd0d6
$ overlay -remove 43b2d008-706e-4378-ae18-b86066ecd0d6
$ overlay -list
```
