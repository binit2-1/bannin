package ipc

// MemStatus is the output of `bannin mem --json`.
type MemStatus struct {
	V       int     `json:"v"`
	MemPct  float64 `json:"mem_pct"`
	SwapPct float64 `json:"swap_pct"`
	UsedGB  float64 `json:"used_gb"`
	TotalGB float64 `json:"total_gb"`
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

// ListResponse is the output of `bannin list --json`.
type ListResponse struct {
	V      int         `json:"v"`
	Groups []GroupInfo `json:"groups"`
}
