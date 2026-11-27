import QtQuick
import org.kde.plasma.plasmoid
import org.kde.plasma.core as PlasmaCore
import org.kde.plasma.plasma5support as P5

PlasmoidItem {
    id: root

    // The last `bannin mem --json` reply, and `bannin list --json` groups.
    property var status: ({})
    property var groups: []
    property string error: ""
    // "off" when bannin itself cannot be run; otherwise the level it reports.
    readonly property string level: error !== "" && !status.level ? "off" : (status.level || "off")

    LevelColors { id: colors }

    toolTipMainText: level === "off"
        ? "Bannin · RAM --"
        : "RAM " + Math.round(status.mem_pct) + "% · "
          + status.used_gb.toFixed(1) + " / " + status.total_gb.toFixed(1) + " GB"
    toolTipSubText: {
        if (level === "off") return error || "Can't read memory"
        let parts = ["Swap " + status.swap_pct.toFixed(1) + "%", colors.label(level)]
        if (status.ttf_s) parts.push("full in ~" + status.ttf_s + "s")
        if (status.warn) parts.push(status.warn + " left running")
        if (status.biggest) parts.push("biggest: " + status.biggest)
        if (!status.service) parts.push("service off")
        return parts.join(" · ")
    }

    // A local build in ~/.local/bin wins over a packaged /usr/bin/bannin.
    readonly property string bannin: "PATH=\"$HOME/.local/bin:$PATH\" bannin"
    readonly property string cmdMem: bannin + " mem --json"
    // --all: the popup filters by the "Hide apps under" setting itself.
    readonly property string cmdList: bannin + " list --all --json"
    readonly property int pollMs: Math.max(1, Plasmoid.configuration.pollSeconds) * 1000

    // Commands whose output we don't parse: we only refresh afterwards.
    property var pending: ({})
    function run(cmd) {
        pending[cmd] = true
        exec.connectSource(cmd)
    }

    // Group IDs are "g" + hex, so they are safe to put in a shell command.
    function stopGroup(id) {
        if (/^g[0-9a-f]{8}$/.test(id)) run(bannin + " stop " + id)
    }
    function snooze(minutes) { run(bannin + " snooze " + Math.round(minutes)) }
    function startService() { run("systemctl --user start bannind.service") }
    // `bannin config` creates the file with defaults if needed and prints its path.
    function openSettings() { run("xdg-open \"$(" + bannin + " config)\"") }

    P5.DataSource {
        id: exec
        engine: "executable"
        onNewData: (source, data) => {
            disconnectSource(source)
            const ok = data["exit code"] === 0

            if (root.pending[source]) {
                delete root.pending[source]
                root.error = ok ? "" : (data.stderr || "command failed").trim()
                exec.connectSource(root.cmdMem) // refresh right away
                if (root.expanded) exec.connectSource(root.cmdList)
                return
            }
            if (!ok) {
                root.error = (data.stderr || "bannin failed").trim()
                if (source === root.cmdMem) root.status = ({})
                return
            }
            try {
                const s = JSON.parse(data.stdout)
                // `source` is the command that produced this output
                if (source === root.cmdMem) {
                    root.status = s
                } else if (source === root.cmdList) {
                    root.groups = s.groups
                }
                root.error = ""
            } catch (e) {
                root.error = "Bad JSON: " + e
            }
        }
    }

    // The bar: always on
    Timer {
        interval: root.pollMs
        running: true
        repeat: true
        triggeredOnStart: true
        onTriggered: exec.connectSource(root.cmdMem)
    }

    // The process list: only while the popup is open
    Timer {
        interval: root.pollMs
        running: root.expanded
        repeat: true
        triggeredOnStart: true
        onTriggered: exec.connectSource(root.cmdList)
    }

    compactRepresentation: CompactView {
        memPct: root.status.mem_pct || 0
        swapPct: root.status.swap_pct || 0
        level: root.level
        warn: root.status.warn || 0
        ttfSec: root.status.ttf_s || 0
        showSwap: Plasmoid.configuration.showSwap
        vertical: Plasmoid.formFactor === PlasmaCore.Types.Vertical
        onClicked: root.expanded = !root.expanded
    }

    fullRepresentation: FullView {
        status: root.status
        level: root.level
        groups: root.groups
        minMb: Plasmoid.configuration.minMb
        error: root.error
        onStopRequested: id => root.stopGroup(id)
        onSnoozeRequested: minutes => root.snooze(minutes)
        onStartServiceRequested: root.startService()
        onSettingsRequested: root.openSettings()
    }
}
