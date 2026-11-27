# Bannin

A memory watchman for KDE Plasma on Linux. Bannin shows RAM and swap in your
panel, warns you **before** the system runs out of memory and freezes, and can
stop dev tools you forgot about (Metro, Gradle daemons, the Android emulator).

<!-- Screenshots: panel bar in each level, the popup, a notification. -->
<!-- ![Panel](docs/panel.png) ![Popup](docs/popup.png) -->

## What it does

- **Panel bar**: shows RAM and swap, coloured by level (all good, elevated, high,
  critical) using your Plasma theme. It also shows a `⚠2` count of tools left
  running, or `~40s` when RAM will be full soon. Hover for details.
- **Popup**: RAM and swap, a 60-second sparkline, and the apps using the most
  memory, with icons. Tools whose terminal was closed get a Stop button. Click a
  row to see its PIDs and command line.
- **Notifications**: one when the level rises to High or Critical, with
  **Stop** and **Snooze 10 min** buttons. A cool-down keeps them from repeating.
- **Auto-stop** (dry run by default): if RAM stays critical, Bannin stops dev
  tools left running after their terminal closed. Until you turn dry run off,
  it only logs what it *would have* stopped. Browsers, editors and other apps are
  never auto-stopped.

Bannin reads `/proc/meminfo` and memory pressure (`/proc/pressure/memory`).
Levels go up straight away, but only come down once RAM is clearly lower, so
they don't flap. The time-to-full estimate is a straight-line fit over the last
30 seconds.

Keep `systemd-oomd` enabled: it is the safety net underneath Bannin.

## Install

### From source

Needs Go 1.26+, Plasma 6 and `kpackagetool6`.

```sh
make service-install   # builds, installs to ~/.local/bin, starts the user service
make widget-install    # installs the panel widget
```

Then right-click the panel → *Add Widgets…* → **Bannin**.

### RPM (Fedora)

```sh
make rpm               # build/rpm/RPMS/x86_64/bannin-*.rpm
sudo dnf install build/rpm/RPMS/x86_64/bannin-*.rpm
systemctl --user enable --now bannind
```

## Command line

```
bannin mem [--json]            RAM, swap and level
bannin list [--all] [--json]   apps grouped, biggest first (⚠ = left running)
bannin top                     biggest single processes
bannin stop <id> [--force]     stop a group (id from `list --json`)
bannin snooze [minutes]        mute warnings
bannin config                  print the settings path (creates it if missing)
```

The CLI asks the service over `$XDG_RUNTIME_DIR/bannin.sock`. If the service
isn't running, it reads `/proc` itself.

## Settings

`~/.config/bannin/config.toml`. Every key is optional. Open it with the ⚙
button in the popup.

```toml
[thresholds]
elevated = 75
high = 85
critical = 92
psi_high = 10        # memory pressure that alone counts as "high"
psi_critical = 25
hysteresis = 3

[notify]
enabled = true
cooldown_minutes = 2

[auto_stop]
enabled = true
dry_run = true       # turn off once the log looks right
kinds = ["metro", "gradle", "emulator"]
only_detached = true
after_seconds = 10
```

Apply changes with `systemctl --user reload bannind`. Check what auto-stop did,
or would have done:

```sh
journalctl --user -u bannind
```

The widget's own settings (swap bar, minimum app size, refresh rate) are under
*Configure Bannin…* on the widget.

## Trying it safely

```sh
systemctl --user stop bannind
make fake              # replays a made-up curve: 50 % → 97 % → back, every 2 min
```

In fake mode, the bar changes colour and notifications fire, but nothing is ever
stopped.

## Development

```sh
make test              # unit tests
make widget-popup      # popup in a window, QML errors in the terminal
make widget-test       # as a panel
make widget-refresh    # reinstall and restart plasmashell
```

## Not yet

macOS and desktops other than Plasma, CPU and disk-pressure warnings, and
auto-stopping apps that aren't dev tools.

## Credits

The React, Gradle, Android and Node.js icons are from
[Simple Icons](https://simpleicons.org) (CC0). See
`ui/plasma/contents/icons/LICENSE`.
