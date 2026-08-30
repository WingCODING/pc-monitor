package pcmonitor.viewmodel

/**
 * Critério de ordenação da lista de processos.
 *
 * Vive no ViewModel e não na UI porque muda o pedido feito ao agente: a lista
 * chega cortada nos 50 primeiros, e reordenar localmente os 50 mais pesados de
 * CPU por memória mostraria os processos errados.
 */
enum class ProcessSort(val apiValue: String) {
    Cpu("cpu"),
    Memory("memory"),
}
