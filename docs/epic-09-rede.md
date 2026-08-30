# Epic 09 — Rede

**Status:** concluído
**Data:** 2026-08-30

Cobre as Tasks 9.1 (collector), 9.2 (velocidade instantânea), 9.3
(`GET /api/v1/network`) e 9.4 (download/upload no snapshot agregado).

Primeiro epic com **estado entre chamadas**: velocidade não existe numa leitura
isolada, só na diferença entre duas.

---

## O que foi feito

| Arquivo | Papel |
|---|---|
| `internal/collector/network.go` | Collector gopsutil com memória da leitura anterior |
| `internal/service/network.go` | `NetworkService` com `List` e `Summary` |
| `internal/api/network.go` | Handler de `GET /api/v1/network` |
| `internal/service/metrics.go` | Passou a preencher `network` no snapshot |

---

## O problema central: o sistema só reporta contadores acumulados

`/proc/net/dev` — e o `net.IOCounters` do gopsutil sobre ele — devolve **bytes
desde o boot**, não velocidade:

```
enp13s0  bytesReceived=46464583473  bytesSent=631339751
```

Sozinho esse número não diz nada sobre o agora. A velocidade sai da diferença
entre duas leituras dividida pelo tempo decorrido, o que obriga o collector a
guardar a leitura anterior por interface (`map[string]counterSnapshot`,
protegido por mutex porque o servidor HTTP atende requisições concorrentes).

### O intervalo é medido, não assumido

Cada snapshot guarda o instante da leitura, e `speeds` divide pelo intervalo
**real**:

```go
elapsed := current.at.Sub(previous.at).Seconds()
```

Assumir "1 segundo" porque esse é o intervalo nominal de coleta erraria sempre
que o cliente chamasse fora de cadência — que é exatamente o que acontece: o
mesmo collector serve `/api/v1/network`, `/api/v1/metrics` e, mais adiante, o
loop do WebSocket.

### Reinício de contador não pode virar wraparound

`byteRate` compara **antes** de subtrair:

```go
if current < previous {
    return 0
}
```

Os contadores são `uint64`. Com a interface derrubada e levantada, ou o driver
recarregado, o contador reinicia e `current - previous` não produz um número
negativo detectável — produz algo da ordem de 1e19 por wraparound. A UI
mostraria um pico de 18 exabytes por segundo.

### Primeira leitura devolve zero

Sem leitura anterior não há intervalo. A alternativa — dividir o contador
acumulado pelo uptime — reportaria a **média desde o boot** como se fosse a
velocidade instantânea, um valor tão plausível quanto errado.

### Loopback fora

`lo` é tráfego interno da máquina. Incluí-lo faria qualquer comunicação entre
processos locais aparecer como uso de rede.

---

## Somar interfaces é seguro aqui (ao contrário dos discos)

O Epic 08 teve de deduplicar por dispositivo antes de somar. Aqui não:
cada interface tem contadores próprios, sem risco de contagem dupla.
`summarizeNetwork` soma direto, e `TestSummarizeNetwork` fixa esse
comportamento.

---

## Como os testes foram escritos

| Teste | O que prova |
|---|---|
| `TestByteRate` | taxa por intervalo, incluindo intervalo fracionário |
| `TestByteRateNuncaEstouraComWraparound` | contador quase no máximo de `uint64` reiniciando devolve 0, não 1e19 |
| `TestSpeeds` | primeira leitura = 0; intervalo real e não nominal; duas leituras no mesmo instante = 0 |
| `TestNetworkCollectorMaquinaReal` | interfaces reais, loopback ausente, nada negativo/NaN/Inf |
| `TestNetworkCollectorCalculaEntreColetas` | caminho completo do collector com relógio injetado |
| `TestNetworkServiceListConverteErro` | falha vira `ErrUnavailable` sem perder a causa |
| `TestSummarizeNetwork` | soma entre interfaces e lista vazia |

O relógio é injetável (`now func() time.Time`) justamente para que o teste do
delta não dependa de tráfego real nem de `time.Sleep`.

NaN e Inf são verificados explicitamente: ambos serializam como JSON inválido e
quebrariam a desserialização no cliente KMP.

---

## Validação de ponta a ponta

Primeira chamada a `GET /api/v1/network`, sem baseline:

```json
[{"interfaceName":"enp13s0","bytesReceived":46464583473,"bytesSent":631339751,
  "downloadBytesPerSecond":0,"uploadBytesPerSecond":0}]
```

20 MB baixados entre as duas chamadas:

```json
[{"interfaceName":"enp13s0","bytesReceived":46485973826,"bytesSent":631681580,
  "downloadBytesPerSecond":2233281.08,"uploadBytesPerSecond":35688.99}]
```

21 390 353 bytes de diferença sobre a janela decorrida — e não sobre os 0,35 s
que o download levou. **A velocidade é a média do intervalo entre coletas**,
não o pico de transferência; com o WebSocket coletando a cada segundo (Epic 15)
a janela fica curta e o número se aproxima do instantâneo.

Snapshot agregado com rede presente:

```json
{"timestamp":"2026-08-30T21:01:05Z",
 "cpu":{"model":"AMD Ryzen 9 9950X3D 16-Core Processor","cores":16,"threads":32,"usagePercent":5.97},
 "memory":{"total":32449835008,"usagePercent":40.33},
 "disk":{"total":1015740882944,"usagePercent":38.34},
 "network":{"downloadBytesPerSecond":71471.76,"uploadBytesPerSecond":0},
 "uptimeSeconds":23176}
```

Falha do collector de rede continua degradando o snapshot em vez de derrubá-lo:
`WARN` no log e o campo `network` omitido do JSON.
