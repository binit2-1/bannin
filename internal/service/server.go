package service

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/binit2-1/bannin/internal/group"
	"github.com/binit2-1/bannin/internal/icon"
	"github.com/binit2-1/bannin/internal/ipc"
	"github.com/binit2-1/bannin/internal/report"
)

// Server answers CLI and widget requests on a Unix socket.
type Server struct {
	svc    *Service
	alerts *Alerts // may be nil when notifications are off

	mu       sync.Mutex
	groups   []group.Group
	groupsAt time.Time
	icons    *icon.Resolver
	iconsAt  time.Time
}

// NewServer makes a server for svc. alerts may be nil.
func NewServer(svc *Service, alerts *Alerts) *Server {
	return &Server{svc: svc, alerts: alerts}
}

// Listen serves on path until ctx is cancelled.
func (s *Server) Listen(ctx context.Context, path string) error {
	// A socket file left over from a crash would make Listen fail.
	// If something answers on it, another bannind is running.
	if c, err := net.DialTimeout("unix", path, 200*time.Millisecond); err == nil {
		c.Close()
		return fmt.Errorf("another bannind is already listening on %s", path)
	}
	os.Remove(path)

	ln, err := net.Listen("unix", path)
	if err != nil {
		return err
	}
	os.Chmod(path, 0o600)
	go func() {
		<-ctx.Done()
		ln.Close() // also removes the socket file
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			log.Printf("accept: %v", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(10 * time.Second))

	line, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		return
	}
	var req ipc.Request
	var reply any
	if err := json.Unmarshal(line, &req); err != nil {
		reply = ipc.Result{Error: "bad request"}
	} else {
		reply = s.dispatch(req)
	}
	json.NewEncoder(conn).Encode(reply)
}

func (s *Server) dispatch(req ipc.Request) any {
	switch req.Method {
	case "status":
		return s.status()
	case "list":
		minMB := 100.0
		if req.All {
			minMB = 0
		}
		groups, icons := s.cached()
		return report.List(groups, minMB, icons)
	case "stop":
		res := report.Stop(req.ID, req.Force)
		if res.OK {
			log.Print(res.Message + " (by request)")
			s.invalidate()
		}
		return res
	case "snooze":
		if s.alerts == nil {
			return ipc.Result{Error: "notifications are off"}
		}
		d := time.Duration(req.Minutes) * time.Minute
		if d <= 0 {
			d = SnoozeFor
		}
		s.alerts.Snooze(d)
		return ipc.Result{OK: true, Message: fmt.Sprintf("snoozed for %v", d)}
	}
	return ipc.Result{Error: "unknown method " + req.Method}
}

func (s *Server) status() ipc.MemStatus {
	st := s.svc.Status()
	totalGB := float64(st.Total) / 1024 / 1024
	out := ipc.MemStatus{
		V:       1,
		MemPct:  st.Sample.MemPct,
		SwapPct: st.Sample.SwapPct,
		UsedGB:  totalGB * st.Sample.MemPct / 100,
		TotalGB: totalGB,
		Level:   st.Level.String(),
		PSI:     st.Sample.PSI,
		Service: true,
		Fake:    st.Fake,
	}
	if st.HasTTF {
		out.TTFSec = max(1, int(st.TTF.Seconds()))
	}
	for _, h := range s.svc.History(60 * time.Second) {
		out.History = append(out.History, h.MemPct)
	}
	if s.alerts != nil {
		if t := s.alerts.SnoozedUntil(); t.After(time.Now()) {
			out.SnoozedUntil = t.Unix()
		}
	}
	groups, _ := s.cached()
	report.AddGroupSummary(&out, groups)
	return out
}

// cached returns the process groups, re-reading /proc at most every 2 s
// (the widget polls every 2 s), and the icon lookup, rebuilt every 10 min.
func (s *Server) cached() ([]group.Group, *icon.Resolver) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	if now.Sub(s.groupsAt) > 1500*time.Millisecond {
		if g, err := report.Groups(); err == nil {
			s.groups, s.groupsAt = g, now
		}
	}
	if s.icons == nil || now.Sub(s.iconsAt) > 10*time.Minute {
		s.icons, s.iconsAt = icon.New(icon.DefaultDirs()...), now
	}
	return s.groups, s.icons
}

func (s *Server) invalidate() {
	s.mu.Lock()
	s.groupsAt = time.Time{}
	s.mu.Unlock()
}
