package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
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

func actionUUID(context *cli.Context) error {
	if context.Bool("gen") {
		actionUUIDGen(context)
		return nil
	}
	if context.Bool("unmount") {
		for _, arg := range context.Args() {
			if err := mount.UnmountPath(arg); err != nil {
				log.Errorf("ERROR: unmounting %q: %s\n", arg, err)
			} else {
				fmt.Printf("unmounted %q\n", arg)
			}
		}
	}
	return nil
}
func actionUUIDGen(context *cli.Context) error {
	fmt.Println(state.NewUUID())
	return nil
}
func actionMounts(context *cli.Context) error {
	return nil
}
func actionMountsMount(context *cli.Context) error {
	return nil
}
func actionMountsUnmount(context *cli.Context) error {
	return nil
}
func actionMountsCreate(context *cli.Context) error {
	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "overlay"
	app.Usage = "overlayfs helper for limited users"
	app.Version = "0.1.0"
	app.Authors = []cli.Author{{Name: "@vbatts", Email: "vbatts@thisco.de"}}
	app.Before = preload
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "root",
			Value: filepath.Join(os.Getenv("HOME"), ".local/share/overlay/"),
			Usage: "Directory to story state of previous overlay mounts",
		},
		cli.BoolFlag{
			Name:  "debug",
			Usage: "enable debug output",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:  "mounts",
			Usage: "operations dealing with mount paths",
			Flags: []cli.Flag{},
			Subcommands: []cli.Command{
				{
					Name:   "mount",
					Usage:  "perform an overlay mount for the provided source directory",
					Action: actionMountsMount,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "src",
							Usage: "source directory to overlay",
						},
						cli.StringFlag{
							Name:  "target",
							Usage: "destination to overlay to (default a subdirectory of the -root)",
						},
					},
				},
				{
					Name:   "unmount",
					Usage:  "unmount all provided paths",
					Action: actionMountsUnmount,
				},
				{
					Name:   "create",
					Usage:  "prepare a new mount point",
					Flags:  []cli.Flag{},
					Action: actionMountsCreate,
				},
			},
			Action: actionMounts,
		},
		{
			Name:  "uuid",
			Usage: "operations via UUID",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "unmount",
					Usage: "unmount all provided UUIDs",
				},
				cli.BoolFlag{
					Name:  "gen",
					Usage: "output a generated UUID",
				},
			},
			Subcommands: []cli.Command{
				{
					Name:   "gen",
					Usage:  "output a generated UUID",
					Action: actionUUIDGen,
				},
			},
			Action: actionUUID,
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
	os.Exit(0)

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
		fmt.Printf("TARGET\t\tSOURCE\t\tUUID\n")
		for _, m := range mounts {
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

// do this before any subcommands
func preload(context *cli.Context) error {
	log.SetOutput(os.Stderr)
	if context.GlobalBool("debug") {
		os.Setenv("DEBUG", "1")
		log.SetLevel(log.DebugLevel)
	}
	return nil
}
