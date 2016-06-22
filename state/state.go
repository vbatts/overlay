package state

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pborman/uuid"
	"github.com/vbatts/overlay/types"
)

type Config struct {
	Mounts []types.Mount
}

type Context struct {
	Root string
}

func (ctx *Context) Mounts() ([]types.Mount, error) {
	c, err := ctx.config()
	if err != nil {
		return nil, err
	}
	return c.Mounts, nil
}

func (ctx *Context) PutMount(m types.Mount) error {
	c, err := ctx.config()
	if err != nil {
		return err
	}
	c.Mounts = append(c.Mounts, m)
	return ctx.putConfig(c)
}

// NewMount prepares a Mount context with a new UUID.
// Once the Mount is populated, it must be saved with PutMount
func (ctx *Context) NewMount() types.Mount {
	u := uuid.New()
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

func Initialize(root string) (*Context, error) {
	return &Context{Root: root}, nil
}
