# API do agente

Base: `http://127.0.0.1:8080` (configurável — ver [Configuração](#configuração))

Todas as respostas são `application/json`. Os exemplos abaixo são respostas
reais desta máquina, não amostras inventadas.

---

## Convenções

- **Bytes** são inteiros sem sinal.
- **Percentuais** são `float64` entre 0 e 100 — com uma exceção documentada:
  `cpuPercent` de processo é relativo a um núcleo e passa de 100 em processos
  multithread, como no `top`.
- **Listas vazias** serializam como `[]`, nunca `null`.
- Uma métrica indisponível é **omitida** do snapshot agregado em vez de
  aparecer zerada — a UI precisa distinguir "0 %" de "indisponível".

---

## `GET /health`

```json
{"status":"ok"}
```

Sempre 200 enquanto o processo estiver de pé. É o que o `scripts/start.sh`
consulta antes de abrir a janela.

---

## `GET /api/v1/system`

```json
{
  "hostname": "omarchy",
  "os": "linux",
  "osVersion": "4.0.1",
  "architecture": "amd64",
  "uptimeSeconds": 26210
}
```

Não muda enquanto a máquina estiver ligada; o desktop lê uma vez por sessão.

---

## `GET /api/v1/cpu`

```json
{
  "model": "AMD Ryzen 9 9950X3D 16-Core Processor",
  "cores": 16,
  "threads": 32,
  "usagePercent": 5.508464065611258
}
```

`usagePercent` é a média da máquina, derivada da diferença entre duas leituras
dos contadores acumulados. A primeira chamada depois de subir o agente mede
sobre o intervalo desde o boot.

---

## `GET /api/v1/memory`

```json
{
  "total": 32449835008,
  "used": 18034794496,
  "available": 14415040512,
  "usagePercent": 55.57746130775026
}
```

`available` é o que o sistema considera disponível — inclui cache recuperável,
e por isso `used + available` não bate com `total`.

---

## `GET /api/v1/disks`

Array, uma entrada por **dispositivo** (não por ponto de montagem):

```json
[
  {"name":"/dev/dm-0","mountPoint":"/","total":998037782528,"used":383127552000,"free":614170443776,"usagePercent":38.38808096318251},
  {"name":"/dev/nvme0n1p1","mountPoint":"/boot","total":2143281152,"used":581492736,"free":1561788416,"usagePercent":27.130959251770626}
]
```

Subvolumes btrfs e bind mounts expõem o mesmo dispositivo em vários pontos de
montagem, cada um reportando a capacidade inteira. O collector mantém a
montagem de caminho mais curto — sem isso, o resumo somaria 4 TB num disco de
1 TB.

---

## `GET /api/v1/network`

Array, uma entrada por interface (loopback excluído):

```json
[
  {
    "interfaceName": "enp13s0",
    "bytesReceived": 47165762933,
    "bytesSent": 849918053,
    "downloadBytesPerSecond": 0,
    "uploadBytesPerSecond": 0
  }
]
```

Os contadores são acumulados desde o boot; as velocidades vêm da diferença
entre duas leituras. **A primeira leitura devolve 0**: sem leitura anterior não
há intervalo sobre o qual calcular.

---

## `GET /api/v1/processes`

| Parâmetro | Valores | Padrão |
|---|---|---|
| `sort` | `cpu`, `memory` | `cpu` |
| `limit` | 1 a 500 | 50 |

```json
[
  {"pid":299528,"name":"chromium","cpuPercent":13.478507848702304,"memory":431620096},
  {"pid":111937,"name":"chromium","cpuPercent":8.511898085441027,"memory":794054656},
  {"pid":268910,"name":"java","cpuPercent":6.290804693504719,"memory":1669492736}
]
```

`memory` é a residente (RSS). A ordenação desempata por PID, para que
processos com o mesmo consumo não troquem de posição entre coletas.

Parâmetro fora do domínio devolve **400**:

```json
{"error":"invalid_request","message":"parâmetros inválidos: use sort=cpu|memory e limit entre 1 e 500"}
```

---

## `GET /api/v1/metrics`

Snapshot agregado — o endpoint principal do dashboard:

```json
{
  "timestamp": "2026-08-30T21:51:40.515532252Z",
  "cpu": {"model":"AMD Ryzen 9 9950X3D 16-Core Processor","cores":16,"threads":32,"usagePercent":4.738154614964954},
  "memory": {"total":32449835008,"used":18036494336,"available":14413340672,"usagePercent":55.582699670286104},
  "disk": {"total":1015740882944,"used":389937029120,"usagePercent":38.389419552535436},
  "network": {"downloadBytesPerSecond":4433.790003850863,"uploadBytesPerSecond":2364.6880020537938},
  "uptimeSeconds": 26210
}
```

`timestamp` é sempre UTC e sempre presente. Os demais campos são opcionais: a
falha de um collector omite **aquele** campo e registra um `WARN`, mas a
resposta continua 200 com o resto das métricas.

`disk` é o total do conjunto de dispositivos distintos; `network` é a soma das
interfaces.

---

## `GET /api/v1/ws/metrics` (WebSocket)

```
ws://127.0.0.1:8080/api/v1/ws/metrics
```

Envia um `DashboardMetrics` — o mesmo objeto acima — a cada
`PCMON_COLLECTION_INTERVAL` (1 s por padrão), mais um imediatamente ao
conectar.

Uma única coleta alimenta todos os clientes conectados. Cliente que não
consome fica com o snapshot mais recente e perde os intermediários, em vez de
atrasar os demais. O servidor fecha com `1001 going away` ao encerrar.

---

## Erros

| Situação | Status | `error` |
|---|---|---|
| Collector indisponível | 503 | `collector_unavailable` |
| Parâmetro fora do domínio | 400 | `invalid_request` |
| Rota inexistente | 404 | `not_found` |
| Método errado na rota | 405 | `method_not_allowed` |
| Falha inesperada | 500 | `internal_error` |

Formato:

```json
{"error":"collector_unavailable","message":"métrica temporariamente indisponível"}
```

O detalhe da causa fica no log do agente; a resposta traz apenas a descrição
genérica.

---

## Configuração

| Variável | Padrão | Limites |
|---|---|---|
| `PCMON_HOST` | `127.0.0.1` | não vazio |
| `PCMON_PORT` | `8080` | 1–65535 |
| `PCMON_COLLECTION_INTERVAL` | `1s` | 100 ms – 1 min |
| `PCMON_LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |

Valor inválido **impede a subida** em vez de cair no padrão em silêncio.

O agente escuta apenas em loopback por padrão: não há autenticação, e a API
expõe a lista de processos da máquina.
