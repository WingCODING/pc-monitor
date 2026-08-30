package pcmonitor.repository

import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow
import pcmonitor.model.DashboardMetrics
import pcmonitor.model.SystemMetrics
import pcmonitor.network.AgentClient
import kotlin.time.Duration
import kotlin.time.Duration.Companion.seconds

/**
 * Fronteira entre o dashboard e o agente.
 *
 * O ViewModel depende desta classe e não do `AgentClient`: é o que mantém Ktor
 * fora das camadas de cima e o que permite testar o ViewModel sem servidor.
 */
class MetricsRepository(
    private val client: AgentClient,
    private val firstRetryDelay: Duration = 1.seconds,
    private val maxRetryDelay: Duration = 10.seconds,
) {
    suspend fun getMetrics(): AgentResult<DashboardMetrics> = callAgent { client.metrics() }

    /**
     * Informações da máquina, lidas uma vez por sessão.
     *
     * Hostname, sistema e arquitetura não mudam enquanto o agente roda, então
     * não fazem parte do snapshot que trafega a cada segundo.
     */
    suspend fun getSystem(): AgentResult<SystemMetrics> = callAgent { client.system() }

    /**
     * Fluxo contínuo de snapshots, com reconexão.
     *
     * O laço só recomeça depois que a conexão anterior terminou, e o fluxo é
     * coletado num único ponto — é o que garante que uma reconexão não deixe
     * duas sessões WebSocket vivas ao mesmo tempo.
     *
     * A espera cresce de 1 s até 10 s e volta ao início a cada conexão
     * bem-sucedida: reconectar num laço apertado enquanto o agente está fora do
     * ar só gastaria CPU dos dois lados, e esperar sempre 10 s faria o
     * dashboard demorar a voltar depois de um reinício rápido do agente.
     */
    fun observeMetrics(): Flow<AgentResult<DashboardMetrics>> = flow {
        var retryDelay = firstRetryDelay

        while (true) {
            try {
                client.snapshots().collect { snapshot ->
                    retryDelay = firstRetryDelay
                    emit(AgentResult.Success(snapshot))
                }

                // O fluxo terminar sem exceção significa fechamento limpo do
                // outro lado: o agente saiu.
                emit(AgentResult.Failure(FailureReason.Unreachable, "O agente encerrou a conexão."))
            } catch (cancellation: CancellationException) {
                throw cancellation
            } catch (failure: Throwable) {
                emit(failureFor(failure))
            }

            delay(retryDelay)

            retryDelay = minOf(retryDelay * 2, maxRetryDelay)
        }
    }
}
