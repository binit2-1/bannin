// Package notify shows desktop notifications over D-Bus
// (org.freedesktop.Notifications) and reports which button was pressed.
package notify

import (
	"sync"

	"github.com/godbus/dbus/v5"
)

const (
	dest  = "org.freedesktop.Notifications"
	path  = dbus.ObjectPath("/org/freedesktop/Notifications")
	iface = "org.freedesktop.Notifications"
)

// Urgency levels from the notification spec.
const (
	Low      byte = 0
	Normal   byte = 1
	Critical byte = 2
)

// Notification is one popup. Actions alternate key and label:
// []string{"stop", "Stop Metro", "snooze", "Snooze 10 min"}.
type Notification struct {
	Summary string
	Body    string
	Icon    string
	Urgency byte
	Actions []string
}

// Action is a button the user pressed.
type Action struct {
	ID  uint32
	Key string
}

// Notifier talks to the notification server on the session bus.
type Notifier struct {
	conn    *dbus.Conn
	actions chan Action

	mu     sync.Mutex
	lastID uint32 // reused so a new warning replaces the old one
}

// New connects to the session bus and starts listening for button presses.
func New() (*Notifier, error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil, err
	}
	if err := conn.AddMatchSignal(
		dbus.WithMatchObjectPath(path),
		dbus.WithMatchInterface(iface),
		dbus.WithMatchMember("ActionInvoked"),
	); err != nil {
		conn.Close()
		return nil, err
	}

	n := &Notifier{conn: conn, actions: make(chan Action, 8)}
	signals := make(chan *dbus.Signal, 16)
	conn.Signal(signals)
	go n.listen(signals)
	return n, nil
}

// Actions delivers the buttons the user presses on our notifications.
func (n *Notifier) Actions() <-chan Action { return n.actions }

// Show displays (or replaces) Bannin's notification and returns its ID.
func (n *Notifier) Show(nt Notification) (uint32, error) {
	n.mu.Lock()
	replaces := n.lastID
	n.mu.Unlock()

	hints := map[string]dbus.Variant{
		"urgency":       dbus.MakeVariant(nt.Urgency),
		"desktop-entry": dbus.MakeVariant("bannin"),
	}
	if nt.Actions == nil {
		nt.Actions = []string{}
	}
	timeout := int32(-1) // server default
	if nt.Urgency == Critical {
		timeout = 0 // stay until dismissed
	}

	var id uint32
	err := n.conn.Object(dest, path).Call(iface+".Notify", 0,
		"Bannin", replaces, nt.Icon, nt.Summary, nt.Body,
		nt.Actions, hints, timeout,
	).Store(&id)
	if err != nil {
		return 0, err
	}
	n.mu.Lock()
	n.lastID = id
	n.mu.Unlock()
	return id, nil
}

// Close closes the D-Bus connection.
func (n *Notifier) Close() error { return n.conn.Close() }

func (n *Notifier) listen(signals <-chan *dbus.Signal) {
	defer close(n.actions)
	for sig := range signals {
		if sig.Name != iface+".ActionInvoked" || len(sig.Body) < 2 {
			continue
		}
		id, _ := sig.Body[0].(uint32)
		key, _ := sig.Body[1].(string)

		// Only our own notifications: other apps' buttons come through too.
		n.mu.Lock()
		ours := id == n.lastID
		n.mu.Unlock()
		if !ours {
			continue
		}
		select {
		case n.actions <- Action{ID: id, Key: key}:
		default: // nobody listening; drop it
		}
	}
}
