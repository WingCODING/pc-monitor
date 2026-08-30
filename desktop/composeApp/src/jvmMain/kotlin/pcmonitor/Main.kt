package pcmonitor

import androidx.compose.ui.unit.DpSize
import androidx.compose.ui.unit.dp
import androidx.compose.ui.window.Window
import androidx.compose.ui.window.application
import androidx.compose.ui.window.rememberWindowState
import pcmonitor.network.DEFAULT_AGENT_URL
import pcmonitor.ui.App

fun main() = application {
    // O agente pode ter sido subido em outra porta (PCMON_PORT); sem esta
    // saída o desktop ficaria preso no 8080 e mostraria "desconectado" para um
    // agente que está de pé.
    val agentUrl = System.getenv("PCMON_AGENT_URL") ?: DEFAULT_AGENT_URL

    // Tamanho inicial escolhido para caber as quatro métricas numa linha só; a
    // janela continua redimensionável e a grade se reorganiza abaixo disso.
    val windowState = rememberWindowState(size = DpSize(1180.dp, 800.dp))

    Window(
        onCloseRequest = ::exitApplication,
        state = windowState,
        title = "PC Monitor",
    ) {
        App(agentUrl)
    }
}
