import QtQuick
import QtQuick.Layouts
import org.kde.plasma.components as PC

Item {
    id: compact
    property real memPct: 0
    signal clicked()

    Layout.minimumWidth: 110
    Layout.preferredWidth: 110

    Rectangle {
        id: track
        anchors.centerIn: parent
        width: 100
        height: 18
        radius: 5
        color: "#3a3a3a"

        Rectangle {
            height: parent.height
            radius: parent.radius
            width: parent.width * Math.min(compact.memPct, 100) / 100
            color: compact.memPct > 85 ? "#c0392b"
                 : compact.memPct > 75 ? "#d68910"
                 : "#27ae60"
            Behavior on width { NumberAnimation { duration: 300 } }
        }

        PC.Label {
            anchors.centerIn: parent
            text: "RAM " + Math.round(compact.memPct) + "%"
            color: "white"
            font.pixelSize: 11
        }
    }

    MouseArea {
        anchors.fill: parent
        onClicked: compact.clicked()
    }
}