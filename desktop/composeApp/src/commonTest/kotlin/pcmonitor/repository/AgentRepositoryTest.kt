package pcmonitor.repository

import io.ktor.client.HttpClient
import io.ktor.client.engine.mock.MockEngine
import io.ktor.client.engine.mock.MockRequestHandleScope
import io.ktor.client.engine.mock.respond
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.client.request.HttpRequestData
import io.ktor.client.request.HttpResponseData
import io.ktor.http.HttpHeaders
import io.ktor.http.HttpStatusCode
import io.ktor.http.headersOf
import io.ktor.serialization.kotlinx.json.json
import io.ktor.utils.io.errors.IOException
import kotlinx.coroutines.test.runTest
import pcmonitor.network.AgentClient
import pcmonitor.network.agentJson
import pcmonitor.viewmodel.ProcessSort
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertIs
import kotlin.test.assertNull
import kotlin.test.assertTrue

private const val SNAPSHOT_COMPLETO = """
{"timestamp":"2026-08-30T21:01:05Z",
 "cpu":{"model":"AMD Ryzen 9 9950X3D","cores":16,"threads":32,"usagePercent":5.96},
 "memory":{"total":32449835008,"used":13086269440,"available":19363565568,"usagePercent":40.32},
 "disk":{"total":1015740882944,"used":389397245952,"usagePercent":38.33},
 "network":{"downloadBytesPerSecond":71471.76,"uploadBytesPerSecond":0},
 "uptimeSeconds":23176}
"""

class AgentRepositoryTest {
    private var ultimaRequisicao: HttpRequestData? = null

    private fun clienteQueResponde(
        handler: suspend MockRequestHandleScope.(HttpRequestData) -> HttpResponseData,
    ): AgentClient {
        val httpClient = HttpClient(MockEngine) {
            engine {
                addHandler { request ->
                    ultimaRequisicao = request
                    handler(request)
                }
            }

            install(ContentNegotiation) { json(agentJson) }
        }

        return AgentClient("http://agente-de-teste", httpClient)
    }

    private fun MockRequestHandleScope.respondeJson(corpo: String, status: HttpStatusCode = HttpStatusCode.OK) =
        respond(corpo, status, headersOf(HttpHeaders.ContentType, "application/json"))

    @Test
    fun leSnapshotCompleto() = runTest {
        val repository = MetricsRepository(clienteQueResponde { respondeJson(SNAPSHOT_COMPLETO) })

        val resultado = repository.getMetrics()

        val sucesso = assertIs<AgentResult.Success<*>>(resultado)
        val snapshot = assertIs<pcmonitor.model.DashboardMetrics>(sucesso.value)

        assertEquals(16, snapshot.cpu?.cores)
        assertEquals(32_449_835_008, snapshot.memory?.total)
        assertEquals(23_176, snapshot.uptimeSeconds)
        assertEquals("/api/v1/metrics", ultimaRequisicao?.url?.encodedPath)
    }

    /**
     * Métrica omitida pelo agente tem de chegar como `null`, e não como zero:
     * a UI distingue "0 %" de "indisponível".
     */
    @Test
    fun metricaOmitidaViraNulo() = runTest {
        val repository = MetricsRepository(
            clienteQueResponde { respondeJson("""{"timestamp":"2026-08-30T21:01:05Z","cpu":{"model":"x","cores":1,"threads":1,"usagePercent":3.0}}""") },
        )

        val sucesso = assertIs<AgentResult.Success<*>>(repository.getMetrics())
        val snapshot = assertIs<pcmonitor.model.DashboardMetrics>(sucesso.value)

        assertNull(snapshot.memory)
        assertNull(snapshot.disk)
        assertNull(snapshot.network)
        assertEquals(3.0, snapshot.cpu?.usagePercent)
    }

    /** Campo novo no agente não pode derrubar uma versão antiga do desktop. */
    @Test
    fun campoDesconhecidoEIgnorado() = runTest {
        val repository = MetricsRepository(
            clienteQueResponde { respondeJson("""{"timestamp":"t","gpu":{"usagePercent":42.0},"uptimeSeconds":10}""") },
        )

        val sucesso = assertIs<AgentResult.Success<*>>(repository.getMetrics())
        val snapshot = assertIs<pcmonitor.model.DashboardMetrics>(sucesso.value)

        assertEquals(10, snapshot.uptimeSeconds)
    }

    @Test
    fun jsonInvalidoViraFalhaDeFormato() = runTest {
        val repository = MetricsRepository(clienteQueResponde { respondeJson("{\"timestamp\":") })

        val falha = assertIs<AgentResult.Failure>(repository.getMetrics())

        assertEquals(FailureReason.Malformed, falha.reason)
    }

    /**
     * 503 é o agente dizendo que o collector caiu. Sem interpretar o corpo, a
     * falha chegaria como erro de desserialização e a tela diria "resposta
     * inválida" para uma resposta perfeitamente bem formada.
     */
    @Test
    fun erroDoAgenteMantemAMensagemDele() = runTest {
        val repository = MetricsRepository(
            clienteQueResponde {
                respondeJson(
                    """{"error":"collector_unavailable","message":"métrica temporariamente indisponível"}""",
                    HttpStatusCode.ServiceUnavailable,
                )
            },
        )

        val falha = assertIs<AgentResult.Failure>(repository.getMetrics())

        assertEquals(FailureReason.Unavailable, falha.reason)
        assertEquals("métrica temporariamente indisponível", falha.message)
    }

    @Test
    fun statusInesperadoViraFalhaInesperada() = runTest {
        val repository = MetricsRepository(clienteQueResponde { respondeJson("nada", HttpStatusCode.InternalServerError) })

        val falha = assertIs<AgentResult.Failure>(repository.getMetrics())

        assertEquals(FailureReason.Unexpected, falha.reason)
    }

    @Test
    fun agenteForaDoArViraIndisponivel() = runTest {
        val repository = MetricsRepository(clienteQueResponde { throw IOException("conexão recusada") })

        val falha = assertIs<AgentResult.Failure>(repository.getMetrics())

        assertEquals(FailureReason.Unreachable, falha.reason)
    }

    @Test
    fun listaDeProcessosVaziaESucesso() = runTest {
        val repository = ProcessesRepository(clienteQueResponde { respondeJson("[]") })

        val sucesso = assertIs<AgentResult.Success<List<*>>>(repository.getProcesses())

        assertTrue(sucesso.value.isEmpty())
    }

    @Test
    fun processosLevamCriterioELimiteNaConsulta() = runTest {
        val repository = ProcessesRepository(clienteQueResponde { respondeJson("[]") })

        repository.getProcesses(sort = ProcessSort.Memory, limit = 25)

        val parametros = ultimaRequisicao?.url?.parameters

        assertEquals("memory", parametros?.get("sort"))
        assertEquals("25", parametros?.get("limit"))
    }

    @Test
    fun leProcessosReais() = runTest {
        val repository = ProcessesRepository(
            clienteQueResponde {
                respondeJson("""[{"pid":900,"name":"java","cpuPercent":180.5,"memory":431616000}]""")
            },
        )

        val sucesso = assertIs<AgentResult.Success<List<pcmonitor.model.ProcessMetrics>>>(repository.getProcesses())

        assertEquals(900, sucesso.value.single().pid)
        assertEquals(180.5, sucesso.value.single().cpuPercent)
    }
}
