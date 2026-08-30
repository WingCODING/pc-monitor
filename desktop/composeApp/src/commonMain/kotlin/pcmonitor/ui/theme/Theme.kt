package pcmonitor.ui.theme

import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.darkColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color

/**
 * Paleta escura fixa.
 *
 * Um monitor de recursos costuma ficar aberto num canto da tela por horas, ao
 * lado de terminal e IDE. O tema escuro é escolha de produto, não preferência
 * do sistema — por isso não há alternância clara/escura.
 */
private val PcMonitorColors = darkColorScheme(
    primary = Color(0xFF4C8DFF),
    onPrimary = Color(0xFF08111F),
    secondary = Color(0xFF37C978),
    onSecondary = Color(0xFF04150C),
    tertiary = Color(0xFFFFB454),
    onTertiary = Color(0xFF1F1300),
    error = Color(0xFFFF5C5C),
    onError = Color(0xFF1F0505),
    background = Color(0xFF0F1115),
    onBackground = Color(0xFFE6E9EF),
    surface = Color(0xFF171A21),
    onSurface = Color(0xFFE6E9EF),
    surfaceVariant = Color(0xFF1E222B),
    onSurfaceVariant = Color(0xFF9AA3B2),
    outline = Color(0xFF2C3140),
)

/** Cores de estado usadas em medidores e no indicador de conexão. */
object StatusColors {
    val ok = Color(0xFF37C978)
    val warning = Color(0xFFFFB454)
    val critical = Color(0xFFFF5C5C)
    val idle = Color(0xFF6B7280)
}

/**
 * Devolve a cor de um medidor conforme o quanto ele está cheio.
 *
 * O limiar existe para que a tela responda antes do problema: 75 % de RAM já
 * merece atenção, 90 % já é aperto.
 */
fun gaugeColor(percent: Double): Color = when {
    percent >= 90 -> StatusColors.critical
    percent >= 75 -> StatusColors.warning
    else -> StatusColors.ok
}

@Composable
fun PcMonitorTheme(content: @Composable () -> Unit) {
    MaterialTheme(colorScheme = PcMonitorColors, content = content)
}
