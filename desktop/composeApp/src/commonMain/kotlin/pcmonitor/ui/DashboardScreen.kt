package pcmonitor.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import pcmonitor.model.DashboardMetrics
import pcmonitor.model.ProcessMetrics
import pcmonitor.model.SystemMetrics
import pcmonitor.ui.components.ConnectionStatus
import pcmonitor.ui.components.MetricCard
import pcmonitor.ui.components.NetworkCard
import pcmonitor.ui.components.ProcessTable
import pcmonitor.ui.format.formatBytes
import pcmonitor.ui.format.formatPercent
import pcmonitor.ui.format.formatUptime
import pcmonitor.viewmodel.ConnectionState
import kotlin.math.max

/**
 * Tela única do dashboard: cabeçalho, área de métricas e área de processos.
 *
 * Recebe tudo por parâmetro e não guarda estado nenhum. Isso mantém o
 * Composable ignorante de onde os dados vêm — mock, REST ou WebSocket — e é o
 * que permitiu montar a tela inteira antes de existir integração.
 */
@Composable
fun DashboardScreen(
    metrics: DashboardMetrics?,
    processes: List<ProcessMetrics>,
    system: SystemMetrics?,
    connection: ConnectionState,
    modifier: Modifier = Modifier,
) {
    Column(
        modifier = modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
            .padding(20.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp),
    ) {
        Header(system, metrics?.uptimeSeconds, connection)

        MetricsArea(metrics, Modifier.fillMaxWidth())

        SectionTitle("Processos")

        ProcessTable(
            processes = processes,
            modifier = Modifier.fillMaxWidth().weight(1f),
        )
    }
}

@Composable
private fun Header(system: SystemMetrics?, uptimeSeconds: Long?, connection: ConnectionState) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
            Text(
                text = "PC Monitor",
                style = MaterialTheme.typography.headlineSmall,
                fontWeight = FontWeight.SemiBold,
                color = MaterialTheme.colorScheme.onBackground,
            )

            Text(
                text = machineLine(system, uptimeSeconds),
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
        }

        ConnectionStatus(connection)
    }
}

/**
 * Monta a linha de identificação da máquina com o que estiver disponível.
 *
 * Cada parte pode faltar — o endpoint de sistema pode ter falhado, o uptime
 * pode não estar no snapshot — e o cabeçalho não deve ficar com traços soltos
 * por causa disso.
 */
private fun machineLine(system: SystemMetrics?, uptimeSeconds: Long?): String {
    val parts = buildList {
        system?.hostname?.takeIf { it.isNotBlank() }?.let(::add)
        system?.let { "${it.os} ${it.osVersion}".trim() }?.takeIf { it.isNotBlank() }?.let(::add)
        system?.architecture?.takeIf { it.isNotBlank() }?.let(::add)
        (uptimeSeconds ?: system?.uptimeSeconds)?.let { add("ligado há ${formatUptime(it)}") }
    }

    return if (parts.isEmpty()) "aguardando o agente" else parts.joinToString("  ·  ")
}

@Composable
private fun SectionTitle(text: String) {
    Text(
        text = text.uppercase(),
        style = MaterialTheme.typography.labelMedium,
        color = MaterialTheme.colorScheme.onSurfaceVariant,
    )
}

/**
 * Área de métricas em grade fluida.
 *
 * O número de colunas sai da largura disponível em vez de ser fixo: a janela é
 * redimensionável, e uma grade de quatro colunas num monitor estreito
 * espremeria os cartões até os números não caberem.
 */
@Composable
private fun MetricsArea(metrics: DashboardMetrics?, modifier: Modifier = Modifier) {
    val cards: List<@Composable (Modifier) -> Unit> = listOf(
        { cardModifier ->
            MetricCard(
                title = "CPU",
                value = metrics?.cpu?.let { formatPercent(it.usagePercent) },
                detail = metrics?.cpu?.let { "${it.model}  ·  ${it.cores}C / ${it.threads}T" },
                percent = metrics?.cpu?.usagePercent,
                showPercentBadge = false,
                modifier = cardModifier,
            )
        },
        { cardModifier ->
            MetricCard(
                title = "Memória",
                value = metrics?.memory?.let { formatBytes(it.used) },
                detail = metrics?.memory?.let { "de ${formatBytes(it.total)}  ·  ${formatBytes(it.available)} livres" },
                percent = metrics?.memory?.usagePercent,
                modifier = cardModifier,
            )
        },
        { cardModifier ->
            MetricCard(
                title = "Disco",
                value = metrics?.disk?.let { formatBytes(it.used) },
                detail = metrics?.disk?.let { "de ${formatBytes(it.total)}" },
                percent = metrics?.disk?.usagePercent,
                modifier = cardModifier,
            )
        },
        { cardModifier ->
            NetworkCard(
                downloadBytesPerSecond = metrics?.network?.downloadBytesPerSecond,
                uploadBytesPerSecond = metrics?.network?.uploadBytesPerSecond,
                modifier = cardModifier,
            )
        },
    )

    BoxWithConstraints(modifier) {
        val columns = max(1, (maxWidth / MIN_CARD_WIDTH).toInt())

        Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
            cards.chunked(columns).forEach { row ->
                Row(horizontalArrangement = Arrangement.spacedBy(16.dp)) {
                    row.forEach { card -> card(Modifier.weight(1f)) }

                    // Mantém a largura dos cartões da última linha igual à das
                    // linhas cheias, em vez de esticá-los para ocupar o vão.
                    repeat(columns - row.size) { Spacer(Modifier.weight(1f)) }
                }
            }
        }
    }
}

private val MIN_CARD_WIDTH: Dp = 270.dp
