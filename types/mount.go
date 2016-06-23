package types

import (
	"fmt"
	"os"
	"path/filepath"
)

// MountPoint type is used for mounting and storing state
type MountPoint struct {
	UUID               string
	Source, Target     string
	Upper, Work, Merge string
}

// Options show the mount options for the given directory points
func (mp MountPoint) Options() string {
	return fmt.Sprintf("lowerdir=%s,workdir=%s,upperdir=%s", mp.Source, mp.Work, mp.Upper)
}

// Mkdir sets up the directories for this MountPoint
func (mp MountPoint) Mkdir(perm os.FileMode) error {
	for _, dir := range []string{mp.Source, mp.Target, mp.Upper, mp.Work, mp.Merge} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("making %q: %s\n", dir, err)
			}

			// when the binary is setuid, the effective uid is 0, so reset these new directories to the user
			if os.Getuid() != os.Geteuid() {
				if err := os.Chown(dir, os.Getuid(), os.Getgid()); err != nil {
					return fmt.Errorf("owning %q: %s\n", dir, err)
				}
			}
		}
	}
	// and chown the parent directory too
	if os.Getuid() != os.Geteuid() {
		if err := os.Chown(mp.Root(), os.Getuid(), os.Getgid()); err != nil {
			return fmt.Errorf("owning %q: %s\n", mp.Work, err)
		}
	}
	return nil
}

// Root provides the base path of the working directories
func (mp MountPoint) Root() string {
	return filepath.Dir(mp.Work)
}
