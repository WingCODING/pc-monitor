package pcmonitor.ui.format

import kotlin.math.abs
import kotlin.math.roundToLong

/**
 * Formatação para exibição.
 *
 * Escrito sem `String.format`: a função não existe no source set comum do
 * Kotlin Multiplatform, e manter os formatadores aqui evita que a UI dependa
 * de APIs de uma plataforma específica.
 *
 * A vírgula decimal é fixa em vez de vir do `Locale` — o resto da interface
 * está em português, e uma vírgula que muda com a máquina só produziria
 * capturas de tela inconsistentes.
 */

private const val UNIT_STEP = 1024.0
private val BYTE_UNITS = listOf("B", "KB", "MB", "GB", "TB", "PB")

/** Formata uma quantidade de bytes em unidade binária: 1024 vira "1 KB". */
fun formatBytes(bytes: Long): String = formatBytes(bytes.toDouble())

fun formatBytes(bytes: Double): String {
    if (bytes < UNIT_STEP) {
        return "${bytes.roundToLong()} B"
    }

    var value = bytes
    var unit = 0

    while (value >= UNIT_STEP && unit < BYTE_UNITS.lastIndex) {
        value /= UNIT_STEP
        unit++
    }

    return "${formatDecimal(value, 1)} ${BYTE_UNITS[unit]}"
}

/** Formata uma velocidade de transferência. */
fun formatBytesPerSecond(bytesPerSecond: Double): String = "${formatBytes(bytesPerSecond)}/s"

/** Formata um tempo de atividade: 532 140 s vira "6 d 3 h 49 min". */
fun formatUptime(seconds: Long): String {
    if (seconds < 0) {
        return "—"
    }

    val days = seconds / 86_400
    val hours = (seconds % 86_400) / 3_600
    val minutes = (seconds % 3_600) / 60

    // Minutos aparecem sempre: um agente recém-iniciado mostraria "0 h" sozinho
    // e pareceria travado.
    return when {
        days > 0 -> "$days d $hours h $minutes min"
        hours > 0 -> "$hours h $minutes min"
        else -> "$minutes min"
    }
}

/** Formata um percentual com uma casa: 45.27 vira "45,3 %". */
fun formatPercent(percent: Double): String = "${formatDecimal(percent, 1)} %"

/**
 * Arredonda para o número de casas pedido e monta a string.
 *
 * A casa decimal some quando é zero: "1 KB" lê melhor que "1,0 KB", e o valor
 * exibido não perde informação.
 */
fun formatDecimal(value: Double, decimals: Int): String {
    if (decimals <= 0) {
        return value.roundToLong().toString()
    }

    var scale = 1L
    repeat(decimals) { scale *= 10 }

    val scaled = (value * scale).roundToLong()
    val whole = scaled / scale
    val fraction = abs(scaled % scale)

    if (fraction == 0L) {
        return whole.toString()
    }

    // Zeros à esquerda da parte fracionária: 1,05 tem fração 5 com escala 100.
    val digits = fraction.toString().padStart(decimals, '0').trimEnd('0')

    // O sinal vive na parte inteira, exceto quando ela arredonda para zero.
    val sign = if (whole == 0L && value < 0) "-" else ""

    return "$sign$whole,$digits"
}
