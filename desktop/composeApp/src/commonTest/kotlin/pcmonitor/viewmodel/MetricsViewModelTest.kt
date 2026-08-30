package pcmonitor.viewmodel

import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.test.UnconfinedTestDispatcher
import kotlinx.coroutines.test.TestScope
import kotlinx.coroutines.test.runCurrent
import kotlinx.coroutines.test.runTest
import pcmonitor.model.CpuMetrics
import pcmonitor.model.DashboardMetrics
import pcmonitor.model.MemoryMetrics
import pcmonitor.model.ProcessMetrics
import pcmonitor.model.SystemMetrics
import pcmonitor.repository.AgentResult
import pcmonitor.repository.FailureReason
import pcmonitor.repository.MetricsSource
import pcmonitor.repository.ProcessesSource
import kotlin.test.Test
import kotlin.test.assertContains
import kotlin.test.assertEquals
import kotlin.test.assertNotNull
import kotlin.test.assertTrue

private val SISTEMA = SystemMetrics("omarchy", "linux", "4.0.1", "amd64", 23_176)

private fun snapshot(cpu: Double) = DashboardMetrics(
    timestamp = "2026-08-30T21:00:00Z",
    cpu = CpuMetrics("Ryzen", 16, 32, cpu),
    memory = MemoryMetrics(32, 16, 16, 50.0),
    uptimeSeconds = 23_176,
)

private class MetricsSourceFalso(
    private val fluxo: Flow<AgentResult<DashboardMetrics>>,
    private val sistema: AgentResult<SystemMetrics> = AgentResult.Success(SISTEMA),
) : MetricsSource {
    override suspend fun getMetrics(): AgentResult<DashboardMetrics> =
        AgentResult.Failure(FailureReason.Unexpected, "não usado pelo ViewModel")

    override suspend fun getSystem(): AgentResult<SystemMetrics> = sistema

    override fun observeMetrics(): Flow<AgentResult<DashboardMetrics>> = fluxo
}

private class ProcessesSourceFalso(
    private val resultado: AgentResult<List<ProcessMetrics>> = AgentResult.Success(emptyList()),
) : ProcessesSource {
    val criterios = mutableListOf<ProcessSort>()

    override suspend fun getProcesses(sort: ProcessSort, limit: Int): AgentResult<List<ProcessMetrics>> {
        criterios += sort

        return resultado
    }
}

class MetricsViewModelTest {
    private fun TestScope.viewModel(
        metrics: MetricsSource,
        processes: ProcessesSource = ProcessesSourceFalso(),
    ) = MetricsViewModel(
        metricsRepository = metrics,
        processesRepository = processes,
        scope = backgroundScope,
        dispatcher = UnconfinedTestDispatcher(testScheduler),
    )

    @Test
    fun comecaEmLoading() = runTest {
        val viewModel = viewModel(MetricsSourceFalso(MutableSharedFlow()))

        assertEquals(ConnectionState.Loading, viewModel.state.value.connection)
        assertEquals(null, viewModel.state.value.metrics)
    }

    @Test
    fun loadingViraConectadoComOPrimeiroSnapshot() = runTest {
        val fluxo = MutableSharedFlow<AgentResult<DashboardMetrics>>(extraBufferCapacity = 8)
        val viewModel = viewModel(MetricsSourceFalso(fluxo))

        viewModel.start()
        fluxo.emit(AgentResult.Success(snapshot(cpu = 12.5)))
        runCurrent()

        val estado = viewModel.state.value

        assertEquals(ConnectionState.Connected, estado.connection)
        assertEquals(12.5, estado.metrics?.cpu?.usagePercent)
        assertEquals(null, estado.message)
        assertEquals(SISTEMA, estado.system)
    }

    @Test
    fun falhaDeFormatoViraErroComMensagem() = runTest {
        val fluxo = MutableSharedFlow<AgentResult<DashboardMetrics>>(extraBufferCapacity = 8)
        val viewModel = viewModel(MetricsSourceFalso(fluxo))

        viewModel.start()
        fluxo.emit(AgentResult.Failure(FailureReason.Malformed, "O agente respondeu num formato inesperado."))
        runCurrent()

        val estado = viewModel.state.value

        assertEquals(ConnectionState.Error, estado.connection)
        assertNotNull(estado.message)
        // A UI precisa dizer que a repetição é automática, senão o usuário
        // procura um botão de reconectar que não existe.
        assertContains(estado.message!!, "Tentando novamente")
    }

    /**
     * O dashboard não pode esvaziar quando o agente cai: a última leitura boa
     * é o que interessa enquanto ele volta.
     */
    @Test
    fun quedaMantemAsMetricasNaTela() = runTest {
        val fluxo = MutableSharedFlow<AgentResult<DashboardMetrics>>(extraBufferCapacity = 8)
        val viewModel = viewModel(MetricsSourceFalso(fluxo))

        viewModel.start()
        fluxo.emit(AgentResult.Success(snapshot(cpu = 30.0)))
        runCurrent()

        fluxo.emit(AgentResult.Failure(FailureReason.Unreachable, "Não foi possível falar com o agente."))
        runCurrent()

        val estado = viewModel.state.value

        assertEquals(ConnectionState.Disconnected, estado.connection)
        assertEquals(30.0, estado.metrics?.cpu?.usagePercent)
    }

    @Test
    fun reconexaoVoltaParaConectado() = runTest {
        val fluxo = MutableSharedFlow<AgentResult<DashboardMetrics>>(extraBufferCapacity = 8)
        val viewModel = viewModel(MetricsSourceFalso(fluxo))

        viewModel.start()
        fluxo.emit(AgentResult.Failure(FailureReason.Unreachable, "fora do ar"))
        runCurrent()

        assertEquals(ConnectionState.Disconnected, viewModel.state.value.connection)

        fluxo.emit(AgentResult.Success(snapshot(cpu = 7.0)))
        runCurrent()

        val estado = viewModel.state.value

        assertEquals(ConnectionState.Connected, estado.connection)
        assertEquals(null, estado.message)
    }

    @Test
    fun historicoAcumulaOsSnapshots() = runTest {
        val fluxo = MutableSharedFlow<AgentResult<DashboardMetrics>>(extraBufferCapacity = 8)
        val viewModel = viewModel(MetricsSourceFalso(fluxo))

        viewModel.start()

        listOf(1.0, 2.0, 3.0).forEach {
            fluxo.emit(AgentResult.Success(snapshot(cpu = it)))
            runCurrent()
        }

        assertEquals(listOf(1.0, 2.0, 3.0), viewModel.state.value.history.cpu)
    }

    /**
     * Com as duas fontes mexendo no estado da ligação, a etiqueta oscilaria
     * entre Conectado e Desconectado nas cadências diferentes de cada uma.
     */
    @Test
    fun falhaEmProcessosNaoMexeNaLigacao() = runTest {
        val fluxo = MutableSharedFlow<AgentResult<DashboardMetrics>>(extraBufferCapacity = 8)
        val processos = ProcessesSourceFalso(AgentResult.Failure(FailureReason.Unavailable, "sem /proc"))
        val viewModel = viewModel(MetricsSourceFalso(fluxo), processos)

        viewModel.start()
        fluxo.emit(AgentResult.Success(snapshot(cpu = 5.0)))
        runCurrent()

        val estado = viewModel.state.value

        assertEquals(ConnectionState.Connected, estado.connection)
        assertTrue(estado.processes.isEmpty())
    }

    @Test
    fun trocarCriterioPedeNovaListaNaHora() = runTest {
        val processos = ProcessesSourceFalso()
        val viewModel = viewModel(MetricsSourceFalso(MutableSharedFlow()), processos)

        viewModel.start()
        runCurrent()

        assertEquals(listOf(ProcessSort.Cpu), processos.criterios)

        viewModel.setProcessSort(ProcessSort.Memory)
        runCurrent()

        assertEquals(ProcessSort.Memory, viewModel.state.value.processSort)
        // Sem esperar os dois segundos do intervalo.
        assertContains(processos.criterios, ProcessSort.Memory)
    }
}
