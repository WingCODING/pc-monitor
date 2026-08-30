package pcmonitor.viewmodel

import pcmonitor.model.DashboardMetrics

/**
 * Últimos pontos de cada série, para os gráficos.
 *
 * Cada série é independente: um snapshot com a CPU indisponível não entra na
 * série de CPU, mas continua alimentando as outras. A alternativa — furar
 * todas as séries junto — jogaria fora leituras boas por causa de um collector
 * quebrado.
 *
 * Ausência não vira zero. Um buraco desenhado como zero mentiria: "a máquina
 * estava parada" em vez de "não sabemos o que houve aqui".
 */
data class MetricsHistory(
    val cpu: List<Double> = emptyList(),
    val memory: List<Double> = emptyList(),
    val download: List<Double> = emptyList(),
    val upload: List<Double> = emptyList(),
) {
    /** Devolve um histórico novo com o snapshot no fim, descartando o excesso. */
    fun plus(snapshot: DashboardMetrics): MetricsHistory = MetricsHistory(
        cpu = cpu.append(snapshot.cpu?.usagePercent),
        memory = memory.append(snapshot.memory?.usagePercent),
        download = download.append(snapshot.network?.downloadBytesPerSecond),
        upload = upload.append(snapshot.network?.uploadBytesPerSecond),
    )

    companion object {
        /**
         * Um ponto por segundo durante um minuto.
         *
         * Sessenta cópias de uma lista de sessenta doubles por minuto é ruído
         * perto de qualquer outra coisa que a tela faz; um buffer mutável
         * compartilhado exigiria sincronização com o Compose para ganhar nada.
         */
        const val CAPACITY = 60
    }
}

private fun List<Double>.append(value: Double?): List<Double> {
    if (value == null) {
        return this
    }

    val appended = this + value

    return if (appended.size > MetricsHistory.CAPACITY) {
        appended.subList(appended.size - MetricsHistory.CAPACITY, appended.size)
    } else {
        appended
    }
}
