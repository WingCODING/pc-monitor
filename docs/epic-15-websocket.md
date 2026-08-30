# Epic 15 — WebSocket e Realtime

**Status:** concluído
**Data:** 2026-08-30

Cobre as Tasks 15.1 (endpoint Go), 15.2 (loop de publicação), 15.3 (cliente
KMP), 15.4 (integração no repository) e 15.5 (reconexão).

O epic que transforma dois programas que se falam sob demanda em um dashboard
que acompanha a máquina.

---

## O que foi feito

| Arquivo | Papel |
|---|---|
| `agent/internal/service/hub.go` | `MetricsHub`: uma coleta por intervalo, muitos clientes |
| `agent/internal/api/websocket.go` | Handler de `GET /api/v1/ws/metrics` |
| `desktop/.../network/AgentClient.kt` | `snapshots(): Flow<DashboardMetrics>` |
| `desktop/.../repository/MetricsRepository.kt` | `observeMetrics()` com reconexão |
| `desktop/.../viewmodel/MetricsViewModel.kt` | Passou a coletar o fluxo em vez de fazer polling |

Dependência nova no agente: `github.com/coder/websocket`.

---

## 15.1 / 15.2 — O hub existe para não multiplicar a coleta

A alternativa óbvia — cada conexão com seu próprio loop — quebra de duas
formas:

1. **Custo:** dez janelas abertas viram dez varreduras de `/proc` por segundo.
2. **Correção:** CPU, rede e processos calculam **deltas**. Dois loops
   coletando em paralelo consomem o baseline um do outro, e as duas leituras
   saem erradas.

O hub coleta uma vez por tique e distribui o **mesmo** snapshot.
`TestMetricsHubUmaColetaAlimentaTodos` fixa isso: 3 clientes, 1 coleta.

### Fila de um lugar por cliente, com descarte

```go
select {
case subscriber <- snapshot:
default:
    // cliente lento: descarta
}
```

Se o cliente ainda não consumiu o snapshot anterior, mandar um mais novo por
cima não ajuda — ele está atrasado, e o que interessa é sempre a leitura mais
recente. Bloquear seria pior: um cliente travado atrasaria todos os outros.

### Sem clientes, sem coleta

O agente pode ficar aberto o dia inteiro sem nenhuma janela. Ler `/proc` nesse
tempo é gasto puro.

### Conexão nova acorda o loop

Sem isso, uma janela recém-aberta esperaria um intervalo inteiro pelo primeiro
snapshot. `Subscribe` enfileira um pedido de publicação imediata (sem bloquear,
porque um pedido pendente já serve para todos).

### O desligamento fecha os canais

É assim que os handlers sabem que devem encerrar — o hub nunca precisa
conhecê-los. Uma conexão que chega no meio do desligamento recebe um canal já
fechado em vez de ficar pendurada.

---

## O handler não coleta nada

Ele se inscreve, repassa e sai. Dois detalhes que separam uma conexão bem
encerrada de um vazamento:

- **`conn.CloseRead(ctx)`** descarta o que o cliente mandar e cancela o
  contexto assim que a conexão cai. Sem ele, a goroutine publicaria para um
  socket morto até a primeira escrita falhar — e o protocolo exige alguém
  lendo para responder aos pings.
- **Prazo de 5 s por escrita.** Um cliente que parou de ler manteria a
  goroutine presa até o TCP desistir, o que leva minutos.

`TestWebSocketRemoveClienteQueSomeSemFechar` cobre o caminho feio — processo
morto, sem handshake de fechamento — e verifica que o hub volta a zero
clientes.

---

## 15.3 — `channelFlow`, não `flow`

As mensagens chegam na corrotina da sessão WebSocket. Emitir de outra corrotina
quebra a regra de contexto de um `flow` comum; `channelFlow` existe exatamente
para esse caso.

O cliente **não** reconecta: o fluxo termina quando o agente fecha e lança
quando a conexão cai. A política de repetição pertence ao repository, uma
camada acima.

---

## 15.4 / 15.5 — Reconexão com espera crescente

```
1 s → 2 s → 4 s → 8 s → 10 s → 10 s …
```

Volta a 1 s a cada conexão bem-sucedida. Um laço apertado gastaria CPU dos dois
lados enquanto o agente está fora; esperar sempre 10 s faria o dashboard
demorar a voltar depois de um reinício rápido.

**Nada de conexões duplicadas:** o laço só recomeça depois que a sessão
anterior terminou, e o fluxo é coletado num único ponto do ViewModel. Não há
`launch` por tentativa.

O snapshot deixou de ser buscado por polling REST. A lista de processos
continua em REST a cada 2 s — ela não faz parte do contrato
`DashboardMetrics`, e a varredura de `/proc` é cara demais para ir junto no
fluxo de 1 s.

---

## Validação de ponta a ponta

Agente e desktop rodando juntos, com o log do agente registrando cada conexão:

```
INFO msg="servidor iniciado" addr=127.0.0.1:8080 collectionInterval=1s
INFO msg="cliente conectado ao WebSocket" remote=127.0.0.1:58020
```

| Momento | Etiqueta | Tela |
|---|---|---|
| Conectado | `Conectado` (verde) | valores atualizando a cada segundo |
| Agente derrubado | `Desconectado` (vermelho) em ~2 s | **últimos valores permanecem** |
| Agente de volta | `Conectado` em ~2 s | valores voltam a atualizar |

```
INFO msg="cliente conectado ao WebSocket" remote=127.0.0.1:55812
```

O dashboard não piscou nem esvaziou durante a queda — o que a Task 17.2 vai
exigir e que já estava resolvido no desenho do estado.
