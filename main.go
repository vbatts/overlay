package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/vbatts/overlay/mount"
	"github.com/vbatts/overlay/state"
)

var (
	flSrc        = flag.String("src", "", "source directory to overlay")
	flTarget     = flag.String("target", "", "destination to overlay to (default is ${src}.overlay)")
	flUnmount    = flag.Bool("unmount", false, "unmount directory all provided args")
	flRemove     = flag.String("remove", "", "remove the provided UUID")
	flRoot       = flag.String("root", filepath.Join(os.Getenv("HOME"), ".local/share/overlay/"), "Directory to story state of previous overlay mounts")
	flListMounts = flag.Bool("list", false, "list previously recorded mounts")
	flDebug      = flag.Bool("debug", false, "enable debug output")
)

func main() {
	flag.Parse()
	if *flDebug {
		os.Setenv("DEBUG", "1")
	}

	if *flUnmount {
		for _, arg := range flag.Args() {
			if err := mount.UnmountPath(arg); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: unmounting %q: %s\n", arg, err)
				os.Exit(1)
			}
		}
		if *flRemove == "" {
			os.Exit(0)
		}
	}

	ctx, err := state.Initialize(*flRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}

	if *flRemove != "" {
		mounts, err := ctx.Mounts()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			os.Exit(1)
		}
		for _, m := range mounts {
			if m.UUID == *flRemove {
				if err := os.RemoveAll(m.Root()); err != nil {
					fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
					os.Exit(1)
				}
			}
		}
		os.Exit(0)
	}

	if *flListMounts {
		mounts, err := ctx.Mounts()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			os.Exit(1)
		}
		for _, m := range mounts {
			fmt.Printf("TARGET\t\tSOURCE\t\tUUID\n")
			fmt.Printf("%s\t\t%s\t\t%s\n", m.Target, m.Source, m.UUID)
		}
		os.Exit(0)
	}

	if *flSrc == "" {
		fmt.Fprintln(os.Stderr, "ERROR: no source directory provided")
		os.Exit(1)
	}

	// TODO check for supported underlying filesystems (ext4, xfs)

	m := ctx.NewMountPoint()
	m.Source, err = filepath.Abs(*flSrc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}

	if *flTarget != "" {
		m.Target, err = filepath.Abs(*flTarget)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			os.Exit(1)
		}
	}

	// check if target directory already mounted
	infos, err := mount.Infos()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: mountinfo: %s\n", err)
		os.Exit(1)
	}
	for i := range infos {
		if infos[i].MountPoint == m.Target {
			fmt.Fprintf(os.Stderr, "ERROR: %q is already mounted\n", m.Target)
			os.Exit(1)
		}
	}

	// TODO add cleanup mechanism if something later fails
	if err := m.Mkdir(0755); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
	fmt.Println(m.Target)

	if *flDebug {
		fmt.Println(m.Options())
	}

	if err := mount.Mount(m); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
	if err := ctx.SaveMount(m); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
}
