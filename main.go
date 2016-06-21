package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
)

var (
	flSrc    = flag.String("src", "", "source directory to overlay")
	flDest   = flag.String("dest", "", "destination to overlay to (default is ${src}.overlay)")
	flUmount = flag.Bool("umount", false, "un-mount a source directory")
)

func main() {
	flag.Parse()
	if *flSrc == "" {
		fmt.Fprintln(os.Stderr, "ERROR: no source directory provided")
		os.Exit(1)
	}
	srcDir, err := filepath.Abs(*flSrc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}

	if *flUmount {
		if err := syscall.Unmount(srcDir, 0); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: unmounting: %s\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *flDest == "" {
		*flDest = fmt.Sprintf("%s.overlay", srcDir)
	}
	destDir, err := filepath.Abs(*flDest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
	// TODO check for destDir directory already mounted

	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		if err := os.Mkdir(destDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: making %q: %s\n", destDir, err)
			os.Exit(1)
		}

		// when the binary is setuid, the effective uid is 0, so reset these new directories to the user
		if err := os.Chown(destDir, os.Getuid(), os.Getgid()); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: owning %q: %s\n", destDir, err)
			os.Exit(1)
		}
	}
	fmt.Println(destDir)

	// TODO record this state of tmp directories somewhere, to show the user previous iterations or garbage collection
	tmpDir, err := ioutil.TempDir(filepath.Dir(destDir), filepath.Base(srcDir))
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
	// when the binary is setuid, the effective uid is 0, so reset these new directories to the user
	if err := os.Chown(tmpDir, os.Getuid(), os.Getgid()); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: owning %q: %s\n", tmpDir, err)
		os.Exit(1)
	}

	// TODO mkdir upperdir, workdir and merged dir
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

	optionData := fmt.Sprintf("lowerdir=%s,workdir=%s,upperdir=%s", srcDir, filepath.Join(tmpDir, "work"), filepath.Join(tmpDir, "upper"))
	//fmt.Println(optionData)

	if err := syscall.Mount(filepath.Join(tmpDir, "merged"), destDir, "overlay", 0, optionData); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
}
