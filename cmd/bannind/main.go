// bannind is the Bannin background service. It watches memory and warns
// before the system runs out.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/binit2-1/bannin/internal/engine"
	"github.com/binit2-1/bannin/internal/ipc"
	"github.com/binit2-1/bannin/internal/service"
)

var version = "dev"

func main() {
	fake := flag.Bool("fake", false, "replay a made-up memory curve instead of reading /proc")
	interval := flag.Duration("interval", time.Second, "how often to sample memory")
	flag.Parse()

	// systemd's journal adds its own timestamps.
	if os.Getenv("JOURNAL_STREAM") != "" {
		log.SetFlags(0)
	}

	// Stop cleanly on Ctrl+C or `systemctl stop`.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	svc := service.New(service.Options{
		Interval:   *interval,
		Fake:       *fake,
		Thresholds: engine.DefaultThresholds,
	})

	// Notifications are optional: without a notification server we still run.
	alerts, err := service.NewAlerts(*fake)
	if err != nil {
		log.Printf("notifications off: %v", err)
		alerts = nil
	} else {
		defer alerts.Close()
		svc.OnLevel = alerts.OnLevel
	}

	// The socket the CLI and widget talk to.
	sock := ipc.SocketPath()
	sockErr := make(chan error, 1)
	go func() {
		err := service.NewServer(svc, alerts).Listen(ctx, sock)
		if err != nil {
			cancel()
		}
		sockErr <- err
	}()

	log.Printf("bannind %s started (fake=%v)", version, *fake)
	if err := svc.Run(ctx); err != nil {
		log.Fatal(err)
	}
	if err := <-sockErr; err != nil {
		log.Fatalf("socket: %v", err)
	}
	log.Print("bannind stopped")
}
