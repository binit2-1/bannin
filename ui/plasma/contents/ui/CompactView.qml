import QtQuick
import QtQuick.Layouts
import org.kde.plasma.components as PC
import org.kde.kirigami as Kirigami

Item {
    id: compact
    property real memPct: 0
    property real swapPct: 0
    property string level: "ok"     // "ok" | "elevated" | "high" | "critical" | "off"
    property int warn: 0            // detached dev tools
    property int ttfSec: 0          // seconds until RAM is full, 0 = not filling
    signal clicked()

    LevelColors { id: colors }

    Layout.minimumWidth: row.implicitWidth + Kirigami.Units.smallSpacing * 2
    Layout.preferredWidth: Layout.minimumWidth

    // One bar with its name and percentage.
    component Meter: Rectangle {
        id: meter
        property string name
        property real pct: 0
        property bool known: true
        property color fillColor

        implicitWidth: 74
        implicitHeight: 18
        radius: 5
        color: colors.track
        clip: true

        Rectangle {
            height: parent.height
            radius: parent.radius
            width: meter.known ? parent.width * Math.min(meter.pct, 100) / 100 : 0
            color: meter.fillColor
            Behavior on width { NumberAnimation { duration: 300 } }
            Behavior on color { ColorAnimation { duration: 400 } }
        }

        PC.Label {
            anchors.centerIn: parent
            text: meter.name + " " + (meter.known ? Math.round(meter.pct) + "%" : "--")
            font.pixelSize: 11
        }
    }

    RowLayout {
        id: row
        anchors.centerIn: parent
        spacing: 2

        Meter {
            name: "RAM"
            pct: compact.memPct
            known: compact.level !== "off"
            fillColor: colors.of(compact.level)
        }

        Meter {
            name: "Swap"
            pct: compact.swapPct
            known: compact.level !== "off"
            fillColor: compact.swapPct > 80 ? colors.high : Qt.alpha(colors.ok, 0.7)
        }

        // Never colour alone: a count of warnings, or how long until full.
        PC.Label {
            id: badge
            visible: text !== ""
            Layout.leftMargin: 2
            font.pixelSize: 11
            font.bold: compact.level === "critical"
            color: compact.level === "ok" ? Kirigami.Theme.textColor : colors.of(compact.level)
            text: compact.ttfSec > 0 && (compact.level === "critical" || compact.level === "high")
                  ? "~" + compact.ttfSec + "s"
                  : compact.warn > 0 ? "⚠" + compact.warn : ""
        }
    }

    // A single gentle pulse when the level changes; nothing flashes otherwise.
    onLevelChanged: if (level !== "ok" && level !== "off") pulse.restart()
    SequentialAnimation {
        id: pulse
        NumberAnimation { target: row; property: "opacity"; to: 0.35; duration: 180 }
        NumberAnimation { target: row; property: "opacity"; to: 1; duration: 320 }
    }

    MouseArea {
        anchors.fill: parent
        onClicked: compact.clicked()
    }
}
