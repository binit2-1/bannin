import QtQuick
import org.kde.kirigami as Kirigami

// A small line chart of RAM % over the last minute (0-100 on the y axis).
Canvas {
    id: spark
    property var values: []
    property color lineColor: Kirigami.Theme.highlightColor

    implicitHeight: Kirigami.Units.gridUnit * 2.5
    onValuesChanged: requestPaint()
    onLineColorChanged: requestPaint()
    onWidthChanged: requestPaint()

    onPaint: {
        const ctx = getContext("2d")
        ctx.reset()
        const n = values.length
        if (n < 2) return

        // Faint guide line at 90 %, where things get dangerous.
        ctx.strokeStyle = Qt.alpha(Kirigami.Theme.textColor, 0.2)
        ctx.lineWidth = 1
        ctx.setLineDash([3, 3])
        const y90 = height * 0.1
        ctx.beginPath(); ctx.moveTo(0, y90); ctx.lineTo(width, y90); ctx.stroke()
        ctx.setLineDash([])

        const x = i => i * (width - 1) / (n - 1)
        const y = v => height - 1 - Math.max(0, Math.min(100, v)) / 100 * (height - 2)

        ctx.beginPath()
        ctx.moveTo(x(0), y(values[0]))
        for (let i = 1; i < n; i++) ctx.lineTo(x(i), y(values[i]))

        ctx.strokeStyle = lineColor
        ctx.lineWidth = 1.5
        ctx.stroke()

        // Soft fill under the line.
        ctx.lineTo(x(n - 1), height)
        ctx.lineTo(x(0), height)
        ctx.closePath()
        ctx.fillStyle = Qt.alpha(lineColor, 0.15)
        ctx.fill()
    }
}
