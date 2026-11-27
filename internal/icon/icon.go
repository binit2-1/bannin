// Package icon picks an icon for a process group.
//
// The result is either "bundled:<name>", meaning an SVG shipped with the
// widget in contents/icons/<name>.svg, or a name from the icon theme
// (for example "firefox"), which Kirigami.Icon can show directly.
package icon

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Fallback is used when nothing better is found.
const Fallback = "application-x-executable"

// byKind covers dev tools that have no .desktop file of their own.
var byKind = map[string]string{
	"metro":    "bundled:react",
	"gradle":   "bundled:gradle",
	"emulator": "bundled:android",
}

// byName covers single processes whose name does not match their app.
var byName = map[string]string{
	"adb":          "bundled:android",
	"node":         "bundled:nodejs",
	"watchman":     "system-search",
	"chrome":       "google-chrome",
	"gopls":        "text-x-go",
	"zsh":          "utilities-terminal",
	"bash":         "utilities-terminal",
	"fish":         "utilities-terminal",
	"claude":       "utilities-terminal",
	"codex":        "utilities-terminal",
	"java":         "application-x-java",
	"kwin_wayland": "kwin",
	"Firefox":      "firefox",
}

// Resolver maps lowercase process names to icon names using .desktop files.
type Resolver struct {
	icons map[string]string
}

// DefaultDirs are the places .desktop files usually live.
func DefaultDirs() []string {
	home, _ := os.UserHomeDir()
	data := os.Getenv("XDG_DATA_HOME")
	if data == "" {
		data = filepath.Join(home, ".local/share")
	}
	return []string{
		filepath.Join(data, "applications"),
		filepath.Join(data, "flatpak/exports/share/applications"),
		"/var/lib/flatpak/exports/share/applications",
		"/usr/local/share/applications",
		"/usr/share/applications",
	}
}

// New reads every .desktop file in dirs. Earlier dirs win.
func New(dirs ...string) *Resolver {
	r := &Resolver{icons: map[string]string{}}
	for _, dir := range dirs {
		files, _ := filepath.Glob(filepath.Join(dir, "*.desktop"))
		for _, f := range files {
			r.addFile(f)
		}
	}
	return r
}

// For returns the icon for a group of the given kind and name.
func (r *Resolver) For(kind, name string) string {
	if ic, ok := byKind[kind]; ok {
		return ic
	}
	if ic, ok := byName[name]; ok {
		return ic
	}
	if ic, ok := r.icons[strings.ToLower(name)]; ok {
		return ic
	}
	return Fallback
}

func (r *Resolver) addFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	var icon, exec, wmClass string
	inEntry := false
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "[") {
			// Only the main section counts; actions have their own Exec lines.
			inEntry = line == "[Desktop Entry]"
			continue
		}
		if !inEntry {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		switch key {
		case "Icon":
			icon = val
		case "Exec":
			exec = val
		case "StartupWMClass":
			wmClass = val
		}
	}
	if sc.Err() != nil || icon == "" {
		return
	}

	// "com.microsoft.VSCode.desktop" gives "com.microsoft.vscode" and "vscode".
	id := strings.TrimSuffix(filepath.Base(path), ".desktop")
	keys := []string{id, lastPart(id), wmClass, execName(exec)}
	for _, k := range keys {
		k = strings.ToLower(k)
		if k == "" {
			continue
		}
		if _, taken := r.icons[k]; !taken {
			r.icons[k] = icon
		}
	}
}

// execName returns the program name from an Exec line, so
// `env FOO=1 "/usr/share/code/code" %F` gives "code".
func execName(exec string) string {
	for f := range strings.FieldsSeq(exec) {
		f = strings.Trim(f, `"'`)
		if f == "env" || strings.Contains(f, "=") {
			continue
		}
		return filepath.Base(f)
	}
	return ""
}

func lastPart(id string) string {
	if i := strings.LastIndex(id, "."); i >= 0 {
		return id[i+1:]
	}
	return ""
}
