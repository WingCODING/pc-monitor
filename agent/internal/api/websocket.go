package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"pc-monitor-agent/internal/service"
)

// socketWriteTimeout limita quanto tempo uma escrita pode ficar pendurada.
//
// Sem prazo, um cliente que parou de ler mantém a goroutine presa até o TCP
// desistir — o que pode levar minutos.
const socketWriteTimeout = 5 * time.Second

// handleMetricsSocket serve GET /api/v1/ws/metrics.
//
// O handler não coleta nada: ele só se inscreve no hub e repassa. É o que
// garante uma coleta por intervalo, independentemente de quantas janelas
// estejam abertas.
func handleMetricsSocket(hub *service.MetricsHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if hub == nil {
			writeError(w, r, errors.New("hub de métricas não configurado"))

			return
		}

		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			// Accept já respondeu ao cliente com o status adequado.
			slog.Warn("upgrade para WebSocket recusado", "remote", r.RemoteAddr, "error", err)

			return
		}

		defer conn.CloseNow()

		// CloseRead descarta o que o cliente enviar e cancela o contexto assim
		// que a conexão cair. Sem ele, um cliente que some deixaria esta
		// goroutine publicando para um socket morto até a primeira escrita
		// falhar — e o protocolo exige alguém lendo para responder aos pings.
		ctx := conn.CloseRead(r.Context())

		snapshots, unsubscribe := hub.Subscribe()
		defer unsubscribe()

		slog.Info("cliente conectado ao WebSocket", "remote", r.RemoteAddr)
		defer slog.Info("cliente desconectado do WebSocket", "remote", r.RemoteAddr)

		for {
			select {
			case <-ctx.Done():
				return

			case snapshot, open := <-snapshots:
				if !open {
					// O hub encerrou: fecha com status de saída em vez de
					// derrubar a conexão, para o cliente saber que foi o
					// servidor que saiu.
					_ = conn.Close(websocket.StatusGoingAway, "agente encerrando")

					return
				}

				writeCtx, cancel := context.WithTimeout(ctx, socketWriteTimeout)
				err := wsjson.Write(writeCtx, conn, snapshot)

				cancel()

				if err != nil {
					slog.Debug("escrita no WebSocket falhou", "remote", r.RemoteAddr, "error", err)

					return
				}
			}
		}
	}
}
