package pcmonitor.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import pcmonitor.model.ProcessMetrics
import pcmonitor.ui.format.formatBytes
import pcmonitor.ui.format.formatPercent
import pcmonitor.ui.theme.gaugeColor

/**
 * Tabela de processos.
 *
 * `LazyColumn` e não `Column`: a lista chega com dezenas de linhas e é
 * substituída a cada atualização; compor todas as linhas a cada segundo
 * gastaria quadro à toa.
 *
 * A `key` por PID preserva o estado de rolagem quando a ordem muda entre
 * atualizações — sem ela, a lista pula para o topo a cada segundo.
 */
@Composable
fun ProcessTable(
    processes: List<ProcessMetrics>,
    modifier: Modifier = Modifier,
    emptyMessage: String = "Nenhum processo para exibir",
) {
    Card(
        modifier = modifier,
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
    ) {
        Column(modifier = Modifier.fillMaxSize()) {
            HeaderRow()

            if (processes.isEmpty()) {
                Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                    Text(
                        text = emptyMessage,
                        style = MaterialTheme.typography.bodyMedium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                    )
                }

                return@Column
            }

            LazyColumn(state = rememberLazyListState()) {
                items(processes, key = { it.pid }) { process ->
                    ProcessRow(process)
                }
            }
        }
    }
}

@Composable
private fun HeaderRow() {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .background(MaterialTheme.colorScheme.surfaceVariant)
            .padding(horizontal = 16.dp, vertical = 10.dp),
        horizontalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        HeaderCell("PID", Modifier.width(PID_WIDTH))
        HeaderCell("PROCESSO", Modifier.weight(1f))
        HeaderCell("CPU", Modifier.width(NUMBER_WIDTH), TextAlign.End)
        HeaderCell("MEMÓRIA", Modifier.width(NUMBER_WIDTH), TextAlign.End)
    }
}

@Composable
private fun HeaderCell(text: String, modifier: Modifier, align: TextAlign = TextAlign.Start) {
    Text(
        text = text,
        modifier = modifier,
        style = MaterialTheme.typography.labelSmall,
        color = MaterialTheme.colorScheme.onSurfaceVariant,
        textAlign = align,
    )
}

@Composable
private fun ProcessRow(process: ProcessMetrics) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 7.dp),
        horizontalArrangement = Arrangement.spacedBy(12.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(
            text = process.pid.toString(),
            modifier = Modifier.width(PID_WIDTH),
            style = MaterialTheme.typography.bodySmall,
            fontFamily = FontFamily.Monospace,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )

        Text(
            text = process.name,
            modifier = Modifier.weight(1f),
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurface,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
        )

        Text(
            text = formatPercent(process.cpuPercent),
            modifier = Modifier.width(NUMBER_WIDTH),
            style = MaterialTheme.typography.bodySmall,
            fontFamily = FontFamily.Monospace,
            textAlign = TextAlign.End,
            // A cor acompanha o peso do processo: numa lista de dezenas de
            // linhas, quem está pesando aparece sem precisar ler os números.
            color = gaugeColor(process.cpuPercent),
        )

        Text(
            text = formatBytes(process.memory),
            modifier = Modifier.width(NUMBER_WIDTH),
            style = MaterialTheme.typography.bodySmall,
            fontFamily = FontFamily.Monospace,
            textAlign = TextAlign.End,
            color = MaterialTheme.colorScheme.onSurface,
        )
    }
}

// Colunas de número têm largura fixa e o nome fica com o resto: com pesos
// proporcionais, uma janela larga afastava o PID do nome por centenas de
// pixels vazios.
private val PID_WIDTH = 80.dp
private val NUMBER_WIDTH = 110.dp
