package state

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pborman/uuid"
	"github.com/vbatts/overlay/types"
)

// Config contains the metadata for mount information
type Config struct {
	Mounts []types.Mount
}

// Context is used when
type Context struct {
	Root string
}

// Initialize provides a Context from a state root directory
func Initialize(root string) (*Context, error) {
	return &Context{Root: root}, nil
}

// Mounts is the list of all mounts from the Context's Config (all stored mounts)
func (ctx *Context) Mounts() ([]types.Mount, error) {
	c, err := ctx.config()
	if err != nil {
		return nil, err
	}
	return c.Mounts, nil
}

// SaveMount stores a Mount to the Context's Config
func (ctx *Context) SaveMount(m types.Mount) error {
	c, err := ctx.config()
	if err != nil {
		return err
	}
	c.Mounts = append(c.Mounts, m)
	return ctx.putConfig(c)
}

// NewUUID provides a UUID based on RFC 4122 and DCE 1.1
func NewUUID() string {
	return uuid.New()
}

// NewMount prepares a Mount context with a new UUID.
// Once the Mount is populated, it must be saved with SaveMount
func (ctx *Context) NewMount() types.Mount {
	u := NewUUID()
	return types.Mount{
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
	var c Config
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
