package ipc

import (
	"bufio"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Request is one call to the service: a single JSON line.
type Request struct {
	Method  string `json:"method"` // "status", "list", "stop", "snooze"
	ID      string `json:"id,omitempty"`
	Force   bool   `json:"force,omitempty"`
	All     bool   `json:"all,omitempty"`
	Minutes int    `json:"minutes,omitempty"`
}

// SocketPath is where bannind listens: $XDG_RUNTIME_DIR/bannin.sock,
// a folder only this user can open.
func SocketPath() string {
	dir := os.Getenv("XDG_RUNTIME_DIR")
	if dir == "" {
		dir = filepath.Join(os.TempDir(), "bannin-"+strconv.Itoa(os.Getuid()))
	}
	return filepath.Join(dir, "bannin.sock")
}

// ErrNoService means bannind is not running.
var ErrNoService = errors.New("bannind is not running")

// Call sends req and decodes the one-line reply into out.
// A reply of the form {"error": "..."} becomes an error.
func Call(req Request, out any) error {
	conn, err := net.DialTimeout("unix", SocketPath(), 300*time.Millisecond)
	if err != nil {
		return ErrNoService
	}
	defer conn.Close()
	// stop can take up to ~6 s (SIGTERM, wait, SIGKILL).
	conn.SetDeadline(time.Now().Add(10 * time.Second))

	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return err
	}
	line, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		return err
	}
	var e struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(line, &e) == nil && e.Error != "" {
		return errors.New(e.Error)
	}
	return json.Unmarshal(line, out)
}
