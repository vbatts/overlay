package state

import (
	"io/ioutil"
	"os"
	"path/filepath"
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

	m := ctx.NewMountPoint()

	if err := ctx.SaveMount(m); err != nil {
		t.Error(err)
	}
	mounts, err := ctx.Mounts()
	if err != nil {
		t.Error(err)
	}

	// This ought to be 0, since we did not create the directories for the MountPoint, then it was garbage collected.
	if len(mounts) != 0 {
		t.Errorf("expected 0 mount, but got %d: %#v", len(mounts), mounts)
	}

	m = ctx.NewMountPoint()
	m.Source = filepath.Join(ctx.mountsPath(), m.UUID, "source")
	if err := m.Mkdir(0755); err != nil {
		t.Error(err)
	}
	if err := ctx.SaveMount(m); err != nil {
		t.Error(err)
	}
	mounts, err = ctx.Mounts()
	if err != nil {
		t.Error(err)
	}
	// This ought to be 1, since we did create the directories for the MountPoint
	if len(mounts) != 1 {
		t.Errorf("expected 1 mount, but got %d: %#v", len(mounts), mounts)
	}
}
