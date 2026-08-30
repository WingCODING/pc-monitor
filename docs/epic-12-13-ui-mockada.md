# Epics 12 e 13 — Bootstrap do KMP e UI mockada

**Status:** concluído
**Data:** 2026-08-30

Cobre a Task 12.3 (estrutura em camadas do frontend) e as Tasks 13.1 a 13.4
(shell do dashboard, `MetricCard`, `NetworkCard` e tabela de processos).

As Tasks 12.1 e 12.2 — projeto Compose Multiplatform e dependências — já
tinham sido feitas no commit de bootstrap.

---

## Estrutura (Task 12.3)

```
commonMain/kotlin/pcmonitor/
├── model/        modelos de domínio, sem HTTP
├── network/      cliente Ktor            (Epic 14)
├── repository/   fronteira de dados      (Epic 14)
├── viewmodel/    estado da tela          (Epic 14)
└── ui/
    ├── theme/       paleta e cores de estado
    ├── format/      bytes, bytes/s, percentuais, uptime
    ├── components/  MetricCard, NetworkCard, ProcessTable, ConnectionStatus
    ├── preview/     dados de exemplo (removidos na Epic 14)
    ├── DashboardScreen.kt
    └── App.kt
```

`network/`, `repository/` e `viewmodel/` nascem vazios de propósito: a regra
"Composables não conhecem Ktor Client" é fácil de respeitar enquanto o Ktor
ainda não existe, e difícil de recuperar depois que a primeira chamada HTTP
entrar dentro de um `@Composable`.

---

## A tela foi montada antes de existir integração

`DashboardScreen` recebe **tudo por parâmetro** e não guarda estado:

```kotlin
fun DashboardScreen(
    metrics: DashboardMetrics?,
    processes: List<ProcessMetrics>,
    system: SystemMetrics?,
    connection: ConnectionState,
)
```

Foi o que permitiu resolver layout, formatação e estados vazios sem depender do
backend — e é o que vai permitir, na Epic 14, trocar a origem dos dados sem
tocar em nenhum componente.

### Os modelos já existem, sem serialização

`model/Metrics.kt` espelha os contratos do agente com **campos anuláveis** onde
o Go usa `omitempty`. A UI precisa distinguir "0 %" de "indisponível", e essa
distinção some se o modelo usar valores padrão. A anotação `@Serializable` só
entra na Epic 14: enquanto não há transporte, o modelo não precisa saber que um
dia será JSON.

---

## Decisões de componente

### `MetricCard` serve CPU, RAM e disco; a rede tem cartão próprio

Não existe "100 % da rede". Forçar a rede no cartão de medidor exigiria
inventar um teto, então `NetworkCard` mostra duas velocidades independentes com
setas ↓ e ↑.

### O valor ausente é tratado dentro do cartão

Quando um collector falha, o cartão continua na tela mostrando
"indisponível" — se sumisse, o layout inteiro pularia justo no momento em que o
usuário está tentando entender o que houve.

### O medidor é limitado a 0–100 na UI, não no modelo

O backend pode publicar 100,4 % por arredondamento entre núcleos. Isso é um
dado legítimo; uma barra passando do fim é um artefato visual.

### A cor acompanha o valor

Verde até 75 %, âmbar até 90 %, vermelho acima. Vale para os medidores e para a
coluna de CPU da tabela: numa lista de dezenas de linhas, quem está pesando
aparece sem que ninguém precise ler os números.

### `LazyColumn` com `key` por PID

A lista de processos é substituída inteira a cada atualização. Sem `key`, a
rolagem volta ao topo a cada segundo; com ela, o Compose reconhece as linhas e
mantém a posição mesmo quando a ordem muda.

---

## Grade fluida em vez de colunas fixas

O número de colunas sai da largura disponível:

```kotlin
val columns = max(1, (maxWidth / MIN_CARD_WIDTH).toInt())
```

A janela é redimensionável e, numa faixa estreita, quatro colunas fixas
espremeriam os cartões até os números não caberem. Os espaçadores da última
linha incompleta mantêm os cartões com a mesma largura das linhas cheias, em
vez de esticá-los para preencher o vão.

---

## Formatação sem `String.format`

`String.format` não existe no source set comum do Kotlin Multiplatform. Os
formatadores fazem o arredondamento na mão (`roundToLong`) e montam a string —
o que também deixou a vírgula decimal fixa em vez de vir do `Locale` da
máquina, já que o resto da interface está em português.

A casa decimal some quando é zero: `1 KB` lê melhor que `1,0 KB` sem perder
informação.

> Os formatadores são a Task 17.1, adiantada aqui porque os cartões precisavam
> deles para exibir qualquer coisa. A Epic 17 fecha os estados visuais e a
> Epic 18 traz os testes.

---

## Validação

A aplicação foi executada com os dados de exemplo e capturada numa janela de
626×334 — bem abaixo do tamanho inicial de 1180×800:

- a grade se reorganizou em 2 colunas;
- `12,2 GB`, `40,3 %`, `de 30,2 GB · 18 GB livres` e `ligado há 6 h 26 min`
  saíram formatados como esperado;
- o medidor de memória parou em 40 % da largura, em verde.

A captura também expôs um erro: o cartão de CPU mostrava `5%` como valor
principal (truncado por `toInt()`) e `6 %` na etiqueta ao lado (arredondado
pelo formatador) — o mesmo número em duas versões diferentes. O valor principal
passou a usar o formatador, e a etiqueta some quando o valor já é um
percentual.

Os dados de exemplo usam os números reais observados na validação do agente. Um
mock arredondado esconderia justamente esse tipo de problema de formatação.
