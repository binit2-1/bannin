package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestMissingFileGivesDefaults(t *testing.T) {
	c, err := Load(filepath.Join(t.TempDir(), "nope.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(c, Default()) {
		t.Errorf("got %+v", c)
	}
	if !c.AutoStop.DryRun {
		t.Error("dry run must be on by default")
	}
}

func TestExampleMatchesDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bannin", "config.toml")
	if wrote, err := WriteExample(path); err != nil || !wrote {
		t.Fatal(wrote, err)
	}
	c, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(c, Default()) {
		t.Errorf("example differs from defaults:\n%+v\n%+v", c, Default())
	}
	if wrote, _ := WriteExample(path); wrote {
		t.Error("overwrote an existing file")
	}
}

func TestPartialFileKeepsOtherDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "c.toml")
	os.WriteFile(path, []byte("[auto_stop]\ndry_run = false\n"), 0o644)
	c, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if c.AutoStop.DryRun || !c.AutoStop.Enabled || c.Thresholds.High != 85 {
		t.Errorf("got %+v", c)
	}
}

func TestValidate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "c.toml")
	os.WriteFile(path, []byte("[thresholds]\nhigh = 70\n"), 0o644)
	if _, err := Load(path); err == nil {
		t.Error("high below elevated should fail")
	}
	os.WriteFile(path, []byte("[auto_stop]\nkinds = [\"other\"]\n"), 0o644)
	if _, err := Load(path); err == nil {
		t.Error("kind other should fail")
	}
}
