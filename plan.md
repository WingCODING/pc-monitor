# PC Monitor — Plan

## 1. Visão Geral

O **PC Monitor** será um aplicativo desktop para monitoramento em tempo real dos recursos de um computador.

O sistema será dividido em duas aplicações independentes:

- **Backend / Agent:** Go
- **Frontend Desktop:** Kotlin Multiplatform + Compose Multiplatform

O backend será responsável por acessar informações do sistema operacional, processar as métricas e disponibilizá-las por REST e WebSocket.

O frontend será responsável por consumir essas informações e apresentá-las em um dashboard desktop.

---

## 2. Objetivo do Projeto

Criar um software capaz de monitorar:

- CPU
- Memória RAM
- Discos
- Rede
- Processos
- Informações gerais do sistema
- Uptime

As métricas deverão ser atualizadas em tempo real no aplicativo desktop.

---

## 3. MVP

### Backend

- Servidor HTTP em Go
- Endpoint de health check
- Monitoramento de CPU
- Monitoramento de RAM
- Monitoramento de discos
- Monitoramento de rede
- Informações gerais do sistema
- Lista de processos
- REST API
- WebSocket para métricas em tempo real

### Frontend

- Aplicativo desktop KMP
- Compose Multiplatform
- Dashboard principal
- Card de CPU
- Card de RAM
- Card de disco
- Card de rede
- Informações do sistema
- Lista de processos
- Atualização em tempo real
- Estados de conexão, carregamento e erro

---

## 4. Fora do MVP

Não serão implementados inicialmente:

- Cloud
- Login
- Conta de usuário
- Controle remoto do computador
- Banco de dados remoto
- Aplicativo Android
- Aplicativo iOS
- Monitoramento remoto via Internet
- GPU avançada
- Temperatura de hardware
- Integração específica com NVIDIA/AMD
- Sistema complexo de alertas
- Persistência histórica de longo prazo

Esses recursos poderão ser adicionados posteriormente.

---

## 5. Arquitetura Geral

```text
┌─────────────────────────────────────┐
│              PC                     │
│                                     │
│       Sistema Operacional           │
│              ↓                      │
│        Go Monitoring Agent          │
│              ↓                      │
│       Collectors / Services         │
│              ↓                      │
│       REST API / WebSocket          │
└──────────────┬──────────────────────┘
               │
               │ localhost
               │
               ↓
┌─────────────────────────────────────┐
│       KMP Desktop Application       │
│                                     │
│       Network Layer                 │
│              ↓                      │
│          Repository                 │
│              ↓                      │
│          StateFlow                  │
│              ↓                      │
│          ViewModel                  │
│              ↓                      │
│      Compose Multiplatform          │
└─────────────────────────────────────┘
```

---

## 6. Estrutura do Repositório

```text
pc-monitor/
│
├── agent/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   │
│   ├── internal/
│   │   ├── api/
│   │   ├── collector/
│   │   ├── model/
│   │   └── service/
│   │
│   ├── go.mod
│   └── go.sum
│
├── desktop/
│   ├── composeApp/
│   ├── gradle/
│   ├── build.gradle.kts
│   └── settings.gradle.kts
│
├── docs/
│
├── README.md
├── plan.md
└── spec-driven-tasks.md
```

---

## 7. Responsabilidades do Backend

O agente Go será responsável por:

1. Consultar o sistema operacional.
2. Coletar métricas.
3. Normalizar os dados.
4. Transformar dados internos em modelos da API.
5. Disponibilizar REST API.
6. Enviar métricas em tempo real via WebSocket.
7. Tratar erros sem encerrar o processo.
8. Manter baixo consumo de CPU e memória.

---

## 8. Collectors

Cada categoria deverá possuir seu próprio collector.

```text
internal/collector/
├── cpu.go
├── memory.go
├── disk.go
├── network.go
├── process.go
└── system.go
```

O restante da aplicação não deverá conhecer detalhes específicos de `/proc`, APIs nativas ou bibliotecas de monitoramento.

---

## 9. CPU Collector

Responsável por retornar:

- Modelo da CPU
- Quantidade de cores
- Quantidade de threads
- Uso total em percentual

Exemplo:

```json
{
  "model": "AMD Ryzen 7 7800X3D",
  "cores": 8,
  "threads": 16,
  "usagePercent": 24.8
}
```

---

## 10. Memory Collector

Responsável por retornar:

- Memória total
- Memória utilizada
- Memória disponível
- Percentual utilizado

Exemplo:

```json
{
  "total": 34359738368,
  "used": 17179869184,
  "available": 17179869184,
  "usagePercent": 50.0
}
```

Os valores internos deverão usar **bytes**. A UI será responsável por converter para MB, GB ou TB.

---

## 11. Disk Collector

Responsável por identificar discos e partições.

Dados:

- Nome
- Mount point
- Total
- Utilizado
- Disponível
- Percentual

Exemplo:

```json
[
  {
    "name": "/dev/nvme0n1p2",
    "mountPoint": "/",
    "total": 1000204886016,
    "used": 520204886016,
    "free": 480000000000,
    "usagePercent": 52.0
  }
]
```

---

## 12. Network Collector

Responsável por monitorar interfaces de rede.

Dados:

- Nome da interface
- Bytes recebidos
- Bytes enviados
- Download atual em bytes por segundo
- Upload atual em bytes por segundo

A velocidade atual será calculada comparando duas leituras realizadas em intervalos conhecidos.

---

## 13. Process Collector

Responsável por listar processos em execução.

Dados mínimos:

- PID
- Nome
- CPU %
- Memória usada

Exemplo:

```json
{
  "pid": 4321,
  "name": "idea",
  "cpuPercent": 8.2,
  "memory": 1879048192
}
```

O backend deverá evitar coleta excessivamente cara de processos.

---

## 14. System Collector

Responsável por informações gerais.

Dados:

- Hostname
- Sistema operacional
- Versão
- Arquitetura
- Uptime

Exemplo:

```json
{
  "hostname": "desktop",
  "os": "linux",
  "architecture": "amd64",
  "uptimeSeconds": 58232
}
```

---

## 15. REST API

Base inicial:

```text
http://localhost:8080
```

Endpoints:

```text
GET /health

GET /api/v1/system
GET /api/v1/cpu
GET /api/v1/memory
GET /api/v1/disks
GET /api/v1/network
GET /api/v1/processes
GET /api/v1/metrics
```

---

## 16. Endpoint Agregado

O endpoint principal do dashboard será:

```text
GET /api/v1/metrics
```

Exemplo:

```json
{
  "timestamp": "2026-08-30T16:00:00Z",
  "cpu": {
    "usagePercent": 32.8
  },
  "memory": {
    "total": 34359738368,
    "used": 17179869184,
    "usagePercent": 50.0
  },
  "network": {
    "downloadBytesPerSecond": 5242880,
    "uploadBytesPerSecond": 1048576
  },
  "uptimeSeconds": 53214
}
```

---

## 17. WebSocket

Depois da REST API estar estável:

```text
ws://localhost:8080/api/v1/ws/metrics
```

Fluxo:

```text
Collectors
    ↓
Metrics Service
    ↓
WebSocket Hub
    ↓
KMP Desktop
```

Intervalo inicial:

```text
1 segundo
```

O intervalo deverá ser configurável posteriormente.

---

## 18. Frontend — Kotlin Multiplatform

Tecnologias:

- Kotlin Multiplatform
- Compose Multiplatform
- Kotlin Coroutines
- StateFlow
- Ktor Client
- Kotlin Serialization

---

## 19. Arquitetura do Frontend

```text
UI
↓
ViewModel
↓
Repository
↓
Network Client
↓
Go API
```

Estrutura sugerida:

```text
desktop/
└── composeApp/
    └── src/
        └── commonMain/
            └── kotlin/
                └── pcmonitor/
                    ├── data/
                    ├── network/
                    ├── repository/
                    ├── model/
                    ├── viewmodel/
                    └── ui/
```

---

## 20. Models KMP

Os modelos deverão representar os contratos do backend.

Exemplo:

```kotlin
@Serializable
data class CpuMetrics(
    val usagePercent: Double
)

@Serializable
data class MemoryMetrics(
    val total: Long,
    val used: Long,
    val available: Long,
    val usagePercent: Double
)
```

---

## 21. Repository

Interface conceitual:

```text
MetricsRepository

suspend getMetrics()
observeMetrics()
suspend getProcesses()
```

A UI e o ViewModel não deverão acessar HTTP ou WebSocket diretamente.

---

## 22. State Management

```text
Go WebSocket
     ↓
Repository
     ↓
Flow
     ↓
StateFlow
     ↓
ViewModel
     ↓
Compose
```

O ViewModel deverá transformar dados de infraestrutura em estado de UI.

---

## 23. Dashboard

Tela inicial esperada:

```text
┌──────────────────────────────────────────┐
│ PC MONITOR                    Connected  │
│                                          │
│ CPU             RAM            DISK      │
│ 28%             52%            61%       │
│ ███░░           █████░         ██████░   │
│                                          │
│ NETWORK                                  │
│ ↓ 25.4 MB/s            ↑ 3.2 MB/s        │
│                                          │
│ CPU HISTORY                              │
│      ╭────╮                              │
│ ─────╯    ╰──────                       │
│                                          │
│ PROCESSES                                │
│ Chrome       12%        850 MB           │
│ IntelliJ      8%        2.1 GB           │
│ Discord       3%        420 MB           │
└──────────────────────────────────────────┘
```

---

## 24. Estados da UI

A aplicação deverá possuir estados explícitos:

```text
Loading
Connected
Disconnected
Error
```

Também deverá distinguir:

- backend indisponível;
- erro de parsing;
- erro de WebSocket;
- dados parcialmente indisponíveis.

---

## 25. Tratamento de Erros

O frontend deve continuar aberto caso o backend fique indisponível.

Exemplo:

```text
Agent disconnected
Retrying connection...
```

O backend também não deverá encerrar porque uma métrica individual falhou.

---

## 26. Histórico

O MVP manterá histórico curto em memória.

Configuração inicial:

```text
1 ponto por segundo
60 pontos
60 segundos de histórico
```

Será utilizado para gráficos de CPU, RAM e rede.

Sem banco de dados no MVP.

---

## 27. Segurança

No MVP:

```text
bind: 127.0.0.1
```

Não haverá autenticação.

Se futuramente o agente puder ser acessado pela LAN ou Internet, serão necessários:

- autenticação;
- TLS;
- autorização;
- device pairing;
- rate limiting;
- política explícita de interfaces de rede.

---

## 28. Performance

Metas iniciais:

- Coleta principal aproximadamente 1 vez por segundo
- Baixo consumo de CPU do agente
- Baixo consumo de RAM do agente
- UI responsiva
- Coleta de processos otimizada
- Nenhuma operação pesada no thread principal da UI

---

## 29. Compatibilidade

Primeiro alvo:

```text
Linux
```

Segundo alvo:

```text
Windows
```

Futuro:

```text
macOS
```

O projeto deverá minimizar código específico por sistema operacional.

---

## 30. Testes

### Backend

- Unit tests dos cálculos
- Testes dos services
- Testes dos handlers HTTP
- Validação de status codes
- Validação de serialização JSON

### Frontend

- Testes de repository
- Testes de ViewModel
- Testes de conversão/formatação
- Testes dos estados de conexão

---

## 31. Observabilidade do Próprio Projeto

O agente deverá possuir logs simples:

```text
INFO
WARN
ERROR
```

Exemplos:

```text
INFO server started on 127.0.0.1:8080
WARN cpu collector temporarily unavailable
ERROR websocket connection failed
```

Logs não devem imprimir informações sensíveis desnecessárias.

---

## 32. Roadmap em 10 Partes

### Parte 1 — Bootstrap e arquitetura
Criar repositório, módulos e documentação.

### Parte 2 — Servidor Go
Criar HTTP server e `/health`.

### Parte 3 — CPU e RAM
Implementar primeiros collectors reais.

### Parte 4 — Disco, rede e sistema
Expandir monitoramento.

### Parte 5 — Processos
Adicionar lista e métricas de processos.

### Parte 6 — KMP Desktop
Construir dashboard usando dados mockados.

### Parte 7 — Integração REST
Conectar KMP ao Go.

### Parte 8 — Realtime
Implementar WebSocket.

### Parte 9 — Histórico e UX
Gráficos, estados, filtros e reconexão.

### Parte 10 — Distribuição
Builds, instalador, documentação e release.

---

## 33. Definition of Done — MVP

O MVP será considerado concluído quando:

- [ ] O Go Agent iniciar corretamente.
- [ ] `/health` responder HTTP 200.
- [ ] CPU for exibida em tempo real.
- [ ] RAM for exibida em tempo real.
- [ ] Disco for exibido.
- [ ] Rede for monitorada.
- [ ] Processos forem listados.
- [ ] Informações do sistema forem exibidas.
- [ ] A REST API funcionar localmente.
- [ ] WebSocket atualizar as métricas.
- [ ] O KMP Desktop conectar automaticamente.
- [ ] A interface lidar com perda de conexão.
- [ ] O histórico curto funcionar.
- [ ] O frontend puder ser executado como aplicativo desktop.
- [ ] O backend puder ser compilado como executável.
- [ ] README possuir instruções de desenvolvimento e execução.
- [ ] O projeto possuir uma versão de release identificável.
