import QtQuick
import org.kde.plasma.plasmoid
import org.kde.plasma.plasma5support as P5

PlasmoidItem {
    id: root

    property real memPct: 0
    property real swapPct: 0
    property real usedGb: 0
    property real totalGb: 0
    property var groups: []
    property string error: ""

    readonly property string cmdMem: "$HOME/.local/bin/bannin mem --json"
    readonly property string cmdList: "$HOME/.local/bin/bannin list --json"
    readonly property string cmdStop: "$HOME/.local/bin/bannin stop "

    // Group IDs are "g" + hex, so they are safe to put in a shell command.
    function stopGroup(id) {
        if (/^g[0-9a-f]{8}$/.test(id)) {
            exec.connectSource(cmdStop + id)
        }
    }

    P5.DataSource {
        id: exec
        engine: "executable"
        onNewData: (source, data) => {
            disconnectSource(source)
            if (source.startsWith(root.cmdStop)) {
                root.error = data["exit code"] === 0 ? "" : (data.stderr || "stop failed")
                exec.connectSource(root.cmdList) // refresh right away
                return
            }
            if (data["exit code"] !== 0) {
                root.error = data.stderr || "bannin failed"
                return
            }
            try {
                const s = JSON.parse(data.stdout)
                // `source` is the command that produced this output
                if (source === root.cmdMem) {
                    root.memPct = s.mem_pct
                    root.swapPct = s.swap_pct
                    root.usedGb = s.used_gb
                    root.totalGb = s.total_gb
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
        interval: 2000
        running: true
        repeat: true
        triggeredOnStart: true
        onTriggered: exec.connectSource(root.cmdMem)
    }

    // The process list: only while the popup is open
    Timer {
        interval: 2000
        running: root.expanded
        repeat: true
        triggeredOnStart: true
        onTriggered: exec.connectSource(root.cmdList)
    }

    compactRepresentation: CompactView {
        memPct: root.memPct
        swapPct: root.swapPct
        onClicked: root.expanded = !root.expanded
    }

    fullRepresentation: FullView {
        memPct: root.memPct
        swapPct: root.swapPct
        usedGb: root.usedGb
        totalGb: root.totalGb
        groups: root.groups
        error: root.error
        onStopRequested: id => root.stopGroup(id)
    }
}