package pcmonitor.ui.components

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.FilterChip
import androidx.compose.material3.FilterChipDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.OutlinedTextFieldDefaults
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.unit.dp
import pcmonitor.model.ProcessMetrics
import pcmonitor.viewmodel.ProcessSort

/**
 * Tabela de processos com filtro e ordenação.
 *
 * Filtro e sentido da ordem são estado local: ninguém fora daqui precisa saber
 * o que o usuário digitou. Já o **critério** sobe para o ViewModel, porque
 * muda o pedido feito ao agente — a lista chega cortada nos 50 primeiros, e
 * reordenar os 50 mais pesados de CPU por memória mostraria os processos
 * errados.
 */
@Composable
fun ProcessSection(
    processes: List<ProcessMetrics>,
    sort: ProcessSort,
    onSortChange: (ProcessSort) -> Unit,
    modifier: Modifier = Modifier,
) {
    var query by remember { mutableStateOf("") }
    var descending by remember { mutableStateOf(true) }

    val visible = remember(processes, query, sort, descending) {
        processes.filterByName(query).sortedBy(sort, descending)
    }

    Column(modifier = modifier, verticalArrangement = Arrangement.spacedBy(10.dp)) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(10.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(
                text = "PROCESSOS",
                style = MaterialTheme.typography.labelMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                modifier = Modifier.padding(end = 6.dp),
            )

            OutlinedTextField(
                value = query,
                onValueChange = { query = it },
                placeholder = { Text("Filtrar por nome", style = MaterialTheme.typography.bodySmall) },
                singleLine = true,
                keyboardOptions = KeyboardOptions(imeAction = ImeAction.Search),
                textStyle = MaterialTheme.typography.bodySmall,
                colors = OutlinedTextFieldDefaults.colors(
                    focusedContainerColor = MaterialTheme.colorScheme.surface,
                    unfocusedContainerColor = MaterialTheme.colorScheme.surface,
                ),
                modifier = Modifier.widthIn(max = 260.dp).weight(1f, fill = false),
            )

            SortChip("CPU", sort == ProcessSort.Cpu) { onSortChange(ProcessSort.Cpu) }
            SortChip("Memória", sort == ProcessSort.Memory) { onSortChange(ProcessSort.Memory) }

            SortChip(
                label = if (descending) "↓ Maior primeiro" else "↑ Menor primeiro",
                selected = false,
                onClick = { descending = !descending },
            )
        }

        ProcessTable(
            processes = visible,
            modifier = Modifier.fillMaxWidth().weight(1f, fill = false),
            emptyMessage = if (query.isBlank()) {
                "Nenhum processo para exibir"
            } else {
                "Nenhum processo corresponde a \"$query\""
            },
        )
    }
}

@Composable
private fun SortChip(label: String, selected: Boolean, onClick: () -> Unit) {
    FilterChip(
        selected = selected,
        onClick = onClick,
        label = { Text(label, style = MaterialTheme.typography.labelMedium) },
        colors = FilterChipDefaults.filterChipColors(
            containerColor = MaterialTheme.colorScheme.surface,
            selectedContainerColor = MaterialTheme.colorScheme.primary,
            selectedLabelColor = MaterialTheme.colorScheme.onPrimary,
        ),
    )
}

/**
 * Filtra por nome, sem diferenciar maiúsculas.
 *
 * Quem procura "chrome" não deve precisar saber se o processo se chama
 * "Chromium" ou "chromium".
 */
fun List<ProcessMetrics>.filterByName(query: String): List<ProcessMetrics> {
    val trimmed = query.trim()

    if (trimmed.isEmpty()) {
        return this
    }

    return filter { it.name.contains(trimmed, ignoreCase = true) }
}

/**
 * Ordena pelo critério escolhido.
 *
 * O PID desempata para que processos com o mesmo consumo não troquem de lugar
 * entre atualizações — a mesma regra que o agente aplica do lado dele.
 */
fun List<ProcessMetrics>.sortedBy(sort: ProcessSort, descending: Boolean): List<ProcessMetrics> {
    val byMetric = when (sort) {
        ProcessSort.Cpu -> compareBy<ProcessMetrics> { it.cpuPercent }
        ProcessSort.Memory -> compareBy<ProcessMetrics> { it.memory }
    }

    val ordered = if (descending) byMetric.reversed() else byMetric

    return sortedWith(ordered.thenBy { it.pid })
}
