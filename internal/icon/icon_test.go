package icon

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFor(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("com.microsoft.VSCode.desktop", "[Desktop Entry]\nExec=/usr/share/code/code %F\nIcon=vscode\n")
	write("org.mozilla.firefox.desktop", "[Desktop Entry]\nExec=env MOZ=1 firefox %u\nIcon=org.mozilla.firefox\n[Desktop Action new]\nExec=other\n")
	write("obsidian.desktop", "[Desktop Entry]\nExec=\"/opt/Obsidian/obsidian\"\nIcon=obsidian\n")
	write("noicon.desktop", "[Desktop Entry]\nExec=noicon\n")

	r := New(dir)
	cases := []struct{ kind, name, want string }{
		{"metro", "Metro · app", "bundled:react"},
		{"gradle", "Gradle daemons", "bundled:gradle"},
		{"devtool", "adb", "bundled:android"},
		{"other", "code", "vscode"},
		{"other", "firefox", "org.mozilla.firefox"},
		{"other", "obsidian", "obsidian"},
		{"other", "other", Fallback}, // must not match the action's Exec
		{"other", "noicon", Fallback},
		{"other", "Firefox", "firefox"},
	}
	for _, c := range cases {
		if got := r.For(c.kind, c.name); got != c.want {
			t.Errorf("For(%q, %q) = %q, want %q", c.kind, c.name, got, c.want)
		}
	}
}
