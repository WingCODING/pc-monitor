package pcmonitor.ui.format

import kotlin.test.Test
import kotlin.test.assertEquals

class FormattersTest {
    @Test
    fun formataBytesEmUnidadeBinaria() {
        assertEquals("0 B", formatBytes(0))
        assertEquals("512 B", formatBytes(512))
        // O exemplo da spec: 1024 vira 1 KB, e não "1,0 KB".
        assertEquals("1 KB", formatBytes(1024))
        assertEquals("1,5 KB", formatBytes(1536))
        assertEquals("1 MB", formatBytes(1024L * 1024))
        assertEquals("30,2 GB", formatBytes(32_449_835_008))
        assertEquals("946 GB", formatBytes(1_015_740_882_944))
        assertEquals("1 TB", formatBytes(1024L * 1024 * 1024 * 1024))
    }

    @Test
    fun formataVelocidade() {
        assertEquals("0 B/s", formatBytesPerSecond(0.0))
        assertEquals("2,1 MB/s", formatBytesPerSecond(2_233_281.08))
    }

    @Test
    fun formataPercentualComUmaCasa() {
        assertEquals("0 %", formatPercent(0.0))
        assertEquals("45,3 %", formatPercent(45.27))
        assertEquals("100 %", formatPercent(100.0))
        // Processo multithread passa de 100 de propósito; o formatador não
        // pode "consertar" o valor.
        assertEquals("400 %", formatPercent(400.0))
    }

    @Test
    fun formataUptimeLegivel() {
        assertEquals("0 min", formatUptime(0))
        assertEquals("1 min", formatUptime(90))
        assertEquals("1 h 1 min", formatUptime(3_700))
        assertEquals("6 d 3 h 49 min", formatUptime(532_140))
        // Uptime negativo não existe; se vier, não pode virar "-1 d".
        assertEquals("—", formatUptime(-5))
    }

    @Test
    fun arredondaSemCasaDecimalInutil() {
        assertEquals("1", formatDecimal(1.0, 1))
        assertEquals("1,1", formatDecimal(1.05, 1))
        assertEquals("2", formatDecimal(1.96, 1))
        assertEquals("1,05", formatDecimal(1.05, 2))
        assertEquals("-1,5", formatDecimal(-1.5, 1))
        // A parte inteira arredonda para zero, mas o sinal não pode sumir.
        assertEquals("-0,5", formatDecimal(-0.5, 1))
    }
}
