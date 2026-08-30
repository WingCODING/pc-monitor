package pcmonitor.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import pcmonitor.ui.theme.StatusColors
import pcmonitor.viewmodel.ConnectionState

/**
 * Etiqueta com bolinha colorida indicando a ligação com o agente.
 *
 * O enum vive no `viewmodel` e não aqui: o estado da ligação é estado da
 * aplicação, e deixá-lo na camada de componentes obrigaria o ViewModel a
 * depender da UI para descrever a si mesmo.
 */
@Composable
fun ConnectionStatus(state: ConnectionState, modifier: Modifier = Modifier) {
    val (color, label) = when (state) {
        ConnectionState.Loading -> StatusColors.idle to "Conectando"
        ConnectionState.Connected -> StatusColors.ok to "Conectado"
        ConnectionState.Disconnected -> StatusColors.critical to "Desconectado"
        ConnectionState.Error -> StatusColors.warning to "Erro no agente"
    }

    Row(
        modifier = modifier
            .background(MaterialTheme.colorScheme.surfaceVariant, RoundedCornerShape(999.dp))
            .padding(horizontal = 12.dp, vertical = 6.dp),
        horizontalArrangement = Arrangement.spacedBy(8.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Dot(color)

        Text(
            text = label,
            style = MaterialTheme.typography.labelLarge,
            color = MaterialTheme.colorScheme.onSurface,
        )
    }
}

@Composable
private fun Dot(color: Color) {
    Row(
        modifier = Modifier
            .size(8.dp)
            .background(color, CircleShape),
    ) {}
}
