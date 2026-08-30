package pcmonitor.viewmodel

import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.currentCoroutineContext
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import kotlinx.coroutines.withTimeoutOrNull
import pcmonitor.repository.AgentResult
import pcmonitor.repository.FailureReason
import pcmonitor.repository.MetricsRepository
import pcmonitor.repository.ProcessesRepository
import kotlin.time.Duration
import kotlin.time.Duration.Companion.seconds

/**
 * Estado do dashboard a partir dos repositories.
 *
 * Não é um ViewModel de framework: o desktop não tem ciclo de vida de Android,
 * então o escopo de corrotinas vem de fora — da janela, nos testes de um
 * `TestScope`. Nenhuma linha aqui conhece HTTP.
 */
class MetricsViewModel(
    private val metricsRepository: MetricsRepository,
    private val processesRepository: ProcessesRepository,
    private val scope: CoroutineScope,
    // A varredura de processos custa uma volta inteira em /proc no agente;
    // pedir na mesma cadência das métricas gastaria CPU para mostrar uma
    // tabela que ninguém consegue ler mudando a cada segundo.
    private val processesInterval: Duration = 2.seconds,
    private val systemRetryInterval: Duration = 5.seconds,
) {
    private val mutableState = MutableStateFlow(DashboardUiState())
    val state: StateFlow<DashboardUiState> = mutableState.asStateFlow()

    /**
     * Critério pedido ao agente. Separado do estado da UI porque o laço de
     * processos precisa esperar por mudanças nele, e não por qualquer
     * atualização de métrica.
     */
    private val processSort = MutableStateFlow(ProcessSort.Cpu)

    private val jobs = mutableListOf<Job>()

    /** Começa a acompanhar o agente. Chamar duas vezes não duplica os loops. */
    fun start() {
        if (jobs.isNotEmpty()) {
            return
        }

        // Dispatchers.Default: o escopo vem da composição, que roda na thread
        // da interface. Desserializar um snapshot por segundo e uma lista de
        // 50 processos a cada dois lá disputaria o tempo do desenho.
        jobs += scope.launch(Dispatchers.Default) { observeMetrics() }
        jobs += scope.launch(Dispatchers.Default) { pollProcesses() }
        jobs += scope.launch(Dispatchers.Default) { loadSystem() }
    }

    /** Troca o critério de ordenação e refaz o pedido sem esperar o intervalo. */
    fun setProcessSort(sort: ProcessSort) {
        processSort.value = sort
        mutableState.update { it.copy(processSort = sort) }
    }

    fun stop() {
        jobs.forEach(Job::cancel)
        jobs.clear()
    }

    /**
     * Acompanha o fluxo de snapshots do WebSocket.
     *
     * O fluxo do repository já reconecta sozinho e nunca termina, então não há
     * laço aqui — e é justamente por existir um único coletor que uma
     * reconexão não deixa duas sessões vivas.
     */
    private suspend fun observeMetrics() {
        metricsRepository.observeMetrics().collect { result ->
            when (result) {
                is AgentResult.Success -> mutableState.update {
                    it.copy(
                        connection = ConnectionState.Connected,
                        metrics = result.value,
                        history = it.history.plus(result.value),
                        message = null,
                    )
                }

                is AgentResult.Failure -> mutableState.update {
                    it.copy(
                        connection = result.reason.toConnectionState(),
                        // A repetição é automática; dizer isso evita que o
                        // usuário fique procurando um botão de reconectar que
                        // não existe.
                        message = "${result.message} Tentando novamente…",
                    )
                }
            }
        }
    }

    /**
     * A lista de processos não mexe no estado da ligação.
     *
     * Quem manda nele é o fluxo de métricas: se as duas fontes escrevessem no
     * mesmo campo em cadências diferentes, a etiqueta de conexão ficaria
     * oscilando entre Conectado e Desconectado a cada segundo.
     */
    private suspend fun pollProcesses() {
        while (currentCoroutineContext().isActive) {
            val sort = processSort.value

            val result = processesRepository.getProcesses(sort)

            if (result is AgentResult.Success) {
                mutableState.update { it.copy(processes = result.value) }
            }

            // Espera o intervalo, mas acorda na hora se o usuário trocar o
            // critério: esperar dois segundos depois de um clique faria a
            // interface parecer travada.
            withTimeoutOrNull(processesInterval) { processSort.first { it != sort } }
        }
    }

    /**
     * Informações da máquina são lidas uma vez e não mudam mais.
     *
     * A tentativa se repete enquanto falhar porque o desktop pode abrir antes
     * do agente — caso comum quando os dois sobem juntos.
     */
    private suspend fun loadSystem() {
        while (currentCoroutineContext().isActive) {
            val result = metricsRepository.getSystem()

            if (result is AgentResult.Success) {
                mutableState.update { it.copy(system = result.value) }

                return
            }

            delay(systemRetryInterval)
        }
    }
}

private fun FailureReason.toConnectionState(): ConnectionState = when (this) {
    // Só a falha de transporte significa "o agente não está lá"; as demais
    // significam que ele está e respondeu algo que a UI não esperava.
    FailureReason.Unreachable -> ConnectionState.Disconnected
    FailureReason.Unavailable, FailureReason.Malformed, FailureReason.Unexpected -> ConnectionState.Error
}

