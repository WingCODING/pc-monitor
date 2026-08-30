package pcmonitor.ui

import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Modifier
import pcmonitor.network.AgentClient
import pcmonitor.network.DEFAULT_AGENT_URL
import pcmonitor.repository.MetricsRepository
import pcmonitor.repository.ProcessesRepository
import pcmonitor.ui.theme.PcMonitorTheme
import pcmonitor.viewmodel.MetricsViewModel

/**
 * Raiz da aplicação: monta a cadeia cliente → repositories → ViewModel e
 * entrega o estado à tela.
 *
 * A montagem acontece aqui, e não dentro do ViewModel, para que os testes
 * possam construir o ViewModel com repositories falsos sem nenhum HTTP por
 * perto.
 */
@Composable
fun App(agentUrl: String = DEFAULT_AGENT_URL, modifier: Modifier = Modifier) {
    val scope = rememberCoroutineScope()

    val client = remember(agentUrl) { AgentClient(agentUrl) }

    val viewModel = remember(client, scope) {
        MetricsViewModel(
            metricsRepository = MetricsRepository(client),
            processesRepository = ProcessesRepository(client),
            scope = scope,
        )
    }

    DisposableEffect(viewModel) {
        viewModel.start()

        onDispose {
            viewModel.stop()
            client.close()
        }
    }

    val state by viewModel.state.collectAsState()

    PcMonitorTheme {
        DashboardScreen(
            metrics = state.metrics,
            processes = state.processes,
            system = state.system,
            connection = state.connection,
            modifier = modifier,
        )
    }
}
