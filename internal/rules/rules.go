// Package rules decides which process groups auto-stop may touch.
package rules

import (
	"slices"

	"github.com/binit2-1/bannin/internal/config"
	"github.com/binit2-1/bannin/internal/group"
)

// Pick returns the groups that auto-stop may stop, biggest first.
// Only dev-tool kinds listed in the config qualify; "other" never does,
// so browsers, editors and the desktop are safe whatever the config says.
func Pick(groups []group.Group, a config.AutoStop) []group.Group {
	var out []group.Group
	for _, g := range groups {
		if g.Kind == "other" || !slices.Contains(a.Kinds, g.Kind) {
			continue
		}
		if a.OnlyDetached && !g.Detached {
			continue
		}
		out = append(out, g)
	}
	return out
}
