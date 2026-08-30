package pcmonitor.ui

import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import pcmonitor.ui.components.ConnectionState
import pcmonitor.ui.preview.sampleDashboardMetrics
import pcmonitor.ui.preview.sampleProcesses
import pcmonitor.ui.preview.sampleSystem
import pcmonitor.ui.theme.PcMonitorTheme

/**
 * Raiz da aplicação.
 *
 * Ainda alimentada por dados de exemplo: a integração com o agente entra na
 * Epic 14, e a tela foi montada antes para que o layout pudesse ser resolvido
 * sem depender do backend.
 */
@Composable
fun App(modifier: Modifier = Modifier) {
    PcMonitorTheme {
        DashboardScreen(
            metrics = sampleDashboardMetrics,
            processes = sampleProcesses,
            system = sampleSystem,
            connection = ConnectionState.Connected,
            modifier = modifier,
        )
    }
}
