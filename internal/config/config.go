// Package config reads ~/.config/bannin/config.toml. Every setting has a
// default, so a missing file (or a missing key) is fine.
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"github.com/binit2-1/bannin/internal/engine"
)

// Config is everything a user can set.
type Config struct {
	Thresholds engine.Thresholds `toml:"thresholds"`
	Notify     Notify            `toml:"notify"`
	AutoStop   AutoStop          `toml:"auto_stop"`
}

// Notify controls desktop notifications.
type Notify struct {
	Enabled         bool `toml:"enabled"`
	CooldownMinutes int  `toml:"cooldown_minutes"`
}

// AutoStop controls stopping dev tools without asking.
type AutoStop struct {
	Enabled bool `toml:"enabled"`
	// DryRun only logs what would be stopped. Leave it on for a few days
	// and read `journalctl --user -u bannind` before trusting it.
	DryRun bool `toml:"dry_run"`
	// Kinds of group that may be stopped: "metro", "gradle", "emulator", "devtool".
	// Browsers, editors and everything else are never auto-stopped.
	Kinds []string `toml:"kinds"`
	// OnlyDetached limits auto-stop to tools left running after their
	// terminal closed (the ⚠ ones).
	OnlyDetached bool `toml:"only_detached"`
	// AfterSeconds: RAM must stay critical this long before acting.
	AfterSeconds int `toml:"after_seconds"`
}

// Default is used for anything the file does not set.
func Default() Config {
	return Config{
		Thresholds: engine.DefaultThresholds,
		Notify:     Notify{Enabled: true, CooldownMinutes: 2},
		AutoStop: AutoStop{
			Enabled:      true,
			DryRun:       true,
			Kinds:        []string{"metro", "gradle", "emulator"},
			OnlyDetached: true,
			AfterSeconds: 10,
		},
	}
}

// Path is $XDG_CONFIG_HOME/bannin/config.toml.
func Path() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "bannin", "config.toml")
}

// Load reads path on top of the defaults. A missing file is not an error.
func Load(path string) (Config, error) {
	c := Default()
	_, err := toml.DecodeFile(path, &c)
	if errors.Is(err, fs.ErrNotExist) {
		return Default(), nil
	}
	if err != nil {
		return Default(), err
	}
	return c, c.Validate()
}

// Validate catches settings that would make Bannin behave strangely.
func (c Config) Validate() error {
	t := c.Thresholds
	if !(0 < t.Elevated && t.Elevated < t.High && t.High < t.Critical && t.Critical <= 100) {
		return fmt.Errorf("thresholds must go up: elevated %v < high %v < critical %v ≤ 100",
			t.Elevated, t.High, t.Critical)
	}
	if t.Hysteresis < 0 || t.Hysteresis > 20 {
		return fmt.Errorf("hysteresis %v should be between 0 and 20", t.Hysteresis)
	}
	ok := map[string]bool{"metro": true, "gradle": true, "emulator": true, "devtool": true}
	for _, k := range c.AutoStop.Kinds {
		if !ok[k] {
			return fmt.Errorf("auto_stop.kinds: %q is not a dev-tool kind (use metro, gradle, emulator, devtool)", k)
		}
	}
	return nil
}

// Example is written by `bannin config` when there is no file yet.
const Example = `# Bannin settings. Delete a line to use its default.
# After editing: systemctl --user reload bannind

[thresholds]
elevated = 75      # RAM % where the bar turns amber
high = 85          # orange, and a notification
critical = 92      # red; auto-stop may act
psi_high = 10      # memory pressure (% time stalled) that alone means "high"
psi_critical = 25
hysteresis = 3     # RAM must drop this far below a line to go down a level

[notify]
enabled = true
cooldown_minutes = 2

[auto_stop]
enabled = true
dry_run = true     # only log "would have stopped ..." — check the log before turning off
kinds = ["metro", "gradle", "emulator"]
only_detached = true   # only tools whose terminal was closed (marked ⚠)
after_seconds = 10     # RAM must stay critical this long first
`

// WriteExample creates the config file with Example if it does not exist.
// It returns true if it wrote one.
func WriteExample(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return false, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, err
	}
	return true, os.WriteFile(path, []byte(Example), 0o644)
}
