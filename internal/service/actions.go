package service

import (
	"log"
	"sync"
	"time"

	"github.com/binit2-1/bannin/internal/ipc"
)

// ActionLog remembers the last few things Bannin did (or would have done).
type ActionLog struct {
	mu    sync.Mutex
	items []ipc.ActionInfo // newest last
}

// Add records an action and writes it to the log (journalctl).
func (l *ActionLog) Add(text string, dryRun bool) {
	if dryRun {
		log.Printf("dry run: %s", text)
	} else {
		log.Print(text)
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.items = append(l.items, ipc.ActionInfo{Time: time.Now().Unix(), Text: text, DryRun: dryRun})
	if len(l.items) > 20 {
		l.items = l.items[len(l.items)-20:]
	}
}

// Recent returns up to n actions, newest first.
func (l *ActionLog) Recent(n int) []ipc.ActionInfo {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]ipc.ActionInfo, 0, n)
	for i := len(l.items) - 1; i >= 0 && len(out) < n; i-- {
		out = append(out, l.items[i])
	}
	return out
}
