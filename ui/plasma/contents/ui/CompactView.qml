import QtQuick
import QtQuick.Layouts
import org.kde.plasma.components as PC

Item {
    id: compact
    property real memPct: 0
    property real swapPct: 0
    signal clicked()

    Layout.minimumWidth: 152
    Layout.preferredWidth: 152

    // one half of the widget: a bar with its name and percentage
    component Meter: Rectangle {
        id: meter
        property string name
        property real pct: 0
        property color baseColor

        width: 70
        height: 18
        radius: 5
        color: "#3a3a3a"
        clip: true

        Rectangle {
            height: parent.height
            radius: parent.radius
            width: parent.width * Math.min(meter.pct, 100) / 100
            color: meter.pct > 85 ? "#c0392b"
                 : meter.pct > 75 ? "#d68910"
                 : meter.baseColor
            Behavior on width { NumberAnimation { duration: 300 } }
        }

        PC.Label {
            anchors.centerIn: parent
            text: meter.name + " " + Math.round(meter.pct) + "%"
            color: "white"
            font.pixelSize: 11
        }
    }

    Row {
        anchors.centerIn: parent
        spacing: 2

        Meter {
            name: "RAM"
            pct: compact.memPct
            baseColor: "#27ae60"
        }

        Meter {
            name: "Swap"
            pct: compact.swapPct
            baseColor: "#3498db"
        }
    }

    MouseArea {
        anchors.fill: parent
        onClicked: compact.clicked()
    }
}
