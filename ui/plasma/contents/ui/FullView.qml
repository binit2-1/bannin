import QtQuick
import QtQuick.Layouts
import org.kde.plasma.components as PC
import org.kde.kirigami as Kirigami

ColumnLayout {
    id: full

    property var status: ({})
    property string level: "off"
    property var groups: []
    property string error: ""
    property real minMb: 100

    readonly property real memPct: status.mem_pct || 0
    readonly property real swapPct: status.swap_pct || 0
    readonly property bool snoozed: (status.snoozed_until || 0) * 1000 > Date.now()

    signal stopRequested(string id)
    signal snoozeRequested(int minutes)
    signal startServiceRequested()
    signal settingsRequested()

    function clock(unix) {
        return Qt.formatTime(new Date(unix * 1000), "hh:mm")
    }

    LevelColors { id: colors }

    // Plasma sizes the popup from these; the user can still drag it bigger.
    Layout.minimumWidth: Kirigami.Units.gridUnit * 20
    Layout.preferredWidth: Kirigami.Units.gridUnit * 24
    Layout.minimumHeight: Kirigami.Units.gridUnit * 18
    Layout.preferredHeight: Kirigami.Units.gridUnit * 30
    spacing: Kirigami.Units.smallSpacing

    // Title and a level chip: colour plus a word, never colour alone.
    RowLayout {
        Layout.fillWidth: true
        Kirigami.Heading {
            text: "Bannin"
            level: 3
            Layout.fillWidth: true
        }
        Rectangle {
            radius: height / 2
            color: Qt.alpha(colors.of(full.level), 0.18)
            implicitHeight: chip.implicitHeight + Kirigami.Units.smallSpacing
            implicitWidth: chip.implicitWidth + Kirigami.Units.largeSpacing * 2
            Row {
                id: chip
                anchors.centerIn: parent
                spacing: Kirigami.Units.smallSpacing
                Rectangle {
                    width: 8; height: 8; radius: 4
                    anchors.verticalCenter: parent.verticalCenter
                    color: colors.of(full.level)
                }
                PC.Label { text: colors.label(full.level) }
            }
        }
    }

    Kirigami.InlineMessage {
        Layout.fillWidth: true
        visible: full.level !== "off" && !full.status.service
        type: Kirigami.MessageType.Warning
        text: "The Bannin service isn't running, so you won't be warned before a freeze."
        actions: [
            Kirigami.Action {
                text: "Start"
                icon.name: "media-playback-start"
                onTriggered: full.startServiceRequested()
            }
        ]
    }

    Kirigami.InlineMessage {
        Layout.fillWidth: true
        visible: !!full.status.fake
        type: Kirigami.MessageType.Information
        text: "Test mode: the service is replaying a made-up memory curve."
    }

    // One bar per resource: name and numbers above, bar below.
    component Bar: ColumnLayout {
        property string name
        property string detail
        property real pct
        property color barColor
        Layout.fillWidth: true
        spacing: 2
        RowLayout {
            Layout.fillWidth: true
            PC.Label { text: name; font.bold: true; Layout.fillWidth: true }
            PC.Label { text: detail; opacity: 0.8 }
        }
        Rectangle {
            Layout.fillWidth: true
            implicitHeight: 6
            radius: 3
            color: colors.track
            Rectangle {
                height: parent.height
                radius: parent.radius
                width: parent.width * Math.min(pct, 100) / 100
                color: barColor
                Behavior on width { NumberAnimation { duration: 300 } }
                Behavior on color { ColorAnimation { duration: 400 } }
            }
        }
    }

    Bar {
        name: "RAM"
        pct: full.memPct
        barColor: colors.of(full.level)
        detail: full.level === "off" ? "--"
              : (full.status.used_gb || 0).toFixed(1) + " / " + (full.status.total_gb || 0).toFixed(1)
                + " GB  (" + Math.round(full.memPct) + "%)"
    }

    Bar {
        name: "Swap"
        pct: full.swapPct
        barColor: full.swapPct > 80 ? colors.high : Qt.alpha(colors.ok, 0.7)
        detail: full.level === "off" ? "--" : full.swapPct.toFixed(1) + "%"
    }

    // The last minute, from the service's history.
    ColumnLayout {
        Layout.fillWidth: true
        visible: (full.status.history || []).length > 1
        spacing: 0
        RowLayout {
            Layout.fillWidth: true
            PC.Label { text: "Last 60 s"; opacity: 0.7; font: Kirigami.Theme.smallFont; Layout.fillWidth: true }
            PC.Label {
                visible: !!full.status.ttf_s
                text: "~" + full.status.ttf_s + " s until full"
                color: colors.of(full.level)
                font.bold: true
            }
        }
        Sparkline {
            Layout.fillWidth: true
            values: full.status.history || []
            lineColor: colors.of(full.level)
        }
    }

    Kirigami.Separator { Layout.fillWidth: true }

    GroupList {
        Layout.fillWidth: true
        Layout.fillHeight: true
        Layout.minimumHeight: Kirigami.Units.gridUnit * 6
        groups: full.groups
        minMb: full.minMb
        onStopRequested: id => full.stopRequested(id)
    }

    // What Bannin did lately (or, in dry run, would have done).
    ColumnLayout {
        Layout.fillWidth: true
        visible: (full.status.actions || []).length > 0
        spacing: 2

        Kirigami.Separator { Layout.fillWidth: true }
        PC.Label { text: "Recent"; font.bold: true; opacity: 0.8 }

        Repeater {
            model: full.status.actions || []
            delegate: RowLayout {
                Layout.fillWidth: true
                PC.Label {
                    text: full.clock(modelData.t)
                    opacity: 0.6
                    font.features: { "tnum": 1 }
                }
                PC.Label {
                    Layout.fillWidth: true
                    text: modelData.text
                    font.italic: !!modelData.dry_run
                    opacity: modelData.dry_run ? 0.75 : 1
                    elide: Text.ElideRight
                }
            }
        }
    }

    // Footer: dry-run note, snooze and settings.
    RowLayout {
        Layout.fillWidth: true
        spacing: Kirigami.Units.smallSpacing

        PC.Label {
            Layout.fillWidth: true
            elide: Text.ElideRight
            opacity: 0.75
            font: Kirigami.Theme.smallFont
            text: {
                if (!full.status.service) return ""
                if (!full.status.auto_stop) return "Auto-stop off"
                return full.status.dry_run ? "Dry run: auto-stop only logs" : "Auto-stop on"
            }
            PC.ToolTip.text: "Auto-stop acts only on dev tools (Metro, Gradle, emulator) left running "
                             + "after their terminal closed, when RAM stays critical. "
                             + "Change it in the settings file."
            PC.ToolTip.visible: footerHover.hovered
            HoverHandler { id: footerHover }
        }

        PC.ToolButton {
            visible: !!full.status.service
            icon.name: full.snoozed ? "notifications-disabled" : "notifications"
            text: full.snoozed ? "Snoozed till " + full.clock(full.status.snoozed_until) : "Snooze 10 min"
            onClicked: full.snoozeRequested(10)
            PC.ToolTip.text: full.snoozed ? "Click to snooze 10 more minutes" : "Mute warnings for 10 minutes"
            PC.ToolTip.visible: hovered
        }

        PC.ToolButton {
            icon.name: "configure"
            display: PC.AbstractButton.IconOnly
            text: "Open settings file"
            onClicked: full.settingsRequested()
            PC.ToolTip.text: text
            PC.ToolTip.visible: hovered
        }
    }

    PC.Label {
        visible: error !== ""
        text: error
        color: Kirigami.Theme.negativeTextColor
        wrapMode: Text.Wrap
        Layout.fillWidth: true
    }
}