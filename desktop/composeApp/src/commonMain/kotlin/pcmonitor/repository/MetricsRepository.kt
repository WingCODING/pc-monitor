package pcmonitor.repository

import pcmonitor.model.DashboardMetrics
import pcmonitor.model.SystemMetrics
import pcmonitor.network.AgentClient

/**
 * Fronteira entre o dashboard e o agente.
 *
 * O ViewModel depende desta classe e não do `AgentClient`: é o que mantém Ktor
 * fora das camadas de cima e o que permite testar o ViewModel sem servidor.
 */
class MetricsRepository(private val client: AgentClient) {
    suspend fun getMetrics(): AgentResult<DashboardMetrics> = callAgent { client.metrics() }

    /**
     * Informações da máquina, lidas uma vez por sessão.
     *
     * Hostname, sistema e arquitetura não mudam enquanto o agente roda, então
     * não fazem parte do snapshot que trafega a cada segundo.
     */
    suspend fun getSystem(): AgentResult<SystemMetrics> = callAgent { client.system() }
}
