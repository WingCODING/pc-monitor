package pcmonitor.viewmodel

import pcmonitor.model.DashboardMetrics
import pcmonitor.model.ProcessMetrics
import pcmonitor.model.SystemMetrics

/**
 * Ligação com o agente, do ponto de vista da tela.
 *
 * São quatro estados porque a reação do usuário é diferente em cada um:
 * esperar (Loading), nada a fazer (Connected), subir o agente (Disconnected) e
 * investigar o que ele respondeu (Error).
 */
enum class ConnectionState {
    Loading,
    Connected,
    Disconnected,
    Error,
}

/**
 * Estado completo do dashboard.
 *
 * As métricas ficam no estado mesmo quando a ligação cai: apagar a tela a cada
 * oscilação de conexão faria o dashboard piscar e esconderia justamente a
 * última leitura boa, que é o que o usuário quer ver enquanto o agente volta.
 */
data class DashboardUiState(
    val connection: ConnectionState = ConnectionState.Loading,
    val metrics: DashboardMetrics? = null,
    val history: MetricsHistory = MetricsHistory(),
    val processes: List<ProcessMetrics> = emptyList(),
    val system: SystemMetrics? = null,
    val message: String? = null,
)
