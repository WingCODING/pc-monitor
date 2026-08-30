package pcmonitor.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import pcmonitor.ui.theme.StatusColors
import pcmonitor.viewmodel.ConnectionState

/**
 * Faixa de estado acima do dashboard.
 *
 * Só aparece quando há algo a dizer: com tudo funcionando, a etiqueta do
 * cabeçalho já basta, e uma faixa permanente roubaria altura da tabela para
 * repetir "está tudo bem".
 *
 * Os dados continuam na tela por trás dela — a última leitura boa é o que o
 * usuário quer ver enquanto o agente volta.
 */
@Composable
fun StatusBanner(connection: ConnectionState, message: String?, modifier: Modifier = Modifier) {
    val (color, text) = when (connection) {
        ConnectionState.Connected -> return

        ConnectionState.Loading -> StatusColors.idle to (message ?: "Conectando ao agente…")

        ConnectionState.Disconnected -> StatusColors.critical to
            (message ?: "Agente indisponível. Tentando novamente…")

        ConnectionState.Error -> StatusColors.warning to
            (message ?: "O agente respondeu com erro.")
    }

    Row(
        modifier = modifier
            .fillMaxWidth()
            .background(color.copy(alpha = 0.12f), RoundedCornerShape(10.dp))
            .padding(horizontal = 14.dp, vertical = 10.dp),
        horizontalArrangement = Arrangement.spacedBy(10.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Box(modifier = Modifier.size(8.dp).background(color, CircleShape))

        Text(
            text = text,
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurface,
        )

        // Enquanto a conexão inicial acontece, dizer onde o agente é esperado
        // poupa a pergunta seguinte.
        if (connection == ConnectionState.Loading) {
            Text(
                text = "127.0.0.1:8080",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
        }
    }
}
