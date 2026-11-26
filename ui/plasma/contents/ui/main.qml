import QtQuick 
import org.kde.plasma.plasmoid
import org.kde.plasma.plasma5support as P5

PlasmoidItem{
    id: root

    property real memPct: 0
    property real swapPct: 0
    property real usedGb: 0
    property real totalGb: 0
    property string error: ""

    readonly property string cmd: "$HOME/.local/bin/bannin mem --json"

    P5.DataSource {
        id: exec
        engine: "executable"
        onNewData: (source, data) => {
            disconnectSource(source)   // run once; the Timer starts it again
            if (data["exit code"] !== 0) {
                root.error = data.stderr || "bannin failed"
                return
            }
            try {
                const s = JSON.parse(data.stdout)
                root.memPct = s.mem_pct
                root.swapPct = s.swap_pct
                root.usedGb = s.used_gb
                root.totalGb = s.total_gb
                root.error = ""
            } catch (e) {
                root.error = "Bad JSON: " + e
            }
        }
    }
    Timer {
        interval: 2000
        running: true
        repeat: true
        triggeredOnStart: true
        onTriggered: exec.connectSource(root.cmd)
    }

    compactRepresentation: CompactView {
        memPct: root.memPct
        onClicked: root.expanded = !root.expanded
    }

    fullRepresentation: FullView {
        memPct: root.memPct
        swapPct: root.swapPct
        usedGb: root.usedGb
        totalGb: root.totalGb
        error: root.error
    }
}