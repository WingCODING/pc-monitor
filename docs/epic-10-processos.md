# Epic 10 — Processos

**Status:** concluído
**Data:** 2026-08-30

Cobre as Tasks 10.1 (collector), 10.2 (ordenação e limite) e 10.3
(`GET /api/v1/processes`).

O collector mais caro do agente — uma varredura de centenas de processos por
requisição — e o primeiro endpoint com parâmetros de consulta.

---

## O que foi feito

| Arquivo | Papel |
|---|---|
| `internal/collector/process.go` | Varredura de `/proc` com uso de CPU derivado entre coletas |
| `internal/service/process.go` | `ProcessService` com ordenação e limite |
| `internal/api/process.go` | Handler de `GET /api/v1/processes` e validação da query |
| `internal/service/errors.go` | `ErrInvalidArgument`, traduzido em 400 |

---

## CPU por processo também é um delta

O sistema reporta **tempo de CPU acumulado** (`utime + stime`), não percentual.
A regra é a mesma da rede: guardar a leitura anterior e dividir a diferença
pelo intervalo real.

Três diferenças em relação à rede:

### 1. O PID é reaproveitado

O número de processo é reciclado pelo sistema. Sem checagem, o tempo de CPU de
um processo morto seria comparado com o de outro recém-criado no mesmo PID.
Cada snapshot guarda `createdAt`; quando ele muda, a entrada é tratada como um
processo novo.

### 2. Primeira leitura não pode ser zero

Na rede, a primeira leitura devolve 0 e pronto. Aqui isso zeraria a lista
inteira na primeira requisição — e a ordenação por CPU não teria o que ordenar.
Sem base anterior, o collector usa a **média desde a criação do processo**
(`cpuSeconds / idade`), que é o que o `CPUPercent` do gopsutil faz.

### 3. Janela mínima de amostragem

O tempo de CPU em `/proc` avança em ticks de 10 ms. Duas chamadas em sequência
caem dentro do mesmo tick, e a diferença dá zero. A primeira versão do endpoint
mostrou exatamente isso:

```
GET /processes?limit=5          → chromium 17,5 %, chromium 9,1 %, ...
GET /processes?sort=memory      → idea 0 %, chromium 0 %, chromium 0 %
```

Todos os percentuais zerados na segunda chamada, feita alguns milissegundos
depois. Pior: com o tick caindo *dentro* da janela, o mesmo cálculo produziria
um pico de 50 % igualmente falso.

A correção é uma janela mínima de 500 ms. Abaixo dela o collector **repete a
última medição válida e mantém a base anterior** — a próxima coleta então mede
sobre um intervalo utilizável em vez de recomeçar de uma base recém-gravada.

Depois da correção:

```
GET /processes?limit=3      → chromium 15,05 %, chromium 9,15 %, chromium 6,60 %
GET /processes?sort=memory  → idea 4,51 %, chromium 9,15 %, chromium 15,05 %
```

### O percentual passa de 100 de propósito

`cpuPercent` é relativo a **um núcleo**, como no `top`: um processo com quatro
threads saturadas mostra 400. Limitar a 100 esconderia justamente os processos
que mais pesam. O `usagePercent` da Epic 04 é outra coisa — média da máquina
inteira — e continua em 0–100.

---

## Processos que morrem durante a coleta

Entre listar os PIDs e ler os dados de cada um, alguns processos terminam:
`/proc/<pid>/*` desaparece e as leituras falham. O collector simplesmente
**ignora** esses processos, sem log — isso é rotina em qualquer máquina, e
registrar cada ocorrência encheria a saída a cada coleta.

## O estado interno não pode vazar

O mapa de snapshots é **substituído** a cada coleta, não atualizado. Manter
entradas de processos mortos seria um vazamento lento: uma máquina de
desenvolvimento cria e destrói milhares de processos por dia.
`TestProcessCollectorDescartaProcessosMortos` planta um PID inexistente no
estado e verifica que ele some.

## Cancelamento

A varredura verifica `ctx.Err()` a cada processo. Se o cliente desiste da
requisição no meio, não há motivo para terminar de percorrer centenas de
entradas em `/proc`.

---

## Ordenação e limite

Ficam no **service**, não no collector: o collector precisa ler todos os
processos de qualquer forma para manter os deltas corretos.

- **Ordenação padrão:** CPU decrescente — o que um monitor mostra ao abrir.
- **Desempate por PID.** Sem ele, as dezenas de daemons parados em 0 %
  trocariam de posição entre coletas e a tabela da UI ficaria piscando.
- **Limite padrão 50**, máximo 500. Esta máquina tem 662 processos; a cauda é
  quase toda de daemons ociosos.

### Query string validada

`GET /api/v1/processes?sort=cpu|memory&limit=1..500`

Parâmetro fora do domínio é **400**, não 503: é erro do cliente, e repetir a
mesma requisição não vai funcionar. Isso motivou `ErrInvalidArgument` no
service, mapeado em `writeError`:

```json
{"error":"invalid_request","message":"parâmetros inválidos: use sort=cpu|memory e limit entre 1 e 500"}
```

A mensagem descreve o domínio aceito em vez de ecoar o valor recebido — ecoar
entrada do cliente na resposta é um vetor de injeção gratuito, e o log já
guarda o valor exato.

---

## Como os testes foram escritos

| Teste | O que prova |
|---|---|
| `TestSampleCPU` | média na primeira observação, delta, >100 em multithread, PID reaproveitado, tempo regredido, janela curta |
| `TestSampleCPUPreservaBaseEmJanelaCurta` | a base antiga sobrevive à janela curta e a medição seguinte é correta |
| `TestProcessCollectorMaquinaReal` | processos reais, PID > 0, nome presente, nada NaN/Inf, o próprio teste aparece na lista |
| `TestProcessCollectorNaoBloqueia` | a varredura cabe numa requisição HTTP |
| `TestProcessCollectorDescartaProcessosMortos` | o estado interno não cresce |
| `TestProcessCollectorRespeitaCancelamento` | contexto cancelado interrompe a varredura |
| `TestSortProcessesPorCPU` / `PorMemoria` | os dois critérios |
| `TestSortProcessesDesempateEstavel` | duas ordens de entrada diferentes produzem a mesma saída |
| `TestLimitProcesses` | limite menor, maior, igual, zero e negativo |
| `TestHandleProcesses*` | ordenação padrão, `sort=memory`, `limit`, limite padrão com 300 processos, 5 formas de parâmetro inválido, lista vazia como `[]`, 503 |

---

## Validação de ponta a ponta

```
$ curl 'http://127.0.0.1:8080/api/v1/processes?limit=5'  → 0,037 s
[{"pid":250015,"name":"chromium","cpuPercent":17.47,"memory":690499584}, ...]

$ curl 'http://127.0.0.1:8080/api/v1/processes?sort=memory&limit=3'
[{"pid":113512,"name":"idea","cpuPercent":4.51,"memory":2381553664}, ...]

$ curl 'http://127.0.0.1:8080/api/v1/processes?sort=disco'          → 400
$ curl 'http://127.0.0.1:8080/api/v1/processes?limit=500' | len     → 500 de 662
```

37 ms para varrer 662 processos — dentro do orçamento de uma requisição HTTP, e
folgado em relação ao intervalo de 1 s do WebSocket.
