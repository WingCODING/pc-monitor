#!/usr/bin/env bash
#
# Sobe o agente e o desktop juntos.
#
# O agente vira filho deste script: quando a janela fecha, ele encerra junto.
# Sem isso ficaria um processo escutando na porta sem ninguém para usá-lo — e
# a próxima execução falharia com "porta em uso".

set -euo pipefail

raiz="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

agente="${PCMON_AGENT_BIN:-$raiz/dist/pc-monitor-agent}"
desktop="${PCMON_DESKTOP_BIN:-$raiz/dist/desktop/bin/pc-monitor}"

for binario in "$agente" "$desktop"; do
    if [[ ! -x "$binario" ]]; then
        echo "não encontrei $binario — rode scripts/build.sh primeiro" >&2
        exit 1
    fi
done

host="${PCMON_HOST:-127.0.0.1}"
port="${PCMON_PORT:-8080}"

"$agente" &
agente_pid=$!

# Um processo que já morreu continua na tabela como zumbi até o pai recolhê-lo,
# e nesse estado `kill -0` ainda responde que ele existe. Sem olhar o estado, a
# checagem abaixo daria o agente como vivo justamente no caso que ela existe
# para pegar: o de ele ter morrido ao subir.
agente_vivo() {
    local estado
    estado="$(ps -o stat= -p "$agente_pid" 2>/dev/null || true)"

    [[ -n "$estado" && "$estado" != Z* ]]
}

trap 'kill "$agente_pid" 2>/dev/null || true' EXIT INT TERM

# Espera o agente atender antes de abrir a janela. O desktop sobreviveria sem
# isso — ele reconecta sozinho —, mas o motivo mais comum de falha é a porta
# ocupada, e vale dizer isso no terminal em vez de deixar a janela mostrando
# "Desconectado" sem explicação.
pronto=0

for _ in $(seq 1 25); do
    if ! agente_vivo; then
        echo "o agente encerrou ao iniciar; verifique se a porta $port está livre" >&2
        exit 1
    fi

    if curl -sf -m 1 "http://$host:$port/health" >/dev/null 2>&1; then
        pronto=1
        break
    fi

    sleep 0.2
done

if [[ "$pronto" -eq 0 ]]; then
    echo "o agente ainda não respondeu em http://$host:$port — abrindo mesmo assim" >&2
fi

PCMON_AGENT_URL="http://$host:$port" "$desktop"
