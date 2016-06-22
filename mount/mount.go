package mount

import (
	"syscall"

	"github.com/vbatts/overlay/types"
)

// Mount performs an overlay mount for MountPoint
func Mount(m types.MountPoint) error {
	return syscall.Mount(m.Merge, m.Target, "overlay", 0, m.Options())
}
