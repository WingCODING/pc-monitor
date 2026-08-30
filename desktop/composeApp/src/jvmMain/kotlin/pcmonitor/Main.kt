package pcmonitor

import androidx.compose.ui.unit.DpSize
import androidx.compose.ui.unit.dp
import androidx.compose.ui.window.Window
import androidx.compose.ui.window.application
import androidx.compose.ui.window.rememberWindowState
import pcmonitor.ui.App
import javax.swing.JOptionPane

fun main() {
    val agent = try {
        AgentProcess.start()
    } catch (error: Exception) {
        JOptionPane.showMessageDialog(
            null,
            "Não foi possível iniciar o agente do PC Monitor.\n${error.message}",
            "PC Monitor",
            JOptionPane.ERROR_MESSAGE,
        )
        return
    }

    try {
        application {
            // Tamanho inicial escolhido para caber as quatro métricas numa linha só; a
            // janela continua redimensionável e a grade se reorganiza abaixo disso.
            val windowState = rememberWindowState(size = DpSize(1180.dp, 800.dp))

            Window(
                onCloseRequest = ::exitApplication,
                state = windowState,
                title = "PC Monitor",
            ) {
                App(agent.baseUrl)
            }
        }
    } finally {
        agent.close()
    }
}
