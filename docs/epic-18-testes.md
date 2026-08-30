# Epic 18 — Testes

**Status:** concluído
**Data:** 2026-08-30

Cobre as Tasks 18.1 (cálculos Go), 18.2 (handlers Go), 18.3 (repositories KMP)
e 18.4 (ViewModel).

Metade deste epic já estava feita: cada epic anterior entregou seus testes
junto com o código. O que faltava era o **lado KMP**, que até aqui não tinha
sequer um source set de teste.

---

## 18.1 / 18.2 — O que o agente já tinha

| Requisito da spec | Onde está |
|---|---|
| Percentuais | `TestPercentOf`, `TestClampPercent`, `TestSummarizeDisks` (lista vazia não gera NaN) |
| Velocidade de rede | `TestByteRate`, `TestSpeeds` |
| Reset de contador | `TestByteRateNuncaEstouraComWraparound` (uint64 quase no máximo) |
| Tratamento de deltas | `TestSampleCPU`, `TestCPUCollectorLeiturasConsecutivas` |
| Agregações | `TestSummarizeDisks*`, `TestSummarizeNetwork`, `TestMetricsService*` |
| Valores inválidos | tempo de CPU regredido, intervalo zero, total zero, PID reaproveitado |
| `/health`, `/cpu`, `/memory`, `/metrics` | `TestHandleHealth`, `TestHandleCPU*`, `TestHandleMemory*`, `TestHandleMetrics*` |
| Respostas de erro | 503, 400, 404, 405, 500 por panic |
| Contrato JSON | `TestContratoJSON`, `TestDashboardMetricsCompleto`, `TestDashboardMetricsParcial` |

Cobertura do agente:

```
internal/api         88,3 %
internal/collector   90,7 %
internal/config     100,0 %
internal/service     96,9 %
```

---

## 18.3 — Repositories com `MockEngine`

O `MockEngine` do Ktor responde no lugar da rede. Os testes exercitam o
**mesmo** caminho de desserialização e de classificação de erro que a aplicação
usa — não um caminho paralelo montado para o teste.

| Teste | O que prova |
|---|---|
| `leSnapshotCompleto` | JSON real do agente vira `DashboardMetrics` |
| `metricaOmitidaViraNulo` | campo ausente é `null`, não zero |
| `campoDesconhecidoEIgnorado` | agente pode ganhar campos sem derrubar o desktop |
| `jsonInvalidoViraFalhaDeFormato` | JSON truncado → `Malformed` |
| `erroDoAgenteMantemAMensagemDele` | 503 preserva a mensagem do agente |
| `statusInesperadoViraFalhaInesperada` | 500 → `Unexpected` |
| `agenteForaDoArViraIndisponivel` | `IOException` → `Unreachable` |
| `listaDeProcessosVaziaESucesso` | `[]` é sucesso, não erro |
| `processosLevamCriterioELimiteNaConsulta` | `sort=memory&limit=25` chega na URL |

### O teste encontrou um erro real

`jsonInvalidoViraFalhaDeFormato` falhou na primeira execução:

```
expected:<Malformed> but was:<Unexpected>
```

O Ktor **não** propaga `SerializationException`: ele a encapsula em
`ContentConvertException`. Como a classificação só reconhecia a primeira, uma
resposta corrompida caía no ramo genérico — o usuário veria "Falha inesperada
ao falar com o agente" onde a mensagem certa é "O agente respondeu num formato
inesperado". Corrigido incluindo `ContentConvertException` e
`NoTransformationFoundException`.

Esse é exatamente o tipo de defeito que só aparece quando o teste passa pelo
caminho de produção.

---

## 18.4 — ViewModel em tempo virtual

Duas mudanças no código de produção foram necessárias, e as duas melhoraram o
desenho:

### Interfaces `MetricsSource` e `ProcessesSource`

O ViewModel dependia das classes concretas, que dependem do `AgentClient`, que
depende do Ktor. Agora depende de contratos — e a razão de existir dos
repositories (isolar a infraestrutura) passou a valer também para quem os
consome.

### Dispatcher injetável

Os laços rodam em `Dispatchers.Default` para não competir com o desenho. Fixo,
ele escaparia do relógio virtual do teste; injetado, os testes rodam em
microssegundos em vez de esperar segundos reais.

| Teste | Transição |
|---|---|
| `comecaEmLoading` | estado inicial antes de qualquer resposta |
| `loadingViraConectadoComOPrimeiroSnapshot` | Loading → Connected |
| `falhaDeFormatoViraErroComMensagem` | Loading → Error, com "Tentando novamente" |
| `quedaMantemAsMetricasNaTela` | Connected → Disconnected **sem perder os dados** |
| `reconexaoVoltaParaConectado` | Disconnected → Connected |
| `historicoAcumulaOsSnapshots` | três snapshots viram três pontos |
| `falhaEmProcessosNaoMexeNaLigacao` | a etiqueta não oscila entre as duas fontes |
| `trocarCriterioPedeNovaListaNaHora` | o clique não espera o intervalo |

---

## Resultado

```
34 testes KMP  ·  0 falhas
```

Somados aos 60+ testes do agente, cobrem as quatro transições de estado da UI,
os quatro tipos de falha de comunicação, e todos os cálculos que produzem
número na tela.
