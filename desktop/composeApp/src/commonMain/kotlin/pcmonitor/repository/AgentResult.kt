package pcmonitor.repository

import io.ktor.client.plugins.HttpRequestTimeoutException
import io.ktor.utils.io.errors.IOException
import kotlinx.coroutines.CancellationException
import kotlinx.serialization.SerializationException
import pcmonitor.network.AgentHttpException

/**
 * Resultado de uma leitura do agente.
 *
 * A camada de dados devolve isto em vez de lançar: a UI precisa reagir de
 * forma diferente a "agente fora do ar" e a "agente respondeu errado", e um
 * try/catch espalhado pelo ViewModel perderia essa distinção.
 */
sealed interface AgentResult<out T> {
    data class Success<T>(val value: T) : AgentResult<T>

    data class Failure(val reason: FailureReason, val message: String) : AgentResult<Nothing>
}

/**
 * Por que a leitura falhou.
 *
 * `Unreachable` é a única que significa "o agente não está lá" — as outras
 * significam que ele está, mas algo deu errado. A UI mostra Desconectado na
 * primeira e Erro nas demais, porque as ações do usuário são diferentes:
 * subir o agente contra investigar o que ele respondeu.
 */
enum class FailureReason {
    Unreachable,
    Unavailable,
    Malformed,
    Unexpected,
}

/**
 * Executa uma chamada ao agente convertendo as exceções conhecidas.
 *
 * `CancellationException` é relançada: engoli-la transformaria o cancelamento
 * do escopo — fechar a janela, trocar de tela — num erro exibido ao usuário, e
 * ainda quebraria a estrutura de corrotinas do chamador.
 */
internal suspend fun <T> callAgent(block: suspend () -> T): AgentResult<T> = try {
    AgentResult.Success(block())
} catch (cancellation: CancellationException) {
    throw cancellation
} catch (timeout: HttpRequestTimeoutException) {
    AgentResult.Failure(FailureReason.Unreachable, "O agente não respondeu a tempo.")
} catch (io: IOException) {
    AgentResult.Failure(FailureReason.Unreachable, "Não foi possível falar com o agente.")
} catch (http: AgentHttpException) {
    httpFailure(http)
} catch (serialization: SerializationException) {
    AgentResult.Failure(FailureReason.Malformed, "O agente respondeu num formato inesperado.")
}

private fun httpFailure(http: AgentHttpException): AgentResult.Failure {
    // 503 é o "collector indisponível" do agente: ele está de pé e sabe dizer
    // o que faltou, então a mensagem dele vale mais que qualquer texto nosso.
    if (http.status == 503) {
        return AgentResult.Failure(
            FailureReason.Unavailable,
            http.apiError?.message ?: "Métrica temporariamente indisponível.",
        )
    }

    return AgentResult.Failure(
        FailureReason.Unexpected,
        http.apiError?.message ?: "O agente respondeu com erro ${http.status}.",
    )
}
