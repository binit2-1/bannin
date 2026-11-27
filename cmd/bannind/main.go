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

	"github.com/binit2-1/bannin/internal/config"
	"github.com/binit2-1/bannin/internal/ipc"
	"github.com/binit2-1/bannin/internal/service"
)

var version = "dev"

func main() {
	fake := flag.Bool("fake", false, "replay a made-up memory curve instead of reading /proc")
	interval := flag.Duration("interval", time.Second, "how often to sample memory")
	cfgPath := flag.String("config", config.Path(), "settings file")
	flag.Parse()

	// systemd's journal adds its own timestamps.
	if os.Getenv("JOURNAL_STREAM") != "" {
		log.SetFlags(0)
	}

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Printf("config %s: %v (using defaults)", *cfgPath, err)
	}

	// Stop cleanly on Ctrl+C or `systemctl stop`.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	svc := service.New(service.Options{Interval: *interval, Fake: *fake, Config: cfg})

	// Notifications are optional: without a notification server we still run.
	alerts, err := service.NewAlerts(*fake, svc.Actions, cfg.Notify)
	if err != nil {
		log.Printf("notifications off: %v", err)
		alerts = nil
	} else {
		defer alerts.Close()
		svc.OnLevel = alerts.OnLevel
	}

	// `systemctl --user reload bannind` sends SIGHUP: re-read the config.
	hup := make(chan os.Signal, 1)
	signal.Notify(hup, syscall.SIGHUP)
	go func() {
		for range hup {
			c, err := config.Load(*cfgPath)
			if err != nil {
				log.Printf("config %s: %v (keeping current settings)", *cfgPath, err)
				continue
			}
			svc.SetThresholds(c.Thresholds)
			svc.Auto.SetConfig(c.AutoStop)
			if alerts != nil {
				alerts.SetConfig(c.Notify)
			}
			log.Printf("config reloaded (auto-stop %v, dry run %v)", c.AutoStop.Enabled, c.AutoStop.DryRun)
		}
	}()

	// The socket the CLI and widget talk to.
	sockErr := make(chan error, 1)
	go func() {
		err := service.NewServer(svc, alerts).Listen(ctx, ipc.SocketPath())
		if err != nil {
			cancel()
		}
		sockErr <- err
	}()

	a := cfg.AutoStop
	log.Printf("bannind %s started (fake=%v, auto-stop=%v, dry-run=%v)", version, *fake, a.Enabled, a.DryRun)
	if err := svc.Run(ctx); err != nil {
		log.Fatal(err)
	}
	if err := <-sockErr; err != nil {
		log.Fatalf("socket: %v", err)
	}
	log.Print("bannind stopped")
}
