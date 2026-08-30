# Epic 16 — Histórico e Gráficos

**Status:** concluído
**Data:** 2026-08-30

Cobre as Tasks 16.1 (buffer de 60 pontos), 16.2 (gráfico de CPU), 16.3
(gráfico de RAM) e 16.4 (gráfico de rede).

---

## O que foi feito

| Arquivo | Papel |
|---|---|
| `viewmodel/MetricsHistory.kt` | Últimos 60 pontos de cada série |
| `ui/components/MetricChart.kt` | Gráfico de linha desenhado em `Canvas` |
| `ui/DashboardScreen.kt` | Área de gráficos e layout compacto |

---

## 16.1 — Séries independentes, sem preencher buracos

Cada série tem sua própria lista. Um snapshot com a CPU indisponível não entra
na série de CPU, mas continua alimentando memória e rede — furar todas as
séries junto jogaria fora leituras boas por causa de um collector quebrado.

**Ausência não vira zero.** Um buraco desenhado como zero mentiria: diria "a
máquina estava parada" onde a verdade é "não sabemos o que houve aqui".

### Lista imutável, não buffer circular

Sessenta cópias de uma lista de sessenta `Double` por minuto é ruído perto de
qualquer outra coisa que a tela faz. Um buffer mutável compartilhado exigiria
sincronização com o Compose para não ganhar nada — e o estado do dashboard já é
um `data class` copiado a cada atualização.

### A atualização saiu da thread da interface

O escopo de corrotinas vem da composição, que roda na thread da UI. Enquanto
era só `getMetrics()`, passava; com um snapshot por segundo e uma lista de 50
processos a cada dois, desserializar ali disputaria o tempo do desenho. Os três
laços passaram a rodar em `Dispatchers.Default`.

---

## 16.2 / 16.3 — CPU e RAM com escala fixa 0–100

A percepção de "muito" ou "pouco" depende de a altura significar sempre a mesma
coisa. Com escala automática, uma máquina ociosa desenharia o mesmo gráfico
imponente de uma máquina saturada, só que com outro rótulo.

## 16.4 — Rede com escala automática

Não existe teto conhecido de banda. Escala fixa deixaria a linha colada no chão
numa conexão rápida, ou estourada numa lenta. O teto é o pico da janela mais
15 % de folga — sem a folga, o pico encosta na borda de cima e some.

**Piso de 1 KB/s.** Sem ele, uma rede parada faria o gráfico ampliar ruído de
alguns bytes até parecer tráfego real.

Download e upload dividem o mesmo gráfico e a mesma escala: é a comparação
entre os dois que interessa.

---

## Os pontos entram pela direita

```kotlin
val stepX = size.width / (CAPACITY - 1)
val offset = CAPACITY - values.size
```

Com poucos pontos, distribuí-los por toda a largura faria o gráfico "esticar"
nos primeiros segundos, como se o intervalo de tempo estivesse mudando. Assim,
a linha entra pela direita e caminha para a esquerda conforme envelhece,
preenchendo o minuto.

Cada série desenha três coisas: a área com gradiente (que dá volume sem
competir com a linha), a linha, e um ponto no valor mais recente.

---

## O layout compacto

Com os gráficos, a tela passou a ter conteúdo demais para uma janela baixa:
cartões e gráficos sozinhos ocupavam a altura toda e a tabela de processos
ficava com zero pixels — sumia sem aviso.

Abaixo de 720 dp de altura a página inteira passa a rolar e a tabela ganha
altura fixa. Acima, o layout continua fixo com a tabela ocupando o que sobra.

---

## Validação

Aplicação aberta por ~1 minuto com o agente publicando pelo WebSocket, com
downloads disparados no meio para mexer a série de rede:

- gráfico de CPU com a linha entrando pela direita e o rótulo `100 % · 60 s`;
- valor corrente na legenda batendo com o cartão (`8,8 %` nos dois);
- gráfico de memória estável em torno de 54,9 %;
- janela de 621×688 rolando por inteiro, com a tabela abaixo dos gráficos.
