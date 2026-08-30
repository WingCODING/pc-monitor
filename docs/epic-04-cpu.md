# Epic 04 — CPU

**Status:** concluído
**Data:** 2026-08-30
**Commit:** ver `git log --grep "Epic 04"`

Primeiro collector real do projeto. Cobre as Tasks 4.1 (collector), 4.2 (service)
e 4.3 (endpoint `GET /api/v1/cpu`).

---

## O que foi feito

| Arquivo | Papel |
|---|---|
| `internal/collector/percent.go` | `clampPercent` e `percentOf`, compartilhados com RAM e disco |
| `internal/collector/cpu.go` | Collector gopsutil com baseline próprio |
| `internal/service/errors.go` | `ErrUnavailable`, erro de domínio traduzido em 503 |
| `internal/service/cpu.go` | `CPUService`, recebe o collector por dependência |
| `internal/api/errors.go` | `writeError`, resposta de erro padronizada |
| `internal/api/cpu.go` | Handler de `GET /api/v1/cpu` |
| `internal/api/router.go` | Passou a receber `Deps` com os services |

Dependência adicionada: `github.com/shirou/gopsutil/v4 v4.26.7`.

---

## Decisões e por quê

### 1. O collector guarda o próprio baseline em vez de usar `cpu.Percent`

Esta foi a decisão mais importante do epic.

O caminho óbvio seria `cpu.Percent(0, false)` do gopsutil, que devolve o uso
desde a chamada anterior. Ao ler a implementação (`cpu/cpu.go:186`), porém, o
baseline fica em **estado global do pacote**:

```go
lastCPUPercent.Lock()
lastTimes = lastCPUPercent.lastCPUTimes
lastCPUPercent.lastCPUTimes = cpuTimes
```

Cada chamada **consome e substitui** o baseline de todo o processo. A partir do
Epic 15 o loop do WebSocket coletará a cada 1s enquanto requisições REST
chegam em paralelo — as duas chamadas roubariam o baseline uma da outra, e a
segunda calcularia o uso sobre um intervalo de poucos milissegundos,
devolvendo valores erráticos.

A solução foi ler `cpu.TimesWithContext` e calcular o delta na própria
instância do collector, protegida por mutex. O cálculo (`busyTimes`) replica a
lógica do gopsutil, incluindo o detalhe de que **no Linux `Guest` e `GuestNice`
já estão contabilizados dentro de `User`** — somá-los de novo inflaria o total
e subestimaria o uso.

### 2. A primeira leitura usa o zero absoluto como baseline

Sem leitura anterior, o baseline é `TimesStat{}`. Como os contadores do kernel
começam em zero no boot, o delta contra zero é exatamente a média desde o boot
— um número real, em vez de um `0%` enganoso na primeira requisição.

### 3. `ErrUnavailable` separa "tente de novo" de "erro definitivo"

O service embrulha falhas do collector em `ErrUnavailable`, e o handler
traduz em **503**, não 500. A distinção importa para o frontend: 503 significa
que a métrica pode voltar, o que alimenta o estado de reconexão do Epic 15.

A causa original é preservada na cadeia (`%w`) para aparecer no log, mas
**nunca** é enviada ao cliente — há teste garantindo isso.

### 4. `cores` cai para o total lógico quando não há núcleos físicos

Contêineres e algumas VMs não expõem núcleos físicos. Em vez de falhar, o
collector usa o total lógico, atendendo o critério "Cores > 0".

---

## Erros encontrados durante a implementação

### Porta 8080 ocupada mascarada como 404

Na primeira validação ponta a ponta, `GET /api/v1/cpu` devolveu **404**. A
rota parecia errada, mas a causa era outra: um `go run ./cmd/server` **do
próprio usuário**, iniciado às 14:02, ainda ocupava a 8080. Meu binário nem
subiu (`bind: address already in use`), e o curl acabou batendo na build
antiga — que não tinha a rota de CPU.

Duas lições:

- O loop de espera do healthcheck confirmava que *algo* respondia na porta,
  não que *o meu processo* respondia. Um healthcheck que não identifica a
  versão do binário dá falso positivo.
- A validação passou a rodar em **porta efêmera** (`127.0.0.1:0`) através de um
  programa descartável, sem tocar no processo do usuário.

Isso reforça a necessidade da **Task 11.3** (host e porta configuráveis): hoje
`defaultAddr` é constante, o que impede rodar duas instâncias em paralelo.

### Import faltando no teste

`cpu_test.go` usou `strings.Contains` sem importar `strings`. Pego pelo
compilador na primeira execução dos testes, corrigido antes do commit.

---

## Como os testes foram escritos

Três níveis, do mais isolado ao mais real:

### Funções puras — sem tocar no sistema operacional

`TestBusyPercent` e `TestClampPercent`/`TestPercentOf` recebem structs
montados à mão. Isso permite cobrir cenários **impossíveis de provocar numa
máquina real**:

- contador reiniciado (valor atual menor que o anterior) → deve dar 0, nunca
  negativo;
- leitura repetida no mesmo tick → delta zero, sem divisão por zero;
- `total = 0` → devolve 0 em vez de `NaN`, que quebraria a desserialização no
  cliente (há asserção explícita de `IsNaN`/`IsInf`);
- `iowait` contando como ocioso.

### Service — com dublê do collector

`cpuCollectorFalso` implementa a interface do Epic 03. É o que prova, na
prática, o critério "implementações podem ser substituídas em testes": dá para
simular falha de hardware sem ter hardware falhando.

O teste de erro verifica **duas** propriedades: que o erro é classificável como
`ErrUnavailable` (para o handler mapear o status) e que a causa original
sobrevive na cadeia (para o log).

### Handler — router completo, collector falso

Os testes de HTTP passam pelo `NewRouter` real, cobrindo roteamento, status,
`Content-Type` e corpo de uma vez. Três casos: sucesso, indisponível (503) e
não vazamento de detalhe interno.

### Máquina real

`TestCPUCollectorMaquinaReal` valida os invariantes contra o sistema em
execução: `Cores > 0`, `Threads > 0`, `Threads >= Cores`, uso entre 0 e 100.
Não fixa valores absolutos, que variariam por máquina.

---

## Validação de ponta a ponta

Comparação contra o sistema:

| Campo | Agente | Fonte do sistema |
|---|---|---|
| `model` | AMD Ryzen 9 9950X3D 16-Core Processor | `/proc/cpuinfo` |
| `cores` | 16 | `lscpu` |
| `threads` | 32 | `nproc --all` |

Resposta ao teste de carga — **8 workers ocupados numa máquina de 32 threads
deveriam dar 25%**:

| Cenário | Medido |
|---|---|
| 8 workers em loop | 25,8 % |
| Repouso | 0,8 % |

A proximidade entre 25% teórico e 25,8% medido confirma que o cálculo do delta
está correto.

---

## Pendências deixadas para outros epics

- **Task 11.3** — `defaultAddr` ainda é constante; a colisão de porta descrita
  acima mostra por que host e porta precisam ser configuráveis.
- **Epic 15** — o REST e o loop do WebSocket vão chamar o mesmo collector.
  Com o baseline por instância isso é seguro, mas o intervalo entre chamadas
  REST esparsas ainda representa "média desde a última leitura". Um amostrador
  único em background resolve, e é o que o Epic 15 pede.
