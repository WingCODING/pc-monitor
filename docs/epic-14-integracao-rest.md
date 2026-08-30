# Epic 14 — Integração REST KMP ↔ Go

**Status:** concluído
**Data:** 2026-08-30

Cobre as Tasks 14.1 a 14.6: models serializáveis, cliente Ktor, repositories,
ViewModel e substituição dos mocks por dados reais.

---

## O que foi feito

| Arquivo | Papel |
|---|---|
| `model/Metrics.kt` | `@Serializable` sobre os modelos que já existiam |
| `network/AgentClient.kt` | Único ponto do frontend que conhece Ktor |
| `repository/AgentResult.kt` | Resultado de domínio e classificação de falhas |
| `repository/MetricsRepository.kt` | `getMetrics()`, `getSystem()` |
| `repository/ProcessesRepository.kt` | `getProcesses()` |
| `viewmodel/MetricsViewModel.kt` | `StateFlow<DashboardUiState>` |
| `ui/App.kt` | Monta a cadeia e entrega o estado à tela |

`ui/preview/SampleData.kt` foi removido: com dados reais na tela, mock parado
no repositório só envelhece.

---

## 14.1 — Os nomes dos campos são o contrato

Nenhuma `@SerialName` foi necessária: as propriedades Kotlin já têm exatamente
os nomes que o Go publica. Isso é proposital — quem renomear de um lado quebra
o outro, e uma camada de tradução esconderia a quebra até virar dado errado na
tela.

`ignoreUnknownKeys = true` no `Json`: o agente pode ganhar campos novos sem
derrubar uma versão antiga do desktop, que é o contrário do padrão estrito do
kotlinx.serialization.

---

## 14.2 — Prazos curtos de propósito

```kotlin
connectTimeoutMillis = 1_500
requestTimeoutMillis = 4_000
```

O agente roda em `localhost`. Uma requisição que passe disso significa que ele
travou, e esperar mais só atrasaria a UI mostrar "desconectado".

O `AgentClient` também é quem monta a URL do WebSocket
(`http` → `ws`), para que exista **um** lugar no frontend onde um endereço de
agente é escrito.

### Status de erro vira exceção com o corpo já interpretado

Sem isso, um 503 chegaria ao repository como falha de desserialização — o
dashboard diria "resposta inválida" onde a resposta era um erro perfeitamente
bem formado dizendo que o collector está fora do ar.

---

## 14.3 / 14.4 — Resultado de domínio, não exceção

```kotlin
sealed interface AgentResult<out T> {
    data class Success<T>(val value: T)
    data class Failure(val reason: FailureReason, val message: String)
}
```

`FailureReason` separa **`Unreachable`** de todo o resto, porque só ela
significa "o agente não está lá". As outras significam que ele está e algo deu
errado — e a ação do usuário muda: subir o agente contra investigar o que ele
respondeu.

`CancellationException` é relançada antes de qualquer conversão. Engoli-la
transformaria o fechamento da janela num erro exibido na tela, e quebraria o
cancelamento estruturado das corrotinas.

Os dois repositories existem separados por causa da **cadência**: o snapshot é
barato e chega a cada segundo; a varredura de processos custa uma volta inteira
em `/proc` no agente.

---

## 14.5 — ViewModel sem framework

O desktop não tem ciclo de vida de Android, então o escopo de corrotinas vem de
fora: da janela em produção, de um `TestScope` nos testes. Nenhuma linha do
ViewModel conhece HTTP.

Três decisões que a tela depende:

### As métricas ficam no estado mesmo com a ligação caída

Apagar a tela a cada oscilação faria o dashboard piscar e esconderia a última
leitura boa — justamente o que o usuário quer ver enquanto o agente volta.

### Só o fluxo de métricas mexe no estado da ligação

Se métricas e processos escrevessem no mesmo campo em cadências diferentes
(1 s e 2 s), a etiqueta de conexão ficaria oscilando entre Conectado e
Desconectado. A falha em processos preserva a lista anterior e não toca na
ligação.

### Informações da máquina são lidas uma vez, com repetição enquanto falhar

Hostname, sistema e arquitetura não mudam enquanto o agente roda — não
precisam trafegar a cada segundo. Mas o desktop pode abrir **antes** do agente,
caso comum quando os dois sobem juntos, então a leitura se repete a cada 5 s até
passar.

---

## 14.6 — Dados reais na tela

O endereço do agente pode vir de `PCMON_AGENT_URL`. Sem essa saída, mudar
`PCMON_PORT` no agente deixaria o desktop preso no 8080, mostrando
"desconectado" para um agente de pé.

Validação com agente e desktop rodando juntos:

| | Na tela |
|---|---|
| Máquina | `omarchy · linux 4.0.1 · amd64 · ligado há 6 h 49 min` |
| CPU | `2,4 %` — AMD Ryzen 9 9950X3D, 16C / 32T |
| Memória | `16,5 GB` de 30,2 GB, 54,7 %, 13,7 GB livres |
| Disco | `362,8 GB` de 946 GB, 38,3 % |
| Rede | ↓ 1,7 KB/s ↑ 316 B/s |
| Processos | java 20,6 % · 431,6 MB, chromium 14,2 %, Hyprland, Discord… |

Os quatro cartões couberam numa linha na janela de 1261 px, e a tabela de
processos ocupou o resto da altura.
