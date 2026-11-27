package ipc

// MemStatus is the output of `bannin mem --json`.
type MemStatus struct {
	V       int     `json:"v"`
	MemPct  float64 `json:"mem_pct"`
	SwapPct float64 `json:"swap_pct"`
	UsedGB  float64 `json:"used_gb"`
	TotalGB float64 `json:"total_gb"`

	Level   string  `json:"level"`             // "ok", "elevated", "high", "critical"
	PSI     float64 `json:"psi"`               // memory pressure, some avg10
	TTFSec  int     `json:"ttf_s,omitempty"`   // seconds until RAM is full; 0 = not filling
	Warn    int     `json:"warn"`              // detached dev-tool groups
	Biggest string  `json:"biggest,omitempty"` // name of the biggest group
	Service bool    `json:"service"`           // answered by bannind (not read directly)

	// Only filled in by the service:
	History      []float64    `json:"history,omitempty"`       // RAM % for the last 60 s, oldest first
	SnoozedUntil int64        `json:"snoozed_until,omitempty"` // unix seconds
	Fake         bool         `json:"fake,omitempty"`
	AutoStop     bool         `json:"auto_stop"`
	DryRun       bool         `json:"dry_run"`
	Actions      []ActionInfo `json:"actions,omitempty"` // newest first
}

// ActionInfo is one thing Bannin did, or in dry run would have done.
type ActionInfo struct {
	Time   int64  `json:"t"` // unix seconds
	Text   string `json:"text"`
	DryRun bool   `json:"dry_run,omitempty"`
}

// GroupInfo is one line in the popup.
type GroupInfo struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Kind     string  `json:"kind"`
	RSSMB    float64 `json:"rss_mb"`
	PIDs     []int   `json:"pids"`
	Detached bool    `json:"detached"`
	Icon     string  `json:"icon"` // theme icon name, or "bundled:<name>"
}

// Result answers stop and snooze requests.
type Result struct {
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ListResponse is the output of `bannin list --json`.
type ListResponse struct {
	V      int         `json:"v"`
	Groups []GroupInfo `json:"groups"`
}
