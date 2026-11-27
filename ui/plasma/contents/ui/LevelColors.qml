import QtQuick
import org.kde.kirigami as Kirigami

// Colours for each warning level, taken from the Plasma theme so light,
// dark and accent-coloured themes all work.
QtObject {
    readonly property color ok: Kirigami.Theme.highlightColor
    readonly property color elevated: Kirigami.Theme.neutralTextColor
    readonly property color critical: Kirigami.Theme.negativeTextColor
    readonly property color high: Kirigami.ColorUtils.linearInterpolation(elevated, critical, 0.5)
    readonly property color off: Qt.alpha(Kirigami.Theme.textColor, 0.3)
    readonly property color track: Qt.alpha(Kirigami.Theme.textColor, 0.15)

    function of(level) {
        switch (level) {
        case "elevated": return elevated
        case "high":     return high
        case "critical": return critical
        case "off":      return off
        }
        return ok
    }

    function label(level) {
        switch (level) {
        case "elevated": return "Elevated"
        case "high":     return "High"
        case "critical": return "Critical"
        case "off":      return "Unknown"
        }
        return "All good"
    }
}
