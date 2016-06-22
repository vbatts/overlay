package state

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pborman/uuid"
	"github.com/vbatts/overlay/types"
)

// Config contains the metadata for mount information
type Config struct {
	Mounts map[string]types.MountPoint // map[UUID]MountPoint
}

// Context is used when
type Context struct {
	Root string
}

// Initialize provides a Context from a state root directory
func Initialize(root string) (*Context, error) {
	for _, dir := range []string{root, filepath.Join(root, "mounts")} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("making %q: %s\n", dir, err)
		}
		if os.Getuid() != os.Geteuid() {
			if err := os.Chown(dir, os.Getuid(), os.Getgid()); err != nil {
				return nil, fmt.Errorf("owning %q: %s\n", dir, err)
			}
		}
	}
	return &Context{Root: root}, nil
}

// Mounts is the list of all mounts from the Context's Config (all stored mounts)
func (ctx *Context) Mounts() (map[string]types.MountPoint, error) {
	// FIXME this will get out of sync if there are failures to Mount.
	// Better to just glob this directory for the listing.
	// Also, if the Target director is not in the mounts/$uuid/ directory, make a symlink from 'rootfs' to the target directory

	c, err := ctx.config()
	if err != nil {
		return nil, err
	}

	// any mount in the config that do not exist on disk, need to be pruned
	for mUUID := range c.Mounts {
		if _, err := os.Stat(filepath.Join(ctx.mountsPath(), mUUID)); os.IsNotExist(err) {
			delete(c.Mounts, mUUID)
		}
	}

	if err := ctx.putConfig(c); err != nil {
		return nil, err
	}

	return c.Mounts, nil
}

// SaveMount stores a MountPoint to the Context's Config
func (ctx *Context) SaveMount(m types.MountPoint) error {
	if m.UUID == "" {
		return fmt.Errorf("MountPoint.UUID must be set")
	}
	c, err := ctx.config()
	if err != nil {
		return err
	}
	c.Mounts[m.UUID] = m
	return ctx.putConfig(c)
}

// NewUUID provides a UUID based on RFC 4122 and DCE 1.1
func NewUUID() string {
	return uuid.New()
}

// NewMountPoint prepares a MountPoint context with a new UUID.
// Once the MountPoint is populated, it must be saved with SaveMount
func (ctx *Context) NewMountPoint() types.MountPoint {
	u := NewUUID()
	return types.MountPoint{
		UUID:   u,
		Target: filepath.Join(ctx.mountsPath(), u, "rootfs"),
		Upper:  filepath.Join(ctx.mountsPath(), u, "upper"),
		Work:   filepath.Join(ctx.mountsPath(), u, "work"),
		Merge:  filepath.Join(ctx.mountsPath(), u, "merge"),
	}
}

func (ctx *Context) putConfig(c *Config) error {
	data, err := json.Marshal(*c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(ctx.configPath(), data, 0644)
}

func (ctx *Context) config() (*Config, error) {
	c := Config{Mounts: map[string]types.MountPoint{}}
	data, err := ioutil.ReadFile(ctx.configPath())
	if err != nil {
		if os.IsNotExist(err) {
			if err := ctx.putConfig(&c); err != nil {
				return nil, err
			}
			return &c, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (ctx *Context) configPath() string {
	return filepath.Join(ctx.Root, "config.json")
}

func (ctx *Context) mountsPath() string {
	return filepath.Join(ctx.Root, "mounts")
}
