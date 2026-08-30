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

/**
 * Estado da ligação com o agente, do ponto de vista da interface.
 *
 * São três e não dois porque a UI reage de forma diferente a cada um: durante
 * a conexão inicial não há o que mostrar, enquanto uma queda depois de
 * conectado mantém os últimos valores na tela em vez de esvaziá-la.
 */
enum class ConnectionState {
    Connecting,
    Connected,
    Disconnected,
}

/** Etiqueta com bolinha colorida indicando a ligação com o agente. */
@Composable
fun ConnectionStatus(state: ConnectionState, modifier: Modifier = Modifier) {
    val (color, label) = when (state) {
        ConnectionState.Connecting -> StatusColors.warning to "Conectando"
        ConnectionState.Connected -> StatusColors.ok to "Conectado"
        ConnectionState.Disconnected -> StatusColors.critical to "Desconectado"
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
