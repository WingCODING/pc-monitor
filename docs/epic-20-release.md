# Epic 20 — Documentação e Release

**Status:** concluído
**Data:** 2026-08-30

Cobre as Tasks 20.1 (README), 20.2 (documentação da API) e 20.3 (release
v0.1.0).

---

## 20.1 — README

Escrito para quem chega no repositório sem contexto: o que o programa faz, como
está montado, o que é preciso para compilar, e como rodar — nessa ordem.

Três decisões:

- **Captura de tela logo no começo.** Um monitor de recursos é um programa
  visual; descrever o dashboard em prosa custa três parágrafos e convence menos
  que uma imagem.
- **`scripts/build.sh` + `scripts/start.sh` como caminho principal**, com os
  comandos de desenvolvimento (`go run`, `./gradlew :composeApp:run`) logo
  abaixo. Quem quer só ver funcionando não precisa saber que existem dois
  projetos.
- **As regras de arquitetura estão no README**, não escondidas nos docs de
  epic: são o que um leitor precisa saber antes de mexer no código.

---

## 20.2 — `docs/api.md`

Cada endpoint com método, caminho, exemplo de resposta e erros esperados.

Os exemplos são **respostas reais desta máquina**, capturadas do agente
empacotado — não amostras inventadas. Um exemplo com `"total": 1000000` seria
mais bonito e esconderia que `used + available` não fecha com `total` em
`/memory`, ou que a primeira leitura de `/network` vem zerada.

A documentação registra as três armadilhas que o código resolve, porque quem
consome a API precisa saber que elas existem:

- disco é **por dispositivo**, não por ponto de montagem;
- `cpuPercent` de processo passa de 100 em multithread, de propósito;
- métrica indisponível **some** do snapshot em vez de vir zerada.

---

## 20.3 — Release v0.1.0

Verificação final:

| Item | Estado |
|---|---|
| Backend compila | `go build ./...`, `go vet ./...` limpos |
| Frontend compila | `:composeApp:compileKotlinJvm` |
| Testes do agente | api 88,3 %, collector 90,7 %, config 100 %, service 96,9 % |
| Testes do desktop | 34 testes, 0 falhas |
| CPU, RAM, disco, rede, processos | reais na tela |
| WebSocket | 1 snapshot/s, uma coleta para todos os clientes |
| Reconexão | queda detectada em ~2 s, volta em ~2 s |
| Empacotamento | agente estático + desktop com JVM embutida |
| README | com captura, arquitetura e comandos |

Tag: `v0.1.0`.

---

## O que ficou de fora do MVP — e por quê

| Ideia | Por que não agora |
|---|---|
| GPU | exige biblioteca por fabricante; nada do desenho atual muda para acomodá-la depois |
| Alertas/limiares | precisa de persistência e de notificação do sistema; é produto novo, não polimento |
| Acesso remoto | sem autenticação, expor a lista de processos na rede é irresponsável |
| Windows e macOS | o gopsutil cobre os três, mas "compila" e "está testado" são coisas diferentes, e a validação aqui foi toda em Linux |
| Histórico persistente | 60 pontos em memória resolvem "o que aconteceu no último minuto"; guardar em disco é outro problema |

---

## Retrospectiva das 20 epics

O que apareceu mais de uma vez:

**Toda métrica interessante é um delta.** CPU, rede e processos reportam
contadores acumulados; o valor que interessa é a diferença entre duas leituras
dividida pelo tempo real decorrido. As três implementações erraram de formas
diferentes na primeira versão — wraparound de `uint64` na rede, quantização de
tick de 10 ms nos processos, baseline global compartilhado na CPU.

**Somar sem verificar duplicidade dá números absurdos.** Os subvolumes btrfs
teriam reportado 4 TB num disco de 1 TB. A rede não tem esse problema, e
registrar *por que* ela não tem foi tão útil quanto corrigir o disco.

**Validar na máquina real encontrou o que o teste não pegava.** O percentual
zerado em chamadas consecutivas, o `5%` contra `6 %` no mesmo cartão, as nove
colunas para quatro cartões, o zumbi passando por processo vivo — nenhum
apareceu em teste unitário, todos apareceram ao rodar e olhar.

**Ausência não é zero.** Vale no JSON (`omitempty`), no modelo KMP (anulável),
no histórico dos gráficos (ponto não entra) e na UI ("indisponível" em vez de
"0 %"). É a mesma decisão tomada quatro vezes em camadas diferentes.
