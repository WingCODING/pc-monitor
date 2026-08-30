# Epic 06 — Endpoint Agregado

**Status:** concluído
**Data:** 2026-08-30

Cobre as Tasks 6.1 (Metrics Service) e 6.2 (`GET /api/v1/metrics`).

Este é o endpoint que o dashboard vai consumir de fato, e o mesmo payload que
o WebSocket publicará no Epic 15.

---

## O que foi feito

| Arquivo | Papel |
|---|---|
| `internal/service/metrics.go` | `MetricsService`, monta o snapshot |
| `internal/api/metrics.go` | Handler de `GET /api/v1/metrics` |

---

## Decisões e por quê

### 1. `Snapshot` não devolve erro

A assinatura é `Snapshot(ctx) model.DashboardMetrics` — sem `error`.

Isso é deliberado e é a decisão central do epic. A falha de um collector
individual **omite aquela métrica** e mantém as demais. Um problema no
`/proc/stat` não pode apagar a leitura de memória da tela.

O critério "falha de uma métrica não causa crash" (Task 6.1) fica satisfeito
por construção: não existe caminho de código em que uma falha isolada aborte o
snapshot.

### 2. O endpoint responde 200 mesmo com métricas faltando

Consequência da decisão anterior. Um snapshot parcial é resposta **legítima**,
não erro. O cliente distingue "indisponível" de "zero" pela **ausência do
campo** no JSON — que é exatamente o que os ponteiros com `omitempty` do
Epic 03 viabilizam.

Isso contrasta de propósito com `/api/v1/cpu`, que devolve **503** quando a CPU
falha. A diferença faz sentido:

- endpoint de recurso único: sem o dado, não há resposta útil → 503;
- endpoint agregado: ainda há dado útil a entregar → 200 parcial.

### 3. O service compõe outros services, não os collectors

`NewMetricsService(cpu, memory)` recebe `*CPUService` e `*MemoryService`, e não
os collectors diretamente. Assim a classificação de erro (`ErrUnavailable`) e
qualquer regra futura de cada service são reaproveitadas, em vez de duplicadas.

### 4. Timestamp em UTC

`time.Now().UTC()`. O agente e a UI podem estar em fusos diferentes — no futuro
mais ainda, se o monitoramento remoto sair do "fora do MVP". UTC torna o
timestamp comparável sem ambiguidade.

### 5. Coleta sequencial

Ler CPU e memória do `/proc` custa microssegundos. Paralelizar traria
concorrência sem ganho mensurável. Está registrado em comentário no código
para que a escolha não pareça descuido.

---

## Sobre o uptime

A Task 6.1 lista "CPU, RAM e uptime", mas **uptime não entrou neste epic**.

Motivo: uptime vem do System Collector, que é a Task 7.1 — e a Task 7.3
("Integrar System ao Metrics Service") existe justamente para conectá-lo ao
snapshot. Implementar uma leitura de uptime provisória aqui seria código
descartado no epic seguinte.

O campo já existe em `DashboardMetrics` como ponteiro, então hoje é
simplesmente omitido do JSON. O critério 6.2 diz "uptime presente **quando
disponível**", o que essa situação satisfaz. O Epic 07 o preenche.

---

## Erros encontrados durante a implementação

Nenhum. O epic reaproveitou infraestrutura já testada dos Epics 03 a 05, e os
ponteiros com `omitempty` definidos no Epic 03 se encaixaram sem ajuste — o
que era exatamente a intenção de fixar os contratos antes de implementar os
collectors.

---

## Como os testes foram escritos

O foco foi o comportamento sob falha parcial, que é o que o epic introduz.

### Três cenários no service

| Teste | O que prova |
|---|---|
| `TestMetricsServiceSnapshotCompleto` | CPU, memória e timestamp presentes |
| `TestMetricsServiceSnapshotParcial` | CPU falha → memória **sobrevive** |
| `TestMetricsServiceTodasAsMetricasFalham` | Snapshot ainda é produzido, com timestamp |

O segundo é o teste central. Ele não verifica apenas que a CPU foi omitida —
verifica que a memória **continua correta**, que é a propriedade que
importa. Um bug que abortasse o snapshot na primeira falha passaria por um
teste que só olhasse o campo ausente.

O terceiro cobre o caso extremo: mesmo sem nenhuma métrica, o timestamp
continua sendo publicado e o agente não quebra.

### Timestamp verificado por janela

Em vez de comparar com um valor fixo, o teste captura `antes` e `depois` da
chamada e verifica que o timestamp caiu **dentro da janela**:

```go
if snapshot.Timestamp.Before(antes) || snapshot.Timestamp.After(depois) {
```

Isso testa que o timestamp é gerado no momento da coleta, sem depender de
relógio mockado nem tolerar um valor arbitrário.

### Handler

Dois casos, espelhando as decisões 1 e 2: resposta completa com todos os
campos, e resposta parcial que **ainda é 200** com o campo indisponível
ausente do JSON.

Os testes desserializam para `map[string]json.RawMessage` em vez do struct.
Desserializar para o struct não distinguiria "campo ausente" de "campo com
valor zero" — justamente a distinção que o epic precisa garantir.

---

## Validação de ponta a ponta

```json
{
    "timestamp": "2026-08-30T18:13:16.420543086Z",
    "cpu": {
        "model": "AMD Ryzen 9 9950X3D 16-Core Processor",
        "cores": 16,
        "threads": 32,
        "usagePercent": 2.363222468565731
    },
    "memory": {
        "total": 32449835008,
        "used": 14909251584,
        "available": 17540583424,
        "usagePercent": 45.94553895366296
    }
}
```

Confere com o esperado:

- `disk`, `network` e `uptimeSeconds` **ausentes** — ainda não implementados,
  e omitidos em vez de aparecerem zerados;
- duas chamadas consecutivas devolveram timestamps 1,01 s distantes,
  confirmando que o snapshot é gerado a cada requisição e não cacheado.
