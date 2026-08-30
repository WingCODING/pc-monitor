package pcmonitor.repository

import pcmonitor.model.ProcessMetrics
import pcmonitor.network.AgentClient

/**
 * Lista de processos.
 *
 * Separada do `MetricsRepository` porque as duas têm cadências diferentes: o
 * snapshot é barato e chega a cada segundo, enquanto a varredura de processos
 * custa uma volta inteira em /proc no agente.
 */
class ProcessesRepository(private val client: AgentClient) {
    suspend fun getProcesses(limit: Int = DEFAULT_LIMIT): AgentResult<List<ProcessMetrics>> =
        callAgent { client.processes(limit = limit) }

    companion object {
        const val DEFAULT_LIMIT = 50
    }
}
