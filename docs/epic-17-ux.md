# Epic 17 — UX e Formatação

**Status:** concluído
**Data:** 2026-08-30

Cobre as Tasks 17.1 (formatadores), 17.2 (estados visuais), 17.3 (filtro de
processos) e 17.4 (ordenação na UI).

---

## 17.1 — Formatadores

Já existiam desde a Epic 13, quando os cartões precisaram deles para exibir
qualquer coisa: `formatBytes`, `formatBytesPerSecond`, `formatPercent` e
`formatUptime`. Escritos sem `String.format`, que não existe no source set
comum. Os testes entram na Epic 18.

---

## 17.2 — Estados visuais

Duas peças, cada uma respondendo a uma pergunta diferente:

| Peça | Pergunta que responde |
|---|---|
| Etiqueta no cabeçalho | "está ligado agora?" |
| Faixa de estado | "o que houve e o que está sendo feito?" |

A faixa **só aparece quando há algo a dizer**. Uma faixa permanente roubaria
altura da tabela para repetir "está tudo bem".

```
Conectando ao agente…                    127.0.0.1:8080
Não foi possível falar com o agente. Tentando novamente…
O agente respondeu num formato inesperado. Tentando novamente…
```

Durante a conexão inicial a faixa mostra **onde** o agente é esperado — poupa a
pergunta seguinte de quem abriu o desktop sem o agente no ar.

O "Tentando novamente…" vem do ViewModel, não do repository: quem sabe que a
repetição é automática é a camada que a agendou, e dizer isso evita que o
usuário procure um botão de reconectar que não existe.

**Os dados continuam na tela** durante a queda — decisão tomada lá na Epic 14,
quando o estado foi desenhado para preservar a última leitura boa. A Task 17.2
pede exatamente isso ("Dashboard não desaparece durante reconexão"), e nada
precisou mudar.

---

## 17.3 — Filtro por nome

Case-insensitive: quem procura `chrome` não deve precisar saber se o processo
se chama `Chromium` ou `chromium`. Campo vazio devolve a lista inteira, e a
tabela troca a mensagem de vazio para dizer que o filtro é que não achou nada:

```
Nenhum processo corresponde a "postgres"
```

---

## 17.4 — Ordenação: critério no ViewModel, sentido na tela

Essa divisão não é arbitrária.

A lista chega **cortada nos 50 primeiros**. Reordenar localmente os 50 mais
pesados de CPU por memória mostraria os processos errados — o verdadeiro
top 50 de memória pode não ter nenhum deles. Por isso o **critério** sobe até o
ViewModel, que refaz o pedido ao agente com `sort=memory`.

O **sentido** (maior ou menor primeiro) é estado local da tela: inverte os
mesmos 50 registros, sem pedir nada a ninguém.

Trocar o critério não espera o intervalo de dois segundos:

```kotlin
withTimeoutOrNull(processesInterval) { processSort.first { it != sort } }
```

O laço acorda na hora se o critério mudar. Sem isso, um clique demoraria até
dois segundos para virar resultado e a interface pareceria travada.

O PID desempata a ordenação, como no agente: sem isso os processos parados em
0 % trocariam de lugar a cada atualização.

---

## Dois defeitos de layout que só a tela cheia revelou

A validação anterior tinha sido feita numa janela de ~620 px. Em 2560 px:

### Cartões espremidos com metade da tela vazia

A grade calculava `colunas = largura / largura_mínima` — em 2560 px isso dá
**nove** colunas para quatro cartões, e cada um ficava com um nono da largura.
Passou a ser `min(cartões, colunas)`.

### PID a 300 px do nome do processo

As colunas da tabela usavam pesos proporcionais. Numa janela larga, a coluna de
PID sozinha ficava com 330 px, quase todos vazios. As colunas numéricas
passaram a ter largura fixa (80 dp para PID, 110 dp para os números) e o nome
ficou com o resto.

---

## Validação

Aplicação em 2560×1440 com o agente publicando:

- quatro cartões ocupando a largura inteira, três gráficos abaixo;
- barra de processos com campo de filtro, chips `CPU` / `Memória` e
  `↓ Maior primeiro`;
- tabela com PID, nome, CPU e memória alinhados, `java 62 %` no topo;
- gráfico de rede com pico de download e teto automático em `545,4 KB/s`.

O filtro e a ordenação são funções puras (`filterByName`, `sortedBy`) e estão
cobertos por teste na Epic 18.
