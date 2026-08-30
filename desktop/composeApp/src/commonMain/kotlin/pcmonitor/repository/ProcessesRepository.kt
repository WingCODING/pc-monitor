package pcmonitor.repository

import pcmonitor.model.ProcessMetrics
import pcmonitor.network.AgentClient
import pcmonitor.viewmodel.ProcessSort

/**
 * Lista de processos.
 *
 * Separada do `MetricsRepository` porque as duas têm cadências diferentes: o
 * snapshot é barato e chega a cada segundo, enquanto a varredura de processos
 * custa uma volta inteira em /proc no agente.
 */
class ProcessesRepository(private val client: AgentClient) {
    suspend fun getProcesses(
        sort: ProcessSort = ProcessSort.Cpu,
        limit: Int = DEFAULT_LIMIT,
    ): AgentResult<List<ProcessMetrics>> = callAgent { client.processes(sort = sort.apiValue, limit = limit) }

    companion object {
        const val DEFAULT_LIMIT = 50
    }
}
