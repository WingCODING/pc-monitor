# Epic 08 — Discos

**Status:** concluído
**Data:** 2026-08-30

Cobre as Tasks 8.1 (collector), 8.2 (`GET /api/v1/disks`) e 8.3 (resumo de
disco no snapshot agregado).

O epic mais delicado do backend até aqui: uma armadilha de contagem dupla que
teria produzido números completamente errados na UI.

---

## O que foi feito

| Arquivo | Papel |
|---|---|
| `internal/collector/disk.go` | Collector gopsutil com deduplicação por dispositivo |
| `internal/service/disk.go` | `DiskService` com `List` e `Summary` |
| `internal/api/disk.go` | Handler de `GET /api/v1/disks` |
| `internal/service/metrics.go` | Passou a preencher `disk` no snapshot |

---

## O problema central: subvolumes btrfs inflam a capacidade

Antes de escrever o collector, sondei o que o gopsutil devolve nesta máquina.
O resultado justificou a sondagem:

```
/dev/dm-0        /                        btrfs  total=998037782528  38.3%
/dev/dm-0        /var/cache/pacman/pkg    btrfs  total=998037782528  38.3%
/dev/dm-0        /var/log                 btrfs  total=998037782528  38.3%
/dev/dm-0        /home                    btrfs  total=998037782528  38.3%
/dev/nvme0n1p1   /boot                    vfat   total=2143281152    27.1%
/dev/sdb1        /run/media/will/Ventoy   exfat  total=15559819264   40.0%
```

**`/dev/dm-0` aparece quatro vezes.** São subvolumes btrfs do mesmo
filesystem, e cada um reporta a capacidade **inteira** do dispositivo — não uma
fatia dele. O mesmo aconteceria com bind mounts.

Somar essas entradas para o resumo do dashboard daria:

```
998 GB × 4 + 2,1 GB + 15,5 GB ≈ 4 TB
```

num disco de 1 TB. E o "usado" seria inflado na mesma proporção. A UI mostraria
um número sem relação com a realidade, e nada no código acusaria erro.

### A solução

`deduplicateByDevice` mantém **uma entrada por dispositivo**. Entre montagens
concorrentes vence a de caminho mais curto, que é a raiz do filesystem — `/`
em vez de `/home`.

A ordem original é preservada em vez de usar um mapa direto, porque iteração de
mapa em Go é aleatória e a saída da API precisa ser determinística.

A deduplicação acontece no **collector**, não no resumo. Assim tanto
`GET /api/v1/disks` quanto o snapshot enxergam a mesma lista, e a soma em
`summarizeDisks` pode assumir entradas únicas. A alternativa — deduplicar só na
soma — deixaria a listagem com quatro linhas idênticas de 998 GB, o que é
confuso e sugere quatro discos.

---

## Outras decisões

### `Partitions(all=false)`

Descarta pseudo-filesystems (`tmpfs`, `devtmpfs`, `efivarfs`, `overlay`). Com
`all=true` a máquina devolve 31 entradas, quase todas sem armazenamento real.

### Falha de uma partição não derruba a coleta

`disk.Usage()` pode falhar por permissão negada, mídia removida ou montagem de
rede fora do ar. Nesses casos o collector registra `WARN` e **pula** a
partição, atendendo "filesystems inválidos são ignorados ou tratados".

### `Total == 0` é descartado

Filesystem sem capacidade própria só adicionaria uma linha "0 B" à lista.

### Lista vazia serializa como `[]`, nunca `null`

Uma slice `nil` em Go vira `null` no JSON. O contrato promete um array, e um
cliente que itere sobre `null` quebra. O handler converte explicitamente.

---

## Erros encontrados durante a implementação

### Teste falhando por 1 ulp de ponto flutuante

`TestSummarizeDisks` falhou:

```
obtido:   33.33333333333333
esperado: 33.333333333333336
```

O valor esperado foi escrito como `500.0 / 1500.0 * 100`. Go avalia
**constantes sem tipo em precisão arbitrária** durante a compilação e arredonda
uma única vez no fim; a função faz duas operações em `float64`, arredondando
duas vezes. Os resultados diferem no último bit.

O erro estava no teste, não no código. Corrigido com comparação por tolerância
(`1e-9`) no campo de percentual, mantendo comparação exata em `Total` e `Used`,
que são inteiros.

Lição registrada: comparar `float64` por igualdade exata em teste é frágil,
mesmo quando a matemática "obviamente" bate. Os outros dois subtestes passaram
apenas porque 25% e 0% são exatamente representáveis em binário.

### `diskCollectorFalso` referenciado antes de existir

Ao atualizar os helpers de `metrics_test.go` referenciei `diskCollectorFalso`
no pacote `api` antes de criá-lo. Compilador acusou; o dublê foi criado junto
com os testes do handler.

---

## Como os testes foram escritos

### Deduplicação como função pura, com dados reais desta máquina

`TestDeduplicateByDevice` é o teste central. Os casos usam os valores
**observados na sondagem**, não números inventados:

| Caso | O que prova |
|---|---|
| subvolumes btrfs colapsam na raiz | 4 entradas de `/dev/dm-0` → 1, em `/` |
| dispositivos distintos preservados | dedupe não é agressivo demais |
| vence o caminho mais curto fora de ordem | `/home/will/dados` antes de `/` ainda escolhe `/` |
| lista vazia | sem panic |

O terceiro caso é o que remove a dependência de ordem: se a implementação
apenas mantivesse a primeira ocorrência, esse teste falharia.

### Invariante de unicidade na máquina real

`TestDiskCollectorMaquinaReal` acumula os dispositivos já vistos e falha se
algum repetir:

```go
if anterior, repetido := dispositivos[atual.Name]; repetido {
    t.Errorf("dispositivo %s repetido em %q e %q; a deduplicação falhou", ...)
}
```

Isso é o que pega uma regressão real nesta máquina — que tem exatamente a
topologia problemática.

### Resumo protegido contra inflação

`TestSummarizeDisksNaoInflaComSubvolumes` documenta, em forma de teste, por que
a deduplicação vive no collector: a soma **assume** entradas únicas. Se alguém
mover a dedupe para outro lugar, o comentário e o teste explicam a consequência.

### Handler

Três casos: sucesso com array, lista vazia serializando como `[]` (não `null`),
e 503 na falha.

---

## Validação de ponta a ponta

`GET /api/v1/disks` — 6 montagens reduzidas a 3 dispositivos:

```json
[
  {"name":"/dev/dm-0","mountPoint":"/","total":998037782528,"usagePercent":38.31},
  {"name":"/dev/nvme0n1p1","mountPoint":"/boot","total":2143281152,"usagePercent":27.13},
  {"name":"/dev/sdb1","mountPoint":"/run/media/will/Ventoy","total":15559819264,"usagePercent":40.03}
]
```

Resumo em `/api/v1/metrics` comparado com a soma dos filesystems distintos
segundo o `df`:

| | Agente | `df` (deduplicado) |
|---|---|---|
| total | 1 015 740 882 944 | 1 015 740 882 944 |
| used | 389 169 307 648 | 389 169 307 648 |
| percentual | 38,31 % | 38,31 % |

Coincidência **byte a byte**. Sem a deduplicação, o total seria de
aproximadamente 4 TB.
