package pcmonitor.ui.components

import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Path
import androidx.compose.ui.graphics.PathEffect
import androidx.compose.ui.graphics.drawscope.DrawScope
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.unit.dp
import pcmonitor.viewmodel.MetricsHistory
import kotlin.math.max

/** Uma linha do gráfico. */
data class ChartSeries(
    val label: String,
    val color: Color,
    val values: List<Double>,
)

/**
 * Gráfico de linha dos últimos 60 segundos.
 *
 * `ceiling` fixo em 100 para CPU e memória, automático para rede — não existe
 * teto conhecido de banda, e uma escala fixa deixaria a linha colada no chão
 * numa conexão rápida ou estourada numa lenta.
 */
@Composable
fun MetricChart(
    title: String,
    series: List<ChartSeries>,
    formatValue: (Double) -> String,
    modifier: Modifier = Modifier,
    fixedCeiling: Double? = null,
    minimumCeiling: Double = 1.0,
) {
    val ceiling = fixedCeiling ?: automaticCeiling(series, minimumCeiling)

    Card(
        modifier = modifier,
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
    ) {
        Column(
            modifier = Modifier.padding(16.dp).fillMaxWidth(),
            verticalArrangement = Arrangement.spacedBy(10.dp),
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Text(
                    text = title.uppercase(),
                    style = MaterialTheme.typography.labelMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )

                Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                    series.forEach { Legend(it, formatValue) }
                }
            }

            Box(modifier = Modifier.fillMaxWidth().height(80.dp)) {
                Canvas(modifier = Modifier.fillMaxWidth().height(80.dp)) {
                    drawGrid()

                    series.forEach { drawSeries(it, ceiling) }
                }
            }

            Text(
                text = "${formatValue(ceiling)} · 60 s",
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
        }
    }
}

@Composable
private fun Legend(series: ChartSeries, formatValue: (Double) -> String) {
    Row(horizontalArrangement = Arrangement.spacedBy(6.dp), verticalAlignment = Alignment.CenterVertically) {
        Box(modifier = Modifier.size(8.dp).background(series.color, CircleShape))

        Text(
            text = series.values.lastOrNull()?.let(formatValue) ?: "—",
            style = MaterialTheme.typography.labelMedium,
            fontFamily = FontFamily.Monospace,
            color = MaterialTheme.colorScheme.onSurface,
        )
    }
}

/**
 * Teto automático com folga de 15 %.
 *
 * Sem a folga, o pico encosta na borda de cima e some visualmente; sem o piso
 * mínimo, uma rede parada em zero dividiria por zero.
 */
private fun automaticCeiling(series: List<ChartSeries>, minimum: Double): Double {
    val peak = series.flatMap { it.values }.maxOrNull() ?: 0.0

    return max(peak * 1.15, minimum)
}

private fun DrawScope.drawGrid() {
    val tracejado = PathEffect.dashPathEffect(floatArrayOf(4f, 6f))

    // Três linhas: base, meio e topo da escala.
    for (fraction in listOf(0f, 0.5f, 1f)) {
        val y = size.height * fraction

        drawLine(
            color = Color.White.copy(alpha = 0.06f),
            start = Offset(0f, y),
            end = Offset(size.width, y),
            pathEffect = tracejado,
        )
    }
}

/**
 * Desenha uma série alinhada à direita.
 *
 * Os pontos entram pela direita e caminham para a esquerda conforme envelhecem.
 * Distribuir poucos pontos por toda a largura faria o gráfico "esticar" nos
 * primeiros segundos, como se o intervalo de tempo mudasse.
 */
private fun DrawScope.drawSeries(series: ChartSeries, ceiling: Double) {
    if (series.values.size < 2) {
        return
    }

    val stepX = size.width / (MetricsHistory.CAPACITY - 1)
    val offset = MetricsHistory.CAPACITY - series.values.size

    fun pointAt(index: Int): Offset {
        val value = series.values[index].coerceIn(0.0, ceiling)
        val y = size.height - (value / ceiling * size.height).toFloat()

        return Offset((offset + index) * stepX, y)
    }

    val linha = Path()
    val area = Path()

    val primeiro = pointAt(0)

    linha.moveTo(primeiro.x, primeiro.y)
    area.moveTo(primeiro.x, size.height)
    area.lineTo(primeiro.x, primeiro.y)

    for (index in 1 until series.values.size) {
        val ponto = pointAt(index)

        linha.lineTo(ponto.x, ponto.y)
        area.lineTo(ponto.x, ponto.y)
    }

    val ultimo = pointAt(series.values.lastIndex)

    area.lineTo(ultimo.x, size.height)
    area.close()

    drawPath(
        path = area,
        brush = Brush.verticalGradient(
            listOf(series.color.copy(alpha = 0.28f), series.color.copy(alpha = 0f)),
        ),
    )

    drawPath(path = linha, color = series.color, style = Stroke(width = 2f))

    drawCircle(color = series.color, radius = 3f, center = ultimo)
}
