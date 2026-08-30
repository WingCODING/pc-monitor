package pcmonitor.network

import io.ktor.client.HttpClient
import io.ktor.client.call.body
import io.ktor.client.plugins.HttpTimeout
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.client.plugins.websocket.WebSockets
import io.ktor.client.request.get
import io.ktor.client.request.parameter
import io.ktor.client.statement.HttpResponse
import io.ktor.http.isSuccess
import io.ktor.serialization.kotlinx.json.json
import kotlinx.serialization.json.Json
import pcmonitor.model.DashboardMetrics
import pcmonitor.model.ProcessMetrics
import pcmonitor.model.SystemMetrics

/** Endereço do agente quando nada é informado. */
const val DEFAULT_AGENT_URL = "http://127.0.0.1:8080"

/**
 * Cliente HTTP do agente.
 *
 * É a única classe do frontend que conhece Ktor. Repositories recebem esta
 * interface, e os Composables nunca a veem — a regra da Task 12.3 só se
 * sustenta porque existe exatamente um lugar onde uma URL é montada.
 */
class AgentClient(
    private val baseUrl: String = DEFAULT_AGENT_URL,
    private val httpClient: HttpClient = defaultHttpClient(),
) {
    suspend fun metrics(): DashboardMetrics = httpClient.get("$baseUrl/api/v1/metrics").decode()

    suspend fun system(): SystemMetrics = httpClient.get("$baseUrl/api/v1/system").decode()

    suspend fun processes(sort: String = "cpu", limit: Int = 50): List<ProcessMetrics> =
        httpClient.get("$baseUrl/api/v1/processes") {
            parameter("sort", sort)
            parameter("limit", limit)
        }.decode()

    /** Endereço do WebSocket de métricas, derivado da mesma base. */
    fun metricsSocketUrl(): String = baseUrl.replaceFirst("http", "ws") + "/api/v1/ws/metrics"

    fun close() = httpClient.close()

    /**
     * Converte a resposta no tipo esperado, transformando status de erro em
     * exceção com o corpo do agente já interpretado.
     *
     * Sem isso, um 503 seria entregue como falha de desserialização — o
     * dashboard mostraria "resposta inválida" onde a resposta era um erro
     * perfeitamente bem formado dizendo que o collector está fora do ar.
     */
    private suspend inline fun <reified T> HttpResponse.decode(): T {
        if (status.isSuccess()) {
            return body()
        }

        throw AgentHttpException(status.value, runCatching { body<ApiError>() }.getOrNull())
    }
}

/**
 * Cliente Ktor com os prazos que fazem sentido para um agente local.
 *
 * Os tempos são curtos de propósito: o agente roda em localhost, e uma
 * requisição que passe de alguns segundos significa que ele travou. Esperar
 * mais só atrasaria a mudança da UI para "desconectado".
 */
fun defaultHttpClient(): HttpClient = HttpClient {
    install(ContentNegotiation) {
        json(agentJson)
    }

    install(HttpTimeout) {
        connectTimeoutMillis = 1_500
        requestTimeoutMillis = 4_000
    }

    install(WebSockets)
}

/**
 * `ignoreUnknownKeys` deixa o agente ganhar campos novos sem quebrar uma
 * versão antiga do desktop — o contrário do que aconteceria com o padrão
 * estrito do kotlinx.serialization.
 */
val agentJson: Json = Json {
    ignoreUnknownKeys = true
}

/** Corpo de erro padronizado do agente. */
@kotlinx.serialization.Serializable
data class ApiError(val error: String, val message: String)

/** Resposta com status de erro; carrega o corpo do agente quando ele veio. */
class AgentHttpException(val status: Int, val apiError: ApiError?) :
    Exception("agente respondeu ${status}${apiError?.let { ": ${it.error}" } ?: ""}")
