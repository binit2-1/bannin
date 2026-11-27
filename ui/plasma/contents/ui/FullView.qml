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

    Layout.minimumWidth: Kirigami.Units.gridUnit * 20
    spacing: Kirigami.Units.smallSpacing

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

    Repeater {
        model: full.groups

        delegate: RowLayout {
            Layout.fillWidth: true
            spacing: Kirigami.Units.smallSpacing

            // A fixed-size slot, so names line up even when there is no warning icon
            Item {
                Layout.preferredWidth: Kirigami.Units.iconSizes.small
                Layout.preferredHeight: Kirigami.Units.iconSizes.small

                Kirigami.Icon {
                    anchors.fill: parent
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

    PC.Label {
        visible: error !== ""
        text: error
        color: Kirigami.Theme.negativeTextColor
        wrapMode: Text.Wrap
        Layout.fillWidth: true
    }
}