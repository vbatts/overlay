package mount

import (
	"errors"
	"syscall"

	"github.com/vbatts/overlay/types"
)

// UnmountMount performs a UMOUNT(2) on the Target for the provided Mount
func UnmountMount(m types.Mount) error {
	return Unmount(m.Target)
}

// Unmount performs a UMOUNT(2) on the target directory
func Unmount(targetDir string) error {
	if targetDir == "" {
		return errors.New("mount target is empty")
	}
	return syscall.Unmount(targetDir, 0)
}
