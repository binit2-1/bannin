import QtQuick
import QtQuick.Layouts
import org.kde.plasma.components as PC
import org.kde.kirigami as Kirigami

// The process list. Rows are kept in a ListModel and patched in place by
// group ID, so refreshing every 2 s doesn't rebuild them: no flicker, and
// hover, focus and expanded rows survive.
PC.ScrollView {
    id: scroll

    property var groups: []
    property real minMb: 100
    signal stopRequested(string id)

    // IDs of rows the user expanded.
    property var expanded: ({})
    function toggle(gid) {
        const e = Object.assign({}, expanded)
        if (e[gid]) delete e[gid]; else e[gid] = true
        expanded = e
    }

    property real biggest: 1 // MB of the biggest row, for the size bars

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

    ListModel { id: model }

    onGroupsChanged: sync()
    onMinMbChanged: sync()

    // Make the model match `groups` with the fewest changes: remove rows
    // that are gone, move rows that changed place, update the rest.
    function sync() {
        const want = (groups || [])
            .filter(g => g.rss_mb >= minMb || g.detached)
            // tools left running first, then biggest first
            .sort((a, b) => (b.detached - a.detached) || (b.rss_mb - a.rss_mb))

        const ids = {}
        let top = 1
        for (const g of want) { ids[g.id] = true; top = Math.max(top, g.rss_mb) }
        biggest = top

        for (let i = model.count - 1; i >= 0; i--) {
            if (!ids[model.get(i).gid]) model.remove(i)
        }
        for (let i = 0; i < want.length; i++) {
            const g = want[i]
            const row = {
                gid: g.id, name: g.name, kind: g.kind, rssMb: g.rss_mb,
                detached: g.detached, icon: g.icon || "", cmd: g.cmd || "",
                pids: (g.pids || []).join(" "), nprocs: (g.pids || []).length,
                section: g.detached ? "Needs attention" : "Using the most",
            }
            let j = -1
            for (let k = i; k < model.count; k++) {
                if (model.get(k).gid === g.id) { j = k; break }
            }
            if (j < 0) {
                model.insert(i, row)
            } else {
                if (j !== i) model.move(j, i, 1)
                model.set(i, row)
            }
        }
    }

    function stopAllDetached() {
        for (let i = 0; i < model.count; i++) {
            const r = model.get(i)
            if (r.detached) stopRequested(r.gid)
        }
    }

    PC.ScrollBar.horizontal.policy: PC.ScrollBar.AlwaysOff

    ListView {
        id: view
        model: model
        clip: true
        spacing: 2
        focus: true
        activeFocusOnTab: true
        keyNavigationEnabled: true
        currentIndex: -1
        highlightFollowsCurrentItem: false

        // Enter/Space expands the selected row; Delete stops it.
        Keys.onReturnPressed: if (currentItem) scroll.toggle(currentItem.gid)
        Keys.onSpacePressed: if (currentItem) scroll.toggle(currentItem.gid)
        Keys.onDeletePressed: if (currentItem) scroll.stopRequested(currentItem.gid)

        section.property: "section"
        section.delegate: RowLayout {
            required property string section
            width: ListView.view.width
            PC.Label {
                Layout.fillWidth: true
                Layout.topMargin: Kirigami.Units.smallSpacing
                text: section.toUpperCase()
                font: Kirigami.Theme.smallFont
                opacity: 0.7
            }
            PC.ToolButton {
                visible: section === "Needs attention"
                text: "Stop all"
                icon.name: "process-stop"
                onClicked: scroll.stopAllDetached()
            }
        }

        // Rows fade in and out; the others slide into place.
        add: Transition { NumberAnimation { property: "opacity"; from: 0; to: 1; duration: Kirigami.Units.longDuration } }
        remove: Transition { NumberAnimation { property: "opacity"; to: 0; duration: Kirigami.Units.longDuration } }
        displaced: Transition { NumberAnimation { property: "y"; duration: Kirigami.Units.longDuration; easing.type: Easing.OutQuad } }

        Kirigami.PlaceholderMessage {
            anchors.centerIn: parent
            width: parent.width - Kirigami.Units.gridUnit * 2
            visible: view.count === 0
            icon.name: "checkmark"
            text: "All quiet"
            explanation: "No app is using more than " + scroll.fmt(scroll.minMb) + "."
        }

        delegate: Item {
            id: row
            required property int index
            required property string gid
            required property string name
            required property string kind
            required property real rssMb
            required property bool detached
            required property string icon
            required property string cmd
            required property string pids
            required property int nprocs

            readonly property bool open: scroll.expanded[gid] === true
            readonly property bool current: ListView.isCurrentItem && view.activeFocus

            width: ListView.view.width
            implicitHeight: col.implicitHeight + Kirigami.Units.smallSpacing * 2

            Rectangle {
                anchors.fill: parent
                radius: Kirigami.Units.cornerRadius
                color: Kirigami.Theme.highlightColor
                opacity: row.current ? 0.25 : hover.hovered ? 0.12 : 0
                Behavior on opacity { NumberAnimation { duration: Kirigami.Units.shortDuration } }
            }
            HoverHandler { id: hover }
            MouseArea {
                anchors.fill: parent
                onClicked: { view.currentIndex = row.index; scroll.toggle(row.gid) }
            }

            ColumnLayout {
                id: col
                anchors.left: parent.left
                anchors.right: parent.right
                anchors.top: parent.top
                anchors.margins: Kirigami.Units.smallSpacing
                spacing: 2

                RowLayout {
                    Layout.fillWidth: true
                    spacing: Kirigami.Units.smallSpacing

                    // App icon, with a small warning badge for detached dev tools
                    Kirigami.Icon {
                        Layout.preferredWidth: Kirigami.Units.iconSizes.smallMedium
                        Layout.preferredHeight: Kirigami.Units.iconSizes.smallMedium
                        source: scroll.iconFor(row.icon)
                        fallback: "application-x-executable"

                        Kirigami.Icon {
                            anchors.right: parent.right
                            anchors.bottom: parent.bottom
                            width: parent.width / 2
                            height: width
                            source: "dialog-warning"
                            visible: row.detached
                        }
                    }

                    ColumnLayout {
                        Layout.fillWidth: true
                        spacing: 2
                        PC.Label {
                            Layout.fillWidth: true
                            text: row.name
                            elide: Text.ElideRight
                        }
                        // Size compared to the biggest app in the list.
                        Rectangle {
                            Layout.fillWidth: true
                            implicitHeight: 3
                            radius: 1.5
                            color: Qt.alpha(Kirigami.Theme.textColor, 0.1)
                            Rectangle {
                                height: parent.height
                                radius: parent.radius
                                width: parent.width * Math.min(1, row.rssMb / scroll.biggest)
                                color: row.detached ? Kirigami.Theme.neutralTextColor : Kirigami.Theme.highlightColor
                                opacity: 0.7
                                Behavior on width { NumberAnimation { duration: 300 } }
                            }
                        }
                    }

                    PC.Label {
                        text: scroll.fmt(row.rssMb)
                        opacity: 0.8
                        font.features: { "tnum": 1 }
                    }

                    PC.ToolButton {
                        icon.name: "process-stop"
                        display: PC.AbstractButton.IconOnly
                        text: "Stop " + row.name
                        // Quiet until needed: full strength on hover, focus, or a warning.
                        opacity: hover.hovered || row.current || row.detached || activeFocus ? 1 : 0.35
                        PC.ToolTip.text: text
                        PC.ToolTip.visible: hovered
                        onClicked: scroll.stopRequested(row.gid)
                    }

                    Kirigami.Icon {
                        Layout.preferredWidth: Kirigami.Units.iconSizes.small
                        Layout.preferredHeight: Kirigami.Units.iconSizes.small
                        source: row.open ? "arrow-up" : "arrow-down"
                        opacity: 0.6
                    }
                }

                // Details, shown when the row is expanded.
                ColumnLayout {
                    Layout.fillWidth: true
                    Layout.leftMargin: Kirigami.Units.iconSizes.smallMedium + Kirigami.Units.smallSpacing
                    visible: row.open
                    spacing: 2
                    PC.Label {
                        Layout.fillWidth: true
                        wrapMode: Text.Wrap
                        font: Kirigami.Theme.smallFont
                        text: (row.nprocs === 1 ? "1 process" : row.nprocs + " processes")
                              + (row.detached ? " · left running after its terminal closed" : "")
                              + "\nPID " + row.pids
                    }
                    PC.Label {
                        Layout.fillWidth: true
                        visible: row.cmd !== ""
                        wrapMode: Text.WrapAnywhere
                        maximumLineCount: 4
                        elide: Text.ElideRight
                        textFormat: Text.PlainText
                        font.family: "monospace"
                        font.pointSize: Kirigami.Theme.smallFont.pointSize
                        opacity: 0.7
                        text: row.cmd
                    }
                }
            }
        }
    }
}
