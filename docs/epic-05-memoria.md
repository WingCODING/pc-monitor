# Epic 05 — Memória RAM

**Status:** concluído
**Data:** 2026-08-30

Cobre as Tasks 5.1 (collector), 5.2 (service) e 5.3 (endpoint
`GET /api/v1/memory`).

---

## O que foi feito

| Arquivo | Papel |
|---|---|
| `internal/collector/memory.go` | Collector gopsutil, sem estado |
| `internal/service/memory.go` | `MemoryService` |
| `internal/api/memory.go` | Handler de `GET /api/v1/memory` |

O epic reaproveitou toda a infraestrutura montada no Epic 04 — `percentOf`,
`ErrUnavailable`, `writeError` e `writeJSON` — então o volume de código novo
foi pequeno. Isso era o objetivo do desenho do Epic 04.

---

## Decisões e por quê

### 1. O collector não guarda estado

Diferente do de CPU, memória é uma leitura **instantânea**, não um contador
acumulado. Não existe delta a calcular, então `memoryCollector` é uma struct
vazia. Vale registrar o contraste: o de CPU precisa de mutex e baseline, o de
memória não precisa de nada.

### 2. O percentual é recalculado, não copiado do gopsutil

`VirtualMemoryStat` já traz `UsedPercent` pronto. Mesmo assim, o collector
recalcula com `percentOf(stat.Used, stat.Total)`.

O motivo é coerência interna: a resposta publica `used`, `total` e
`usagePercent` lado a lado. Se o percentual viesse de um cálculo do gopsutil e
os bytes de outro campo, um cliente que dividisse `used/total` poderia obter um
número diferente do `usagePercent` publicado. Recalculando, os três campos são
sempre consistentes entre si — e há teste verificando exatamente isso.

### 3. `Total == 0` é erro, não valor válido

Uma leitura com total zerado significa que o `/proc/meminfo` não pôde ser
interpretado. Publicar `0` levaria a UI a mostrar "0 B de RAM" como se fosse
verdade. O collector devolve erro, que vira 503 — a UI então mostra
"indisponível", que é honesto.

Isso também protege `percentOf` do caso de divisão por zero, embora a função já
o trate.

---

## Erros encontrados durante a implementação

### Diretório de trabalho perdido entre comandos

O `go mod tidy` falhou com `go.mod file not found`. A causa não foi o projeto:
um `pkill` de um comando anterior encerrou com código 144 e o shell teve o
diretório de trabalho resetado para a raiz do projeto, fora do módulo Go.
Resolvido rodando os comandos com o caminho absoluto do módulo.

Nenhum problema de código — vale a nota apenas para não confundir esse sintoma
com um `go.mod` corrompido no futuro.

---

## Como os testes foram escritos

A estrutura seguiu a do Epic 04, com um teste a mais que é específico deste
epic.

### Coerência entre percentual e bytes

`TestMemoryCollectorMaquinaReal` verifica os invariantes usuais
(`Total > 0`, `Used <= Total`, `Available <= Total`, percentual entre 0 e 100)
e mais um:

```go
esperado := percentOf(metricas.Used, metricas.Total)
if metricas.UsagePercent != esperado { ... }
```

Esse é o teste que protege a decisão 2. Se alguém "simplificar" o collector
trocando o cálculo por `stat.UsedPercent`, o teste falha — e o comentário no
código explica por quê.

### Service e handler

Iguais em estrutura aos do Epic 04: dublê do collector, verificação de que a
falha vira `ErrUnavailable` com a causa preservada, e handler devolvendo 503.

Os helpers de percentual já haviam sido testados no Epic 04
(`TestPercentOf`, incluindo o caso `total = 0` que não pode gerar `NaN`), então
não foram retestados aqui.

---

## Validação de ponta a ponta

Resposta do endpoint:

```json
{"total":32449835008,"used":14910898176,"available":17538936832,"usagePercent":45.95061322291454}
```

Comparação com o sistema:

| Campo | Agente | Sistema | Diferença |
|---|---|---|---|
| `total` | 32449835008 | `MemTotal` = 32449835008 | exato |
| `available` | 17538936832 | `MemAvailable` = 17538428928 | ~0,5 MB |
| `used` | 14910898176 | `free -b` = 14912679936 | ~1,7 MB |

As diferenças em `available` e `used` são esperadas: as leituras aconteceram em
instantes diferentes e a memória muda continuamente. `total`, que é estático,
bate exatamente — o que confirma que não há erro de unidade ou conversão.

Conferência do percentual: 14910898176 / 32449835008 = 45,95%, igual ao
`usagePercent` publicado.
