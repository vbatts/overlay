package mount

import (
	"errors"
	"syscall"

	"github.com/vbatts/overlay/types"
)

func UnmountMount(m types.Mount) error {
	if m.Target == "" {
		return errors.New("mount target is empty")
	}
	return Unmount(m.Target)
}

func Unmount(targetDir string) error {
	return syscall.Unmount(targetDir, 0)
}
