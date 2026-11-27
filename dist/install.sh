#!/bin/sh
# Installs or upgrades Bannin for the current user from a release tarball.
# Needs KDE Plasma 6 and systemd. No root, Go, git or make.
set -eu

id_=io.github.binit2-1.bannin
here=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
config=${XDG_CONFIG_HOME:-$HOME/.config}
data=${XDG_DATA_HOME:-$HOME/.local/share}
bindir=$HOME/.local/bin
unitdir=$config/systemd/user

fail() { echo "error: $*" >&2; exit 1; }
warn() { echo "warning: $*" >&2; }

# Check everything before changing anything.
[ "$(id -u)" -ne 0 ] || fail "run this as your normal user, without sudo"
[ -x "$here/bin/bannind" ] || fail "no binaries here: run install.sh from an unpacked release tarball"
"$here/bin/bannin" version >/dev/null 2>&1 ||
	fail "these binaries don't run on this system ($(uname -s) $(uname -m)): download the tarball for your CPU"
command -v systemctl >/dev/null 2>&1 || fail "systemd is required (systemctl not found)"
systemctl --user show-environment >/dev/null 2>&1 ||
	fail "can't reach your systemd user session: run this in a terminal inside Plasma"
command -v kpackagetool6 >/dev/null 2>&1 || fail "KDE Plasma 6 is required (kpackagetool6 not found)"

# The widget runs `bannin` through Plasma's "executable" data engine.
found=
for dir in /usr/lib64/qt6/qml /usr/lib/qt6/qml /usr/lib/*/qt6/qml /run/current-system/sw/lib/qt-6/qml; do
	if [ -d "$dir/org/kde/plasma/plasma5support" ]; then found=1; fi
done
[ -n "$found" ] ||
	warn "Plasma's plasma5support module wasn't found. If the widget fails to load, install your distro's plasma5support package."

echo "Installing the programs to $bindir"
mkdir -p "$bindir" "$unitdir"
install -m 755 "$here/bin/bannin" "$here/bin/bannind" "$bindir/"

echo "Installing the panel widget"
upgrade=
if [ -d "$data/plasma/plasmoids/$id_" ]; then
	upgrade=1
	kpackagetool6 -t Plasma/Applet -u "$here/plasmoid" || fail "couldn't upgrade the widget"
else
	kpackagetool6 -t Plasma/Applet -i "$here/plasmoid" || fail "couldn't install the widget"
fi

echo "Starting the background service"
install -m 644 "$here/bannind.service" "$unitdir/bannind.service"
systemctl --user daemon-reload
systemctl --user enable bannind.service
systemctl --user restart bannind.service
sleep 2
systemctl --user is-active --quiet bannind.service ||
	fail "the service didn't stay running: see 'journalctl --user -u bannind'"

# Without systemd startup, Plasma never reaches graphical-session.target,
# so start the service from Plasma's autostart as well.
if command -v kreadconfig6 >/dev/null 2>&1 &&
	[ "$(kreadconfig6 --file startkderc --group General --key systemdBoot)" = false ]; then
	mkdir -p "$config/autostart"
	cat >"$config/autostart/bannind.desktop" <<DESKTOP
[Desktop Entry]
Type=Application
Name=Bannin service
Exec=systemctl --user start bannind.service
OnlyShowIn=KDE;
NoDisplay=true
DESKTOP
fi

echo
echo "Bannin is installed."
if [ -n "$upgrade" ]; then
	echo "Restart Plasma to load the new widget:  systemctl --user restart plasma-plasmashell"
else
	echo "Add it to your panel: right-click the panel, 'Add Widgets…', then Bannin."
fi
echo "Auto-stop is in dry run: it only logs what it would stop."
case ":$PATH:" in
*":$bindir:"*) ;;
*) echo "To use the 'bannin' command, add $bindir to your PATH." ;;
esac
