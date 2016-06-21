package state

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestContext(t *testing.T) {
	tmp, err := ioutil.TempDir("", "testing.")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	ctx, err := Initialize(tmp)
	if err != nil {
		t.Fatal(err)
	}

	if ctx.configPath() == "" {
		t.Error("expected a file path, but got nothing")
	}
	c, err := ctx.config()
	if err != nil {
		t.Error(err)
	}
	if c == nil {
		t.Errorf("this should never be nil")
	}

	m := ctx.NewMount()

	if err := ctx.PutMount(m); err != nil {
		t.Error(err)
	}
	mounts, err := ctx.Mounts()
	if err != nil {
		t.Error(err)
	}
	if len(mounts) != 1 {
		t.Errorf("expected 1 mount, but got %d", len(mounts))
	}
}
