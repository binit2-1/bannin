#!/bin/sh
# Removes what install.sh added. Your settings are kept.
set -eu

id_=io.github.binit2-1.bannin
config=${XDG_CONFIG_HOME:-$HOME/.config}
data=${XDG_DATA_HOME:-$HOME/.local/share}

[ "$(id -u)" -ne 0 ] || { echo "error: run this as your normal user, without sudo" >&2; exit 1; }

systemctl --user disable --now bannind.service 2>/dev/null || true
rm -f "$config/systemd/user/bannind.service" "$config/autostart/bannind.desktop"
systemctl --user daemon-reload || true
rm -f "$HOME/.local/bin/bannin" "$HOME/.local/bin/bannind"
if [ -d "$data/plasma/plasmoids/$id_" ]; then
	kpackagetool6 -t Plasma/Applet -r "$id_"
fi

echo "Bannin is removed. Settings are still in $config/bannin (delete it if you like)."
echo "If the widget is still on your panel, remove it there."
