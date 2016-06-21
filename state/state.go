package state

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pborman/uuid"
)

type Config struct {
	Mounts []Mount
}

type Mount struct {
	UUID               string
	Source, Target     string
	Upper, Work, Merge string
}

type Context struct {
	Root string
}

func (ctx *Context) Mounts() ([]Mount, error) {
	c, err := ctx.config()
	if err != nil {
		return nil, err
	}
	return c.Mounts, nil
}

func (ctx *Context) PutMount(m Mount) error {
	c, err := ctx.config()
	if err != nil {
		return err
	}
	c.Mounts = append(c.Mounts, m)
	return ctx.putConfig(c)
}

// NewMount prepares a Mount context with a new UUID.
// Once the Mount is populated, it must be saved with PutMount
func (ctx *Context) NewMount() Mount {
	return Mount{UUID: uuid.New()}
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

func Initialize(root string) (*Context, error) {
	return &Context{Root: root}, nil
}
