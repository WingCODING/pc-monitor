package pcmonitor.viewmodel

import pcmonitor.model.CpuMetrics
import pcmonitor.model.DashboardMetrics
import pcmonitor.model.MemoryMetrics
import pcmonitor.model.NetworkSummary
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class MetricsHistoryTest {
    private fun snapshot(
        cpu: Double? = 10.0,
        memory: Double? = 50.0,
        download: Double? = 1024.0,
    ) = DashboardMetrics(
        timestamp = "2026-08-30T21:00:00Z",
        cpu = cpu?.let { CpuMetrics("Ryzen", 16, 32, it) },
        memory = memory?.let { MemoryMetrics(32, 16, 16, it) },
        network = download?.let { NetworkSummary(it, 512.0) },
    )

    @Test
    fun acumulaNaOrdemDeChegada() {
        val history = MetricsHistory()
            .plus(snapshot(cpu = 1.0))
            .plus(snapshot(cpu = 2.0))
            .plus(snapshot(cpu = 3.0))

        assertEquals(listOf(1.0, 2.0, 3.0), history.cpu)
    }

    @Test
    fun descartaOsMaisAntigosAoPassarDaCapacidade() {
        var history = MetricsHistory()

        repeat(MetricsHistory.CAPACITY + 10) { index ->
            history = history.plus(snapshot(cpu = index.toDouble()))
        }

        assertEquals(MetricsHistory.CAPACITY, history.cpu.size)
        assertEquals(10.0, history.cpu.first())
        assertEquals((MetricsHistory.CAPACITY + 9).toDouble(), history.cpu.last())
    }

    /**
     * Métrica ausente não entra na série e não vira zero: um zero desenhado
     * diria "a máquina estava parada" onde a verdade é "não sabemos".
     */
    @Test
    fun metricaAusenteNaoEntraNaSerie() {
        val history = MetricsHistory()
            .plus(snapshot(cpu = 10.0))
            .plus(snapshot(cpu = null))
            .plus(snapshot(cpu = 30.0))

        assertEquals(listOf(10.0, 30.0), history.cpu)
        assertTrue(history.cpu.none { it == 0.0 })
    }

    @Test
    fun seriesSaoIndependentes() {
        val history = MetricsHistory()
            .plus(snapshot(cpu = null, memory = 50.0, download = 100.0))
            .plus(snapshot(cpu = 20.0, memory = null, download = null))

        assertEquals(listOf(20.0), history.cpu)
        assertEquals(listOf(50.0), history.memory)
        assertEquals(listOf(100.0), history.download)
        assertEquals(listOf(512.0), history.upload)
    }
}
