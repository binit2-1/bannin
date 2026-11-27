import QtQuick
import QtQuick.Layouts
import org.kde.plasma.components as PC
import org.kde.kirigami as Kirigami

ColumnLayout {
    id: full

    property real memPct: 0
    property real swapPct: 0
    property real usedGb: 0
    property real totalGb: 0
    property var groups: []
    property string error: ""

    signal stopRequested(string id)

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

    Kirigami.Heading {
        text: "Bannin"
        level: 3
    }

    PC.Label {
        text: "RAM  " + usedGb.toFixed(1) + " / " + totalGb.toFixed(1)
              + " GB  (" + memPct.toFixed(1) + "%)"
    }
    PC.ProgressBar {
        Layout.fillWidth: true
        from: 0; to: 100
        value: memPct
    }

    PC.Label { text: "Swap " + swapPct.toFixed(1) + "%" }
    PC.ProgressBar {
        Layout.fillWidth: true
        from: 0; to: 100
        value: swapPct
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