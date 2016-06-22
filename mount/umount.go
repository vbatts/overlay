package mount

import (
	"errors"
	"syscall"

	"github.com/vbatts/overlay/types"
)

// UnmountMount performs a UMOUNT(2) on the Target for the provided MountPoint
func UnmountMount(m types.MountPoint) error {
	return UnmountPath(m.Target)
}

// UnmountPath performs a UMOUNT(2) on the target directory
func UnmountPath(targetDir string) error {
	if targetDir == "" {
		return errors.New("mount target is empty")
	}
	return syscall.Unmount(targetDir, 0)
}
