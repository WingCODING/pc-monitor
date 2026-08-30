package pcmonitor

import androidx.compose.ui.window.Window
import androidx.compose.ui.window.application
import pcmonitor.ui.App

fun main() = application {
    Window(onCloseRequest = ::exitApplication, title = "PC Monitor") {
        App()
    }
}
