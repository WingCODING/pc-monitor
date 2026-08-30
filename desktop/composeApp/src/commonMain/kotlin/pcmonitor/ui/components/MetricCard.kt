package pcmonitor.ui.components

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import pcmonitor.ui.format.formatPercent
import pcmonitor.ui.theme.gaugeColor

/**
 * Cartão de uma métrica: título, valor em destaque, detalhe e um medidor.
 *
 * Serve CPU, RAM e disco. O medidor é opcional porque nem toda métrica tem um
 * teto conhecido — a rede, por exemplo, não tem percentual algum.
 *
 * O valor ausente é tratado aqui, e não por quem chama: quando um collector
 * falha, o cartão continua na tela mostrando "indisponível" em vez de sumir e
 * fazer o layout inteiro pular.
 */
@Composable
fun MetricCard(
    title: String,
    value: String?,
    detail: String? = null,
    percent: Double? = null,
    // O percentual aparece ao lado do valor quando o valor é uma quantidade
    // absoluta ("12,2 GB" de RAM). Quando o próprio valor já é o percentual,
    // como na CPU, repeti-lo ao lado só produziria "6 % 6 %".
    showPercentBadge: Boolean = true,
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
                text = title.uppercase(),
                style = MaterialTheme.typography.labelMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )

            Row(verticalAlignment = Alignment.Bottom, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                Text(
                    text = value ?: "—",
                    style = MaterialTheme.typography.headlineMedium,
                    fontFamily = FontFamily.Monospace,
                    fontWeight = FontWeight.Medium,
                    color = MaterialTheme.colorScheme.onSurface,
                )

                if (showPercentBadge && value != null && percent != null) {
                    Text(
                        text = formatPercent(percent),
                        style = MaterialTheme.typography.bodyMedium,
                        fontFamily = FontFamily.Monospace,
                        color = gaugeColor(percent),
                        modifier = Modifier.padding(bottom = 4.dp),
                    )
                }
            }

            Text(
                text = detail ?: if (value == null) "indisponível" else "",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )

            if (percent != null) {
                LinearProgressIndicator(
                    // O medidor é limitado a 0–100 aqui e não no modelo: o
                    // backend pode publicar 100,4 por arredondamento entre
                    // núcleos, e uma barra passando do fim é artefato visual.
                    progress = { (percent.coerceIn(0.0, 100.0) / 100.0).toFloat() },
                    modifier = Modifier.fillMaxWidth().height(6.dp),
                    color = gaugeColor(percent),
                    trackColor = MaterialTheme.colorScheme.surfaceVariant,
                    strokeCap = StrokeCap.Round,
                )
            }
        }
    }
}
