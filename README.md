# overlay

A helper for setting up overlay filesystem mounts in Linux.
This helper is aimed at being usable by limited users.

## Installing

While this application is `go get`'able, for the `mount()` command, the binary needs to be owned by root and setuid.

```bash
$ go get github.com/vbatts/overlay
$ sudo dd if=$GOPATH/bin/overlay of=/usr/local/bin/overlay
$ sudo chmod 4755 /usr/local/bin/overlay
```

## Usage

Here is making a directory, mounting it in an overlay, and showing the differences after modifications.
All as a limited user.

```bash
vbatts@bananaboat ~ $ mkdir x
vbatts@bananaboat ~ $ touch x/file
vbatts@bananaboat ~ $ overlay -src ./x
/home/vbatts/x.overlay
vbatts@bananaboat ~ $ cd ./x.overlay
vbatts@bananaboat ~/x.overlay $ ls
file
vbatts@bananaboat ~/x.overlay $ touch file2
vbatts@bananaboat ~/x.overlay $ rm file
vbatts@bananaboat ~/x.overlay $ ls
file2
vbatts@bananaboat ~/x.overlay $ ls ../x
file
```

Here is unmounting the overlay as a limited user.

```bash
vbatts@bananaboat ~ (master *) $ findmnt | grep -w x
x mq/home/vbatts/x.overlay      /home/vbatts/x796818151/merged overlay         rw,relatime,lowerdir=/home/vbatts/x,upperdir=/home/vbatts/x796818151/upper,workdir=/home/vbatts/x796818151/work
vbatts@bananaboat ~ $ overlay -unmount ./x.overlay 
vbatts@bananaboat ~ $ echo $?
0
```
