package rules

import (
	"testing"

	"github.com/binit2-1/bannin/internal/config"
	"github.com/binit2-1/bannin/internal/group"
)

func names(gs []group.Group) []string {
	var out []string
	for _, g := range gs {
		out = append(out, g.Name)
	}
	return out
}

func TestPick(t *testing.T) {
	groups := []group.Group{ // biggest first, as group.Build returns them
		{Name: "Firefox", Kind: "other", Detached: true},
		{Name: "Gradle daemons", Kind: "gradle", Detached: true},
		{Name: "Metro · app", Kind: "metro", Detached: false},
		{Name: "Android emulator", Kind: "emulator", Detached: true},
		{Name: "adb", Kind: "devtool", Detached: true},
	}
	a := config.Default().AutoStop
	got := names(Pick(groups, a))
	want := []string{"Gradle daemons", "Android emulator"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("detached only: got %v, want %v", got, want)
	}

	a.OnlyDetached = false
	a.Kinds = append(a.Kinds, "other") // must still never pick Firefox
	got = names(Pick(groups, a))
	if len(got) != 3 || got[0] != "Gradle daemons" {
		t.Errorf("all dev tools: got %v", got)
	}
}
