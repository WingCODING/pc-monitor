# Epic 19 — Build e Distribuição

**Status:** concluído
**Data:** 2026-08-30

Cobre as Tasks 19.1 (build do agente), 19.2 (empacotamento do desktop) e 19.3
(estratégia de inicialização conjunta).

---

## O que foi feito

| Arquivo | Papel |
|---|---|
| `scripts/build.sh` | Compila os dois e deixa tudo em `dist/` |
| `scripts/start.sh` | Sobe agente e desktop juntos, e encerra os dois juntos |
| `desktop/composeApp/build.gradle.kts` | `nativeDistributions` |

---

## 19.1 — Binário do agente sem dependências

```bash
CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o dist/pc-monitor-agent ./cmd/server
```

`CGO_ENABLED=0` porque o gopsutil lê `/proc` em Go puro no Linux: sem CGO o
binário sai **estaticamente ligado** e não depende da libc da máquina que
compilou.

```
$ file dist/pc-monitor-agent
ELF 64-bit LSB executable, x86-64, statically linked, Go BuildID=...
```

`-trimpath` tira os caminhos absolutos do build, e `-s -w` descartam a tabela
de símbolos e o DWARF.

---

## 19.2 — Desktop com JVM embutida

`createDistributable` é o alvo principal, e não `packageDeb`: ele gera um
diretório com a aplicação e uma JVM enxuta feita pelo `jlink`, **sem depender
de ferramenta de empacotamento do sistema**. Numa máquina Arch não há `dpkg`
nem `rpmbuild`, e exigir um deles para poder rodar o aplicativo seria trocar
uma dependência (JDK) por outra pior.

```
dist/desktop/bin/    ← lançador
dist/desktop/lib/    ← runtime + classes  (147 MB no total)
```

### `modules(...)` não é enfeite

```kotlin
modules("java.net.http", "jdk.crypto.ec", "java.management")
```

O `jlink` monta a JVM só com os módulos declarados. Sem estes, o aplicativo
compila, empacota e **abre** — e só quebra ao tentar falar com o agente, porque
o suporte a rede/TLS ficou de fora. O erro não aparece no desenvolvimento, só
no artefato empacotado.

---

## 19.3 — Estratégia de inicialização conjunta

Das opções levantadas na spec — lançador, processo filho, serviço separado —
o MVP usa o **processo filho**: `scripts/start.sh` sobe o agente, guarda o PID
e abre a janela.

```bash
trap 'kill "$agente_pid" 2>/dev/null || true' EXIT INT TERM
```

Sem o `trap`, fechar a janela deixaria um agente escutando na porta sem
ninguém para usá-lo — e a próxima execução falharia com "porta em uso".

O script também repassa a porta ao desktop pela variável `PCMON_AGENT_URL`.
Sem isso, `PCMON_PORT=8123` subiria o agente na 8123 e o desktop continuaria
procurando na 8080.

### Um erro encontrado ao testar a porta ocupada

A verificação de "o agente subiu?" usava `kill -0 $pid`. Com a porta ocupada, o
agente morre imediatamente — mas o processo continua na tabela como **zumbi**
até o shell recolhê-lo, e `kill -0` responde que ele existe. Resultado: o
script seguia adiante e abria a janela.

Pior: a checagem seguinte (`curl /health`) **passava**, porque quem respondia na
porta era o outro agente, o que já estava rodando. A falha ficava perfeitamente
escondida.

A verificação passou a olhar o estado do processo:

```bash
estado="$(ps -o stat= -p "$agente_pid" 2>/dev/null || true)"
[[ -n "$estado" && "$estado" != Z* ]]
```

Depois da correção:

```
$ scripts/start.sh
ERROR msg="agente encerrado com erro" error="bind: address already in use"
o agente encerrou ao iniciar; verifique se a porta 8080 está livre
$ echo $?
1
```

### Frontend detecta agente ausente

Requisito já atendido desde a Epic 15: sem agente, a janela abre normalmente
com a etiqueta `Desconectado`, a faixa explicando que a reconexão é automática,
e volta sozinha quando o agente sobe.

---

## Validação

| Cenário | Resultado |
|---|---|
| `scripts/build.sh` do zero | agente estático + distribuível de 147 MB |
| `scripts/start.sh` | janela abre, `Conectado`, dados reais |
| Fechar a janela | agente encerra junto; porta 8080 livre |
| Porta ocupada | mensagem no terminal, saída 1, janela não abre |
| `PCMON_PORT=8123 scripts/start.sh` | agente na 8123 e desktop conectado nela |

O aplicativo foi executado a partir de `dist/`, fora da IDE e com `JAVA_HOME`
desativado no ambiente.
