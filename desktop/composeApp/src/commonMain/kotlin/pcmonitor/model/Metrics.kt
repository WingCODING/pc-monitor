package pcmonitor.model

import kotlinx.serialization.Serializable

/**
 * Modelos de domínio do dashboard.
 *
 * Espelham os contratos REST do agente Go, mas não conhecem HTTP: quem faz a
 * chamada é a camada `network`, e a UI recebe estes tipos já prontos.
 *
 * Os nomes das propriedades são exatamente os nomes dos campos JSON do agente
 * — nenhuma `@SerialName` é necessária, e uma renomeação aqui quebra o
 * contrato tanto quanto uma renomeação lá.
 *
 * Os campos opcionais são anuláveis pelo mesmo motivo que são `omitempty` no
 * Go: a falha de um collector omite aquela métrica do snapshot sem derrubar as
 * demais, e a UI precisa distinguir "0 %" de "indisponível".
 */
@Serializable
data class CpuMetrics(
    val model: String,
    val cores: Int,
    val threads: Int,
    val usagePercent: Double,
)

@Serializable
data class MemoryMetrics(
    val total: Long,
    val used: Long,
    val available: Long,
    val usagePercent: Double,
)

@Serializable
data class DiskSummary(
    val total: Long,
    val used: Long,
    val usagePercent: Double,
)

@Serializable
data class NetworkSummary(
    val downloadBytesPerSecond: Double,
    val uploadBytesPerSecond: Double,
)

@Serializable
data class SystemMetrics(
    val hostname: String,
    val os: String,
    val osVersion: String,
    val architecture: String,
    val uptimeSeconds: Long,
)

@Serializable
data class ProcessMetrics(
    val pid: Int,
    val name: String,
    /** Relativo a um núcleo, como no top: passa de 100 em processos multithread. */
    val cpuPercent: Double,
    /** Memória residente, em bytes. */
    val memory: Long,
)

/** Snapshot agregado, servido por `GET /api/v1/metrics` e pelo WebSocket. */
@Serializable
data class DashboardMetrics(
    val timestamp: String,
    val cpu: CpuMetrics? = null,
    val memory: MemoryMetrics? = null,
    val disk: DiskSummary? = null,
    val network: NetworkSummary? = null,
    val uptimeSeconds: Long? = null,
)
