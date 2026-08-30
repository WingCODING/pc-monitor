package pcmonitor.repository

import kotlinx.coroutines.flow.Flow
import pcmonitor.model.DashboardMetrics
import pcmonitor.model.ProcessMetrics
import pcmonitor.model.SystemMetrics
import pcmonitor.viewmodel.ProcessSort

/**
 * Contratos de leitura consumidos pelo ViewModel.
 *
 * O ViewModel depende destas interfaces, e não das implementações sobre Ktor.
 * É o que permite testar as transições de estado — inclusive queda e volta da
 * conexão — sem servidor, sem rede e sem tempo real.
 */
interface MetricsSource {
    suspend fun getMetrics(): AgentResult<DashboardMetrics>

    suspend fun getSystem(): AgentResult<SystemMetrics>

    /** Fluxo contínuo de snapshots; reconecta sozinho e não termina. */
    fun observeMetrics(): Flow<AgentResult<DashboardMetrics>>
}

interface ProcessesSource {
    suspend fun getProcesses(
        sort: ProcessSort = ProcessSort.Cpu,
        limit: Int = ProcessesRepository.DEFAULT_LIMIT,
    ): AgentResult<List<ProcessMetrics>>
}
