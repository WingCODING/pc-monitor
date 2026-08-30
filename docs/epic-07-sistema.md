# Epic 07 — Informações do Sistema

**Status:** concluído
**Data:** 2026-08-30

Cobre as Tasks 7.1 (collector), 7.2 (`GET /api/v1/system`) e 7.3 (integração do
uptime ao snapshot agregado).

---

## O que foi feito

| Arquivo | Papel |
|---|---|
| `internal/collector/system.go` | Collector gopsutil `host`, sem estado |
| `internal/service/system.go` | `SystemService` |
| `internal/api/system.go` | Handler de `GET /api/v1/system` |
| `internal/service/metrics.go` | Passou a receber `*SystemService` e preencher o uptime |

---

## Decisões e por quê

### 1. `architecture` usa `runtime.GOARCH`, não `KernelArch`

O gopsutil oferece `info.KernelArch`, que nesta máquina devolve `"x86_64"` — o
mesmo que `uname -m`. O `plan.md` §14, porém, exemplifica o contrato com
`"architecture": "amd64"`.

Optei por respeitar o contrato: `runtime.GOARCH` devolve `"amd64"`.

A ressalva registrada: `runtime.GOARCH` é a arquitetura para a qual o **binário
do agente** foi compilado, não a da máquina. Um build 386 rodando num kernel
x86_64 reportaria `"386"`, enquanto `KernelArch` acertaria. Para o MVP os dois
coincidem, e a aderência ao contrato vale mais — trocar depois exige mexer no
model KMP junto.

### 2. `osVersion` tem fonte primária e fallback

`PlatformVersion` descreve a distribuição e é o mais informativo, mas várias
distros não o preenchem. Quando vazio, o collector cai para `KernelVersion`,
que ainda diz algo útil.

Nesta máquina: `PlatformVersion = "4.0.1"` (Omarchy), com
`KernelVersion = "7.1.9-arch1-2"` disponível como reserva.

A escolha ficou isolada na função `osVersion`, que é pura e testável sem
depender do que a máquina de teste expõe.

### 3. O uptime é copiado para variável local antes de virar ponteiro

```go
uptime := metrics.UptimeSeconds
snapshot.UptimeSeconds = &uptime
```

Parece redundante, mas não é: `metrics` é a variável do `if`, e apontar
diretamente para o campo de uma variável de escopo curto é o tipo de coisa que
funciona hoje e quebra numa refatoração futura. A cópia local torna a posse do
valor explícita.

### 4. O uptime entra no snapshot, o resto das informações não

`DashboardMetrics` recebe apenas `uptimeSeconds` do System Collector — não
hostname, OS nem arquitetura.

Motivo: o snapshot é publicado **a cada segundo** pelo WebSocket a partir do
Epic 15. Hostname e arquitetura são estáticos; retransmiti-los 86.400 vezes por
dia seria desperdício. A UI busca esses dados uma vez em `/api/v1/system` e o
uptime, que muda, vem no fluxo contínuo.

---

## Erros encontrados durante a implementação

### Quebra proposital de assinatura

Adicionar o `SystemService` ao `MetricsService` mudou
`NewMetricsService(cpu, memory)` para `NewMetricsService(cpu, memory, system)`,
quebrando a compilação de dois arquivos de teste:

```
vet: internal/service/metrics_test.go:13:72: not enough arguments in call to NewMetricsService
vet: internal/api/metrics_test.go:18:2: not enough arguments in call to service.NewMetricsService
```

Falha esperada e desejável: o compilador apontou exatamente os pontos que
precisavam de atualização. É o mesmo raciocínio que motivou incluir
`context.Context` nas interfaces já no Epic 03 — mudanças de assinatura são
baratas quando o compilador as encontra, e caras quando passam despercebidas.

Aproveitei para reforçar `TestMetricsServiceTodasAsMetricasFalham`, que agora
verifica também que `UptimeSeconds` é omitido quando o System Collector falha.

---

## Como os testes foram escritos

### Comparação com fontes independentes do gopsutil

`TestSystemCollectorMaquinaReal` não confere o resultado contra o próprio
gopsutil — isso seria circular. Compara com fontes independentes:

| Campo | Verificado contra |
|---|---|
| `Hostname` | `os.Hostname()` da stdlib |
| `OS` | `runtime.GOOS` |
| `Architecture` | `runtime.GOARCH` |
| `UptimeSeconds` | apenas `> 0` |
| `OSVersion` | apenas não vazia |

É assim que se detecta um mapeamento de campo trocado — por exemplo, atribuir
`Platform` onde deveria ir `Hostname`. Um teste que só verificasse "campo não
vazio" passaria com os dois campos invertidos.

`UptimeSeconds` e `OSVersion` não têm valor esperado fixo, porque variam por
máquina e por execução.

### Escolha de versão testada como função pura

`TestOSVersionPreferePlatformVersion` monta `host.InfoStat` à mão e cobre os
três casos: só `PlatformVersion`, só `KernelVersion`, e nenhuma das duas. Isso
testa a regra de fallback numa máquina que, por acaso, preenche ambas — o caso
do fallback nunca seria exercitado aqui de outra forma.

---

## Validação de ponta a ponta

`GET /api/v1/system`:

```json
{
    "hostname": "omarchy",
    "os": "linux",
    "osVersion": "4.0.1",
    "architecture": "amd64",
    "uptimeSeconds": 13295
}
```

| Campo | Agente | Sistema |
|---|---|---|
| `hostname` | omarchy | `hostname` → omarchy |
| `architecture` | amd64 | contrato do `plan.md` §14 |
| `uptimeSeconds` | 13295 | `/proc/uptime` → 13294 |

O uptime de 13295 s equivale a 3 h 41 min, batendo com
`uptime -p` → "up 3 hours, 41 minutes". A diferença de 1 s vem do intervalo
entre as duas leituras.

**Task 7.3 confirmada:** `GET /api/v1/metrics` passou a incluir
`"uptimeSeconds": 13295`, o campo que o Epic 06 deixara deliberadamente
ausente.
