package pcmonitor.ui.components

import pcmonitor.model.ProcessMetrics
import pcmonitor.viewmodel.ProcessSort
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class ProcessListingTest {
    private val processos = listOf(
        ProcessMetrics(pid = 4321, name = "chromium", cpuPercent = 12.5, memory = 2_000_000),
        ProcessMetrics(pid = 900, name = "java", cpuPercent = 180.0, memory = 1_000_000),
        ProcessMetrics(pid = 1, name = "systemd", cpuPercent = 0.0, memory = 3_000_000),
        ProcessMetrics(pid = 12, name = "kworker/0:1", cpuPercent = 0.0, memory = 0),
    )

    @Test
    fun filtraSemDiferenciarMaiusculas() {
        assertEquals(listOf("chromium"), processos.filterByName("CHROM").map { it.name })
        assertEquals(listOf("chromium"), processos.filterByName("chrom").map { it.name })
    }

    @Test
    fun filtroVazioDevolveTudo() {
        assertEquals(processos.size, processos.filterByName("").size)
        assertEquals(processos.size, processos.filterByName("   ").size)
    }

    @Test
    fun filtroSemCorrespondenciaDevolveListaVazia() {
        assertTrue(processos.filterByName("postgres").isEmpty())
    }

    @Test
    fun ordenaPorCpuDecrescente() {
        val ordenado = processos.sortedBy(ProcessSort.Cpu, descending = true)

        assertEquals(listOf(900, 4321, 1, 12), ordenado.map { it.pid })
    }

    @Test
    fun ordenaPorMemoriaDecrescente() {
        val ordenado = processos.sortedBy(ProcessSort.Memory, descending = true)

        assertEquals(listOf(1, 4321, 900, 12), ordenado.map { it.pid })
    }

    @Test
    fun ordenaCrescenteQuandoPedido() {
        val ordenado = processos.sortedBy(ProcessSort.Cpu, descending = false)

        assertEquals(listOf(1, 12, 4321, 900), ordenado.map { it.pid })
    }

    /**
     * Sem desempate por PID, os processos parados em 0 % trocariam de lugar a
     * cada atualização e a tabela ficaria piscando.
     */
    @Test
    fun desempatePorPidTornaAOrdemEstavel() {
        val embaralhado = processos.reversed()

        assertEquals(
            processos.sortedBy(ProcessSort.Cpu, descending = true).map { it.pid },
            embaralhado.sortedBy(ProcessSort.Cpu, descending = true).map { it.pid },
        )
    }
}
