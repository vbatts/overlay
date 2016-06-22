package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	"github.com/vbatts/overlay/mount"
	"github.com/vbatts/overlay/state"
)

var (
	flSrc        = flag.String("src", "", "source directory to overlay")
	flTarget     = flag.String("target", "", "destination to overlay to (default is ${src}.overlay)")
	flUnmount    = flag.Bool("unmount", false, "unmount directory all provided args")
	flRoot       = flag.String("root", filepath.Join(os.Getenv("HOME"), ".local/share/overlay/"), "Directory to story state of previous overlay mounts")
	flListMounts = flag.Bool("list", false, "list previously recorded mounts")
	flDebug      = flag.Bool("debug", false, "enable debug output")
)

func main() {
	flag.Parse()

	if *flUnmount {
		for _, arg := range flag.Args() {
			if err := mount.Unmount(arg); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: unmounting %q: %s\n", arg, err)
				os.Exit(1)
			}
		}
		os.Exit(0)
	}

	if *flDebug {
		os.Setenv("DEBUG", "1")
	}

	ctx, err := state.Initialize(*flRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}

	if *flSrc == "" {
		fmt.Fprintln(os.Stderr, "ERROR: no source directory provided")
		os.Exit(1)
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

	// TODO check for supported underlying filesystems (ext4, xfs)

	m := ctx.NewMount()
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

	// TODO check for targetDir directory already mounted

	if _, err := os.Stat(m.Target); os.IsNotExist(err) {
		if err := os.Mkdir(m.Target, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: making %q: %s\n", m.Target, err)
			os.Exit(1)
		}

		// when the binary is setuid, the effective uid is 0, so reset these new directories to the user
		if err := os.Chown(m.Target, os.Getuid(), os.Getgid()); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: owning %q: %s\n", m.Target, err)
			os.Exit(1)
		}
	}
	fmt.Println(m.Target)

	// TODO record this state of tmp directories somewhere, to show the user previous iterations or garbage collection
	tmpDir, err := ioutil.TempDir(filepath.Dir(m.Target), filepath.Base(m.Source))
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
	// when the binary is setuid, the effective uid is 0, so reset these new directories to the user
	if err := os.Chown(tmpDir, os.Getuid(), os.Getgid()); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: owning %q: %s\n", tmpDir, err)
		os.Exit(1)
	}

	for _, name := range []string{"upper", "work", "merged"} {
		if err := os.Mkdir(filepath.Join(tmpDir, name), 0755); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: making %q: %s\n", name, err)
			os.Exit(1)
		}
		// when the binary is setuid, the effective uid is 0, so reset these new directories to the user
		if err := os.Chown(filepath.Join(tmpDir, name), os.Getuid(), os.Getgid()); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: owning %q: %s\n", filepath.Join(tmpDir, name), err)
			os.Exit(1)
		}
	}

	optionData := fmt.Sprintf("lowerdir=%s,workdir=%s,upperdir=%s", m.Source, filepath.Join(tmpDir, "work"), filepath.Join(tmpDir, "upper"))
	if *flDebug {
		fmt.Println(optionData)
	}

	if err := syscall.Mount(filepath.Join(tmpDir, "merged"), m.Target, "overlay", 0, optionData); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
	if err := ctx.SaveMount(m); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
}
