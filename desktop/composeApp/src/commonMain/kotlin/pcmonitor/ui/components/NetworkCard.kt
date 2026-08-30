package pcmonitor.ui.components

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import pcmonitor.ui.format.formatBytesPerSecond
import pcmonitor.ui.theme.StatusColors

/**
 * Cartão de rede: uma linha para download e outra para upload.
 *
 * Não usa o MetricCard porque rede não tem percentual — não existe "100 % da
 * rede". São duas velocidades independentes, e forçá-las num cartão de medidor
 * exigiria inventar um teto.
 */
@Composable
fun NetworkCard(
    downloadBytesPerSecond: Double?,
    uploadBytesPerSecond: Double?,
    modifier: Modifier = Modifier,
) {
    Card(
        modifier = modifier,
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
    ) {
        Column(
            modifier = Modifier.padding(18.dp).fillMaxWidth(),
            verticalArrangement = Arrangement.spacedBy(10.dp),
        ) {
            Text(
                text = "REDE",
                style = MaterialTheme.typography.labelMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )

            TransferRow("↓", "Download", downloadBytesPerSecond, StatusColors.ok)
            TransferRow("↑", "Upload", uploadBytesPerSecond, MaterialTheme.colorScheme.primary)
        }
    }
}

@Composable
private fun TransferRow(arrow: String, label: String, bytesPerSecond: Double?, color: Color) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.spacedBy(10.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(text = arrow, style = MaterialTheme.typography.titleMedium, color = color)

        Text(
            text = label,
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            modifier = Modifier.weight(1f),
        )

        Text(
            text = bytesPerSecond?.let { formatBytesPerSecond(it) } ?: "—",
            style = MaterialTheme.typography.titleMedium,
            fontFamily = FontFamily.Monospace,
            fontWeight = FontWeight.Medium,
            color = MaterialTheme.colorScheme.onSurface,
        )
    }
}
