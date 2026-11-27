import QtQuick
import QtQuick.Controls as QQC2
import QtQuick.Layouts
import org.kde.kirigami as Kirigami
import org.kde.kcmutils as KCM

// Widget appearance. Warning levels, notifications and auto-stop belong to
// the service and live in ~/.config/bannin/config.toml.
KCM.SimpleKCM {
    property alias cfg_showSwap: showSwap.checked
    property alias cfg_minMb: minMb.value
    property alias cfg_pollSeconds: poll.value
    property bool cfg_showSwapDefault
    property int cfg_minMbDefault
    property int cfg_pollSecondsDefault

    Kirigami.FormLayout {
        QQC2.CheckBox {
            id: showSwap
            Kirigami.FormData.label: "Panel:"
            text: "Show swap bar"
        }
        QQC2.SpinBox {
            id: minMb
            Kirigami.FormData.label: "Hide apps under:"
            from: 0; to: 4096; stepSize: 50
            textFromValue: v => v + " MB"
            valueFromText: t => parseInt(t)
        }
        QQC2.SpinBox {
            id: poll
            Kirigami.FormData.label: "Update every:"
            from: 1; to: 30
            textFromValue: v => v + " s"
            valueFromText: t => parseInt(t)
        }

        Kirigami.Separator { Kirigami.FormData.isSection: true }

        QQC2.Label {
            Kirigami.FormData.label: "Warnings:"
            Layout.maximumWidth: Kirigami.Units.gridUnit * 22
            wrapMode: Text.Wrap
            text: "Levels, notifications and auto-stop are set in ~/.config/bannin/config.toml. "
                  + "Use the ⚙ button in the popup to open it, then run "
                  + "“systemctl --user reload bannind”."
        }
    }
}
