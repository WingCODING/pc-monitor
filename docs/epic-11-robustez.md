# Epic 11 — Robustez do Backend

**Status:** concluído
**Data:** 2026-08-30

Cobre as Tasks 11.1 (respostas de erro padronizadas), 11.2 (logging) e 11.3
(configuração).

Nenhuma métrica nova: o epic fecha as bordas do agente antes de o desktop
começar a depender dele.

---

## O que foi feito

| Arquivo | Papel |
|---|---|
| `internal/api/middleware.go` | Padronização de erros, recuperação de panic e log de requisição |
| `internal/service/logthrottle.go` | `warnThrottle`, um WARN por falha a cada 30 s |
| `internal/config/config.go` | `PCMON_HOST`, `PCMON_PORT`, `PCMON_COLLECTION_INTERVAL`, `PCMON_LOG_LEVEL` |
| `cmd/server/main.go` | Passa a subir a partir da configuração |

---

## 11.1 — Os erros que não passavam por `writeError`

O formato já existia desde a Epic 04 (`{"error":…,"message":…}`), mas só valia
para os erros que os handlers produziam. Dois casos escapavam, ambos gerados
pelo próprio `ServeMux`:

```
GET /api/v1/inexistente  → 404 text/plain  "404 page not found"
POST /health             → 405 text/plain  "Method Not Allowed"
```

Um cliente que sempre desserializa JSON quebra nos dois.

A correção é um `ResponseWriter` instrumentado que, ao ver um status ≥ 400 com
`Content-Type` diferente de `application/json`, substitui o corpo:

```json
{"error":"not_found","message":"recurso não encontrado"}
{"error":"method_not_allowed","message":"método não permitido para este recurso"}
```

A alternativa seria registrar uma rota `/` de captura para o 404 — mas isso não
resolve o 405, que o `ServeMux` gera internamente quando o caminho casa e o
método não.

### Panic vira 500, não conexão abortada

Sem `recover`, um panic em handler faz o servidor HTTP **abortar a conexão sem
resposta**. O cliente vê erro de rede, e o dashboard classificaria isso como
"agente fora do ar" em vez de "erro da API" — dois estados de UI diferentes,
com reações diferentes.

O middleware registra o panic com stack em `ERROR` e devolve o 500 padrão. O
stack fica no log, nunca na resposta. Dois detalhes:

- `http.ErrAbortHandler` é repassado: é o sinal documentado para abortar em
  silêncio.
- Se o handler **já respondeu**, nada é escrito por cima — sobrescrever ali só
  produziria um corpo inválido.

### Códigos usados

| Situação | Status | `error` |
|---|---|---|
| Collector indisponível | 503 | `collector_unavailable` |
| Parâmetro fora do domínio (Epic 10) | 400 | `invalid_request` |
| Rota inexistente | 404 | `not_found` |
| Método errado | 405 | `method_not_allowed` |
| Panic ou falha inesperada | 500 | `internal_error` |

---

## 11.2 — Logging que sobrevive a 1 requisição por segundo

O requisito "logs não são excessivos a cada segundo" é o que dita o desenho.
Duas fontes de repetição:

### Uma linha por requisição

O dashboard vai consultar o agente continuamente. O log de requisição existe —
método, caminho, status, duração — mas em **DEBUG**:

```
level=DEBUG msg="requisição atendida" method=GET path=/health status=200 durationMs=0
```

Em INFO, o padrão, o log fica com o que interessa: subida do servidor,
desligamento, falhas.

### Uma linha por falha de collector, por segundo

Um collector quebrado com o loop do WebSocket publicando a cada segundo daria
3 600 linhas idênticas por hora. `warnThrottle` registra a primeira ocorrência
na hora e cala as repetições por 30 s, informando quantas foram omitidas ao
voltar a registrar:

```
level=WARN msg="cpu indisponível no snapshot" error="..." repeticoesOmitidas=29
```

A contagem não é enfeite: sem ela, o log de uma falha contínua fica idêntico ao
de uma falha esporádica.

O throttle é por chave (`cpu`, `memory`, `disk`, `network`, `system`), então uma
falha de disco não silencia uma falha de CPU.

---

## 11.3 — Configuração por ambiente

| Variável | Padrão | Limites |
|---|---|---|
| `PCMON_HOST` | `127.0.0.1` | não vazio |
| `PCMON_PORT` | `8080` | 1–65535 |
| `PCMON_COLLECTION_INTERVAL` | `1s` | 100 ms – 1 min |
| `PCMON_LOG_LEVEL` | `info` | debug, info, warn, error |

Sem nenhuma variável exportada, o agente sobe exatamente como antes — os
padrões do MVP são os da spec.

**Valor inválido é erro, não motivo para cair no padrão.** Quem exporta
`PCMON_PORT=oitenta` quer aquela porta; subir na 8080 esconderia o engano até
alguém estranhar:

```
ERROR agente encerrado com erro error="configuração inválida: PCMON_PORT=\"oitenta\" inválido: use um inteiro entre 1 e 65535"
```

O intervalo de coleta tem limites em vez de aceitar qualquer duração: abaixo de
100 ms a varredura de processos passaria o tempo todo lendo `/proc`, e acima de
1 minuto o "tempo real" da UI deixa de merecer o nome.

---

## Como os testes foram escritos

| Teste | O que prova |
|---|---|
| `TestRotaInexistenteRespondeJSON` | 404 do ServeMux convertido |
| `TestMetodoNaoPermitidoRespondeJSON` | 405 convertido |
| `TestRespostaDeSucessoPassaIntacta` | a conversão não toca o caminho feliz |
| `TestPanicViraErroPadrao` | 500 no formato padrão, sem stack no corpo |
| `TestPanicDepoisDeResponderNaoCorrompeOCorpo` | resposta já iniciada é preservada |
| `TestWarnThrottleSilenciaRepeticoes` | 20 falhas em 20 s → 1 linha |
| `TestWarnThrottleInformaAsOmitidas` | 40 falhas → 2 linhas, `repeticoesOmitidas=29` |
| `TestWarnThrottleSeparaChaves` | disco não silencia CPU |
| `TestLoadRejeitaValorInvalido` | 8 formas de configuração inválida |
| `TestLoadLeOAmbiente` / `TestDefaultAtendeOsPadroesDoMVP` | ambiente e padrões |

O relógio do throttle é injetável, então os testes de repetição rodam em
microssegundos em vez de esperar 30 s reais.

---

## Validação de ponta a ponta

```
$ PCMON_PORT=8099 PCMON_LOG_LEVEL=debug PCMON_COLLECTION_INTERVAL=500ms ./agent
INFO msg="servidor iniciado" addr=127.0.0.1:8099 collectionInterval=500ms logLevel=DEBUG

$ curl -i localhost:8099/api/v1/inexistente
HTTP/1.1 404 Not Found
Content-Type: application/json
{"error":"not_found","message":"recurso não encontrado"}

$ curl -X POST localhost:8099/health
{"error":"method_not_allowed","message":"método não permitido para este recurso"}   [405]

$ curl localhost:8080/health          → conexão recusada (a porta mudou de fato)
```

O `Content-Length` também é corrigido na substituição: o `ServeMux` já tinha
anunciado o tamanho do corpo em texto puro.
