import QtQuick
import QtQuick.Layouts
import org.kde.plasma.components as PC
import org.kde.kirigami as Kirigami

ColumnLayout {
    property real memPct: 0
    property real swapPct: 0
    property real usedGb: 0
    property real totalGb: 0
    property string error: ""

    Layout.minimumWidth: Kirigami.Units.gridUnit * 16
    spacing: Kirigami.Units.smallSpacing

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

    PC.Label {
        visible: error !== ""
        text: error
        color: Kirigami.Theme.negativeTextColor
        wrapMode: Text.Wrap
        Layout.fillWidth: true
    }
}