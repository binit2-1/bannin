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

    readonly property real memPct: status.mem_pct || 0
    readonly property real swapPct: status.swap_pct || 0
    readonly property bool snoozed: (status.snoozed_until || 0) * 1000 > Date.now()

    signal stopRequested(string id)
    signal snoozeRequested(int minutes)
    signal startServiceRequested()

    LevelColors { id: colors }

    // Plasma sizes the popup from these; the user can still drag it bigger.
    Layout.minimumWidth: Kirigami.Units.gridUnit * 20
    Layout.preferredWidth: Kirigami.Units.gridUnit * 24
    Layout.minimumHeight: Kirigami.Units.gridUnit * 18
    Layout.preferredHeight: Kirigami.Units.gridUnit * 30
    spacing: Kirigami.Units.smallSpacing

    // "bundled:react" is an SVG shipped with the widget; anything else is a theme icon.
    function iconFor(name) {
        if (!name) return "application-x-executable"
        return name.startsWith("bundled:")
            ? Qt.resolvedUrl("../icons/" + name.slice(8) + ".svg")
            : name
    }

    // 1536 MB becomes "1.5 GB"
    function fmt(mb) {
        return mb >= 1024 ? (mb / 1024).toFixed(1) + " GB" : Math.round(mb) + " MB"
    }

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

    Kirigami.Heading {
        text: "What's using memory"
        level: 4
    }

    PC.Label {
        visible: groups.length === 0
        text: "Nothing above 100 MB"
        opacity: 0.7
    }

    // Only the process list scrolls; the RAM and swap bars stay in view.
    PC.ScrollView {
        id: scroll
        Layout.fillWidth: true
        Layout.fillHeight: true
        Layout.minimumHeight: Kirigami.Units.gridUnit * 6
        visible: groups.length > 0
        PC.ScrollBar.horizontal.policy: PC.ScrollBar.AlwaysOff

        ColumnLayout {
            id: list
            width: scroll.availableWidth
            spacing: Kirigami.Units.smallSpacing

            Repeater {
                model: full.groups

                delegate: RowLayout {
                    Layout.fillWidth: true
                    spacing: Kirigami.Units.smallSpacing

                    // App icon, with a small warning badge for detached dev tools
                    Kirigami.Icon {
                        Layout.preferredWidth: Kirigami.Units.iconSizes.smallMedium
                        Layout.preferredHeight: Kirigami.Units.iconSizes.smallMedium
                        source: full.iconFor(modelData.icon)
                        fallback: "application-x-executable"

                        Kirigami.Icon {
                            anchors.right: parent.right
                            anchors.bottom: parent.bottom
                            width: parent.width / 2
                            height: width
                            source: "dialog-warning"
                            visible: modelData.detached
                        }
                    }

                    PC.Label {
                        Layout.fillWidth: true
                        text: modelData.name
                        elide: Text.ElideRight
                    }

                    PC.Label {
                        text: full.fmt(modelData.rss_mb)
                        opacity: 0.8
                    }

                    PC.ToolButton {
                        icon.name: "process-stop"
                        display: PC.AbstractButton.IconOnly
                        text: "Stop " + modelData.name
                        PC.ToolTip.text: text
                        PC.ToolTip.visible: hovered
                        onClicked: full.stopRequested(modelData.id)
                    }
                }
            }
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