# PC Monitor — Spec Driven Tasks

Este documento transforma o `plan.md` em tarefas implementáveis e verificáveis.

## Regras de Execução

1. Não avançar para uma Epic dependente enquanto seus critérios mínimos não estiverem atendidos.
2. Cada task deve produzir um resultado verificável.
3. Código de infraestrutura não deve vazar para a UI.
4. Collectors não devem depender da camada HTTP.
5. O frontend não deve acessar HTTP/WebSocket diretamente a partir de Composables.
6. O MVP usa apenas `localhost`.
7. Mudanças no contrato da API devem ser refletidas no backend e nos models KMP.

---

# Epic 01 — Bootstrap do Projeto

## Task 1.1 — Criar repositório

Criar:

```text
pc-monitor/
├── agent/
├── desktop/
├── docs/
├── plan.md
├── spec-driven-tasks.md
└── README.md
```

### Acceptance Criteria

- [ ] Repositório criado
- [ ] Git inicializado
- [ ] `.gitignore` criado
- [ ] `agent/` criado
- [ ] `desktop/` criado
- [ ] `docs/` criado
- [ ] `plan.md` presente
- [ ] `spec-driven-tasks.md` presente

### Depends On

Nenhuma.

---

## Task 1.2 — Criar módulo Go

Dentro de `agent/`:

```text
go mod init pc-monitor-agent
```

### Acceptance Criteria

- [ ] `go.mod` existe
- [ ] Projeto compila
- [ ] `go run .` ou equivalente funciona
- [ ] IntelliJ reconhece o módulo Go

### Depends On

- Task 1.1

---

## Task 1.3 — Definir layout inicial do backend

Criar:

```text
agent/
├── cmd/
│   └── server/
│       └── main.go
│
└── internal/
    ├── api/
    ├── collector/
    ├── model/
    └── service/
```

### Acceptance Criteria

- [ ] `main.go` contém apenas bootstrap do servidor
- [ ] Collectors ficam fora de `main.go`
- [ ] Models não dependem de HTTP
- [ ] Services não dependem do frontend

### Depends On

- Task 1.2

---

# Epic 02 — Servidor HTTP

## Task 2.1 — Criar servidor HTTP

Servidor inicial:

```text
127.0.0.1:8080
```

### Acceptance Criteria

- [ ] Aplicação inicia sem erro
- [ ] Porta 8080 é utilizada por padrão
- [ ] Bind é feito em localhost
- [ ] Erros de inicialização são tratados
- [ ] Log informa endereço do servidor

### Depends On

- Task 1.3

---

## Task 2.2 — Criar Health Endpoint

Endpoint:

```text
GET /health
```

Response:

```json
{
  "status": "ok"
}
```

### Acceptance Criteria

- [ ] HTTP 200
- [ ] `Content-Type: application/json`
- [ ] JSON válido
- [ ] `status = "ok"`
- [ ] Endpoint pode ser aberto no navegador
- [ ] Handler possui teste

### Depends On

- Task 2.1

---

# Epic 03 — Contratos e Base dos Collectors

## Task 3.1 — Criar models básicos

Criar models para:

```text
CpuMetrics
MemoryMetrics
DiskMetrics
NetworkMetrics
ProcessMetrics
SystemMetrics
DashboardMetrics
```

### Acceptance Criteria

- [ ] Models estão em `internal/model`
- [ ] JSON tags estão definidas
- [ ] Nomes JSON usam padrão consistente
- [ ] Bytes são representados por inteiros adequados
- [ ] Percentuais usam `float64`

### Depends On

- Task 2.2

---

## Task 3.2 — Definir interfaces dos collectors

Criar contratos independentes da implementação.

Exemplo conceitual:

```go
type CPUCollector interface {
    Collect() (model.CpuMetrics, error)
}
```

### Acceptance Criteria

- [ ] Cada collector possui responsabilidade única
- [ ] Services dependem de interfaces
- [ ] Implementações podem ser substituídas em testes

### Depends On

- Task 3.1

---

# Epic 04 — CPU

## Task 4.1 — Criar CPU Collector

Coletar:

- modelo;
- cores;
- threads;
- uso percentual.

### Acceptance Criteria

- [ ] CPU real pode ser obtida
- [ ] Uso fica entre 0 e 100
- [ ] Cores > 0
- [ ] Threads > 0
- [ ] Erros retornam `error`
- [ ] Collector não executa `panic`

### Depends On

- Task 3.2

---

## Task 4.2 — Criar CPU Service

O service deverá coordenar o collector e preparar o model para a API.

### Acceptance Criteria

- [ ] Service depende da interface do collector
- [ ] Falhas são propagadas ou convertidas em erro de domínio
- [ ] Service possui teste unitário

### Depends On

- Task 4.1

---

## Task 4.3 — Criar CPU Endpoint

```text
GET /api/v1/cpu
```

Exemplo:

```json
{
  "model": "AMD Ryzen 7 7800X3D",
  "cores": 8,
  "threads": 16,
  "usagePercent": 24.8
}
```

### Acceptance Criteria

- [ ] HTTP 200 em sucesso
- [ ] JSON segue o contrato
- [ ] Dados são reais
- [ ] Erros possuem status HTTP adequado

### Depends On

- Task 4.2

---

# Epic 05 — Memória RAM

## Task 5.1 — Criar Memory Collector

Coletar:

```text
total
used
available
usagePercent
```

### Acceptance Criteria

- [ ] Valores em bytes
- [ ] `total > 0`
- [ ] `used >= 0`
- [ ] `available >= 0`
- [ ] Percentual entre 0 e 100
- [ ] Sem panic

### Depends On

- Task 3.2

---

## Task 5.2 — Criar Memory Service

### Acceptance Criteria

- [ ] Service recebe collector por dependência
- [ ] Cálculo de percentual é testado
- [ ] Dados inconsistentes são tratados

### Depends On

- Task 5.1

---

## Task 5.3 — Criar Memory Endpoint

```text
GET /api/v1/memory
```

### Acceptance Criteria

- [ ] HTTP 200
- [ ] JSON válido
- [ ] Dados reais
- [ ] Contrato documentado

### Depends On

- Task 5.2

---

# Epic 06 — Endpoint Agregado

## Task 6.1 — Criar Metrics Service

Combinar inicialmente:

- CPU
- RAM
- uptime

### Acceptance Criteria

- [ ] Uma chamada retorna snapshot consistente
- [ ] Falha de uma métrica não causa crash
- [ ] Timestamp acompanha o snapshot
- [ ] Service possui teste

### Depends On

- Task 4.2
- Task 5.2

---

## Task 6.2 — Criar `/api/v1/metrics`

```text
GET /api/v1/metrics
```

### Acceptance Criteria

- [ ] HTTP 200
- [ ] CPU presente
- [ ] RAM presente
- [ ] Timestamp presente
- [ ] Uptime presente quando disponível
- [ ] JSON segue model `DashboardMetrics`

### Depends On

- Task 6.1

---

# Epic 07 — Informações do Sistema

## Task 7.1 — Criar System Collector

Coletar:

```text
hostname
os
osVersion
architecture
uptimeSeconds
```

### Acceptance Criteria

- [ ] Hostname retornado
- [ ] OS identificado
- [ ] Arquitetura identificada
- [ ] Uptime >= 0
- [ ] Erros tratados

### Depends On

- Task 3.2

---

## Task 7.2 — Criar System Endpoint

```text
GET /api/v1/system
```

### Acceptance Criteria

- [ ] HTTP 200
- [ ] JSON válido
- [ ] Dados correspondem à máquina local

### Depends On

- Task 7.1

---

## Task 7.3 — Integrar System ao Metrics Service

### Acceptance Criteria

- [ ] Uptime vem do System Collector
- [ ] Snapshot continua funcionando com erro parcial
- [ ] Testes atualizados

### Depends On

- Task 6.1
- Task 7.1

---

# Epic 08 — Discos

## Task 8.1 — Criar Disk Collector

Coletar por partição:

```text
name
mountPoint
total
used
free
usagePercent
```

### Acceptance Criteria

- [ ] Pelo menos a partição principal é detectada
- [ ] Valores são retornados em bytes
- [ ] Percentual fica entre 0 e 100
- [ ] Filesystems inválidos são ignorados ou tratados
- [ ] Nenhum panic

### Depends On

- Task 3.2

---

## Task 8.2 — Criar Disk Endpoint

```text
GET /api/v1/disks
```

### Acceptance Criteria

- [ ] HTTP 200
- [ ] Retorna array JSON
- [ ] Partição principal presente
- [ ] Dados reais

### Depends On

- Task 8.1

---

## Task 8.3 — Integrar resumo de disco no Metrics Service

### Acceptance Criteria

- [ ] Dashboard possui informação de disco
- [ ] Contrato atualizado
- [ ] Testes atualizados

### Depends On

- Task 6.1
- Task 8.1

---

# Epic 09 — Rede

## Task 9.1 — Criar Network Collector

Coletar:

```text
interfaceName
bytesReceived
bytesSent
```

### Acceptance Criteria

- [ ] Interfaces são detectadas
- [ ] Bytes recebidos >= 0
- [ ] Bytes enviados >= 0
- [ ] Loopback pode ser filtrado
- [ ] Collector é testável

### Depends On

- Task 3.2

---

## Task 9.2 — Calcular velocidade atual

Calcular:

```text
downloadBytesPerSecond
uploadBytesPerSecond
```

a partir da diferença entre snapshots.

### Acceptance Criteria

- [ ] Cálculo considera intervalo real
- [ ] Resultado nunca é negativo
- [ ] Primeira leitura possui comportamento definido
- [ ] Reset de contador é tratado

### Depends On

- Task 9.1

---

## Task 9.3 — Criar Network Endpoint

```text
GET /api/v1/network
```

### Acceptance Criteria

- [ ] HTTP 200
- [ ] Retorna velocidade atual
- [ ] Retorna contadores acumulados
- [ ] JSON válido

### Depends On

- Task 9.2

---

## Task 9.4 — Integrar rede no Metrics Service

### Acceptance Criteria

- [ ] Download presente
- [ ] Upload presente
- [ ] Endpoint agregado continua estável

### Depends On

- Task 6.1
- Task 9.2

---

# Epic 10 — Processos

## Task 10.1 — Criar Process Collector

Coletar:

```text
pid
name
cpuPercent
memory
```

### Acceptance Criteria

- [ ] Lista contém processos reais
- [ ] PID > 0
- [ ] Nome disponível quando permitido pelo SO
- [ ] Processos que terminam durante a coleta não causam crash

### Depends On

- Task 3.2

---

## Task 10.2 — Ordenação e limite

Adicionar opções internas para:

- ordenar por CPU;
- ordenar por memória;
- limitar quantidade.

### Acceptance Criteria

- [ ] Ordenação por CPU funciona
- [ ] Ordenação por memória funciona
- [ ] Limite evita retorno excessivo
- [ ] Critérios possuem testes

### Depends On

- Task 10.1

---

## Task 10.3 — Criar Process Endpoint

```text
GET /api/v1/processes
```

### Acceptance Criteria

- [ ] HTTP 200
- [ ] Retorna array
- [ ] Dados reais
- [ ] Ordenação padrão definida
- [ ] Não bloqueia a API por tempo excessivo

### Depends On

- Task 10.2

---

# Epic 11 — Robustez do Backend

## Task 11.1 — Padronizar respostas de erro

Exemplo:

```json
{
  "error": "collector_unavailable",
  "message": "CPU metrics are temporarily unavailable"
}
```

### Acceptance Criteria

- [ ] Erros possuem formato consistente
- [ ] Status HTTP apropriados
- [ ] Logs registram contexto
- [ ] Stack trace não é enviado para o cliente

### Depends On

- Epics 04 a 10

---

## Task 11.2 — Adicionar logging

Níveis:

```text
INFO
WARN
ERROR
```

### Acceptance Criteria

- [ ] Inicialização do servidor é registrada
- [ ] Falhas de collector são registradas
- [ ] Logs não são excessivos a cada segundo
- [ ] Nenhuma informação sensível desnecessária

### Depends On

- Task 11.1

---

## Task 11.3 — Configuração básica

Permitir configurar:

```text
host
port
collectionInterval
```

via constantes, arquivo ou variáveis de ambiente.

### Acceptance Criteria

- [ ] Defaults funcionam sem configuração
- [ ] Host padrão é `127.0.0.1`
- [ ] Porta padrão é `8080`
- [ ] Intervalo padrão é `1s`

### Depends On

- Task 11.2

---

# Epic 12 — Bootstrap do KMP Desktop

## Task 12.1 — Criar projeto Kotlin Multiplatform

Criar módulo desktop com Compose Multiplatform.

### Acceptance Criteria

- [ ] Projeto abre no IntelliJ
- [ ] Gradle sincroniza sem erro
- [ ] Aplicação desktop inicia
- [ ] Janela Compose é exibida

### Depends On

- Epic 02 concluída no backend
- Backend pode continuar evoluindo em paralelo

---

## Task 12.2 — Adicionar dependências

Adicionar:

- Ktor Client
- Kotlin Serialization
- Coroutines
- Compose Multiplatform

### Acceptance Criteria

- [ ] Build funciona
- [ ] Imports resolvem
- [ ] Nenhuma dependência duplicada desnecessária

### Depends On

- Task 12.1

---

## Task 12.3 — Criar estrutura do frontend

```text
model/
network/
repository/
viewmodel/
ui/
```

### Acceptance Criteria

- [ ] Camadas separadas
- [ ] Composables não conhecem Ktor Client
- [ ] ViewModel não possui lógica de HTTP

### Depends On

- Task 12.2

---

# Epic 13 — UI Mockada

## Task 13.1 — Criar shell do dashboard

Criar:

- header;
- status de conexão;
- área de métricas;
- área de processos.

### Acceptance Criteria

- [ ] Dashboard abre
- [ ] Layout se adapta ao redimensionamento
- [ ] Nenhuma integração backend necessária ainda

### Depends On

- Task 12.3

---

## Task 13.2 — Criar MetricCard

Componente reutilizável para:

```text
CPU
RAM
Disk
```

### Acceptance Criteria

- [ ] Título configurável
- [ ] Valor configurável
- [ ] Percentual suportado
- [ ] Reutilizado nas métricas principais

### Depends On

- Task 13.1

---

## Task 13.3 — Criar NetworkCard

### Acceptance Criteria

- [ ] Exibe download
- [ ] Exibe upload
- [ ] Formatação de bytes/s funciona

### Depends On

- Task 13.1

---

## Task 13.4 — Criar tabela de processos mockada

### Acceptance Criteria

- [ ] PID exibido
- [ ] Nome exibido
- [ ] CPU exibida
- [ ] Memória exibida
- [ ] Lista suporta múltiplas linhas

### Depends On

- Task 13.1

---

# Epic 14 — Integração REST KMP ↔ Go

## Task 14.1 — Criar models KMP

Espelhar contratos REST.

### Acceptance Criteria

- [ ] `@Serializable` aplicado
- [ ] Campos correspondem ao backend
- [ ] Tipos correspondem ao JSON

### Depends On

- Task 12.2
- Contratos REST estabilizados

---

## Task 14.2 — Criar HTTP Client

Configurar Ktor Client.

### Acceptance Criteria

- [ ] Base URL configurada
- [ ] JSON instalado
- [ ] Timeout definido
- [ ] Erros de conexão tratados

### Depends On

- Task 14.1

---

## Task 14.3 — Criar MetricsRepository

### Acceptance Criteria

- [ ] Repository possui `getMetrics()`
- [ ] UI não conhece Ktor
- [ ] Erros são convertidos para resultado de domínio

### Depends On

- Task 14.2

---

## Task 14.4 — Criar ProcessesRepository

### Acceptance Criteria

- [ ] `getProcesses()` funciona
- [ ] Parsing funciona
- [ ] Lista vazia é tratada

### Depends On

- Task 14.2

---

## Task 14.5 — Criar MetricsViewModel

### Acceptance Criteria

- [ ] Expõe `StateFlow`
- [ ] Possui Loading
- [ ] Possui Connected
- [ ] Possui Error/Disconnected
- [ ] Sem chamada HTTP em Composable

### Depends On

- Task 14.3

---

## Task 14.6 — Substituir mocks por dados reais

### Acceptance Criteria

- [ ] CPU real aparece na UI
- [ ] RAM real aparece na UI
- [ ] Disco real aparece
- [ ] Rede real aparece
- [ ] Processos reais aparecem

### Depends On

- Task 14.4
- Task 14.5

---

# Epic 15 — WebSocket e Realtime

## Task 15.1 — Criar WebSocket endpoint no Go

```text
ws://127.0.0.1:8080/api/v1/ws/metrics
```

### Acceptance Criteria

- [ ] Cliente consegue conectar
- [ ] JSON válido enviado
- [ ] Snapshot enviado aproximadamente a cada 1 segundo
- [ ] Cliente desconectado é removido corretamente

### Depends On

- Metrics Service completo

---

## Task 15.2 — Criar loop de publicação

### Acceptance Criteria

- [ ] Uma coleta alimenta os clientes conectados
- [ ] Não cria collector novo por cliente
- [ ] Cancelamento funciona
- [ ] Nenhum goroutine leak evidente

### Depends On

- Task 15.1

---

## Task 15.3 — Criar WebSocket client no KMP

### Acceptance Criteria

- [ ] Conecta ao endpoint
- [ ] Desserializa mensagens
- [ ] Expõe `Flow<DashboardMetrics>`
- [ ] Fecha conexão corretamente

### Depends On

- Task 15.1
- Task 14.2

---

## Task 15.4 — Integrar realtime ao repository

### Acceptance Criteria

- [ ] `observeMetrics()` exposto
- [ ] ViewModel recebe atualizações
- [ ] Polling REST não é usado continuamente

### Depends On

- Task 15.3

---

## Task 15.5 — Reconexão

### Acceptance Criteria

- [ ] Frontend detecta desconexão
- [ ] UI mostra estado Disconnected
- [ ] Tentativa de reconexão ocorre com intervalo controlado
- [ ] Reconexão não cria conexões duplicadas

### Depends On

- Task 15.4

---

# Epic 16 — Histórico e Gráficos

## Task 16.1 — Criar buffer de histórico

Configuração inicial:

```text
60 snapshots
```

### Acceptance Criteria

- [ ] Máximo de 60 pontos
- [ ] Dados mais antigos são descartados
- [ ] Atualização não bloqueia UI

### Depends On

- Task 15.4

---

## Task 16.2 — Gráfico de CPU

### Acceptance Criteria

- [ ] Exibe últimos 60 segundos
- [ ] Atualiza em tempo real
- [ ] Valores ficam em escala 0–100

### Depends On

- Task 16.1

---

## Task 16.3 — Gráfico de RAM

### Acceptance Criteria

- [ ] Exibe histórico
- [ ] Escala coerente
- [ ] Atualização suave

### Depends On

- Task 16.1

---

## Task 16.4 — Gráfico de rede

### Acceptance Criteria

- [ ] Download e upload representados
- [ ] Unidades formatadas
- [ ] Sem valores negativos

### Depends On

- Task 16.1

---

# Epic 17 — UX e Formatação

## Task 17.1 — Formatadores

Criar helpers para:

- bytes;
- bytes/s;
- percentuais;
- uptime.

### Acceptance Criteria

- [ ] `1024` pode virar `1 KB`
- [ ] GB/TB formatados corretamente
- [ ] Percentuais possuem precisão consistente
- [ ] Uptime fica legível

### Depends On

- Epic 14

---

## Task 17.2 — Estados visuais

### Acceptance Criteria

- [ ] Loading claramente visível
- [ ] Connected visível
- [ ] Disconnected visível
- [ ] Error possui mensagem útil
- [ ] Dashboard não desaparece durante reconexão

### Depends On

- Task 15.5

---

## Task 17.3 — Filtro de processos

Permitir pesquisar por nome.

### Acceptance Criteria

- [ ] Pesquisa case-insensitive
- [ ] Lista atualiza corretamente
- [ ] Campo vazio exibe lista normal

### Depends On

- Task 14.6

---

## Task 17.4 — Ordenação de processos na UI

### Acceptance Criteria

- [ ] Por CPU
- [ ] Por memória
- [ ] Ordem crescente/decrescente definida

### Depends On

- Task 17.3

---

# Epic 18 — Testes

## Task 18.1 — Testes de cálculos Go

Cobrir:

- percentuais;
- velocidade de rede;
- tratamento de deltas;
- agregações.

### Acceptance Criteria

- [ ] Casos normais
- [ ] Zero
- [ ] Valores inválidos
- [ ] Reset de contador

---

## Task 18.2 — Testes de handlers Go

### Acceptance Criteria

- [ ] `/health`
- [ ] `/api/v1/cpu`
- [ ] `/api/v1/memory`
- [ ] `/api/v1/metrics`
- [ ] Respostas de erro

### Depends On

- Endpoints correspondentes

---

## Task 18.3 — Testes de repository KMP

### Acceptance Criteria

- [ ] Sucesso
- [ ] JSON inválido
- [ ] Backend indisponível
- [ ] Lista vazia

### Depends On

- Epic 14

---

## Task 18.4 — Testes de ViewModel

### Acceptance Criteria

- [ ] Loading → Connected
- [ ] Loading → Error
- [ ] Connected → Disconnected
- [ ] Reconexão → Connected

### Depends On

- Epic 15

---

# Epic 19 — Build e Distribuição

## Task 19.1 — Build do agente Go

Gerar executável para o sistema alvo.

### Acceptance Criteria

- [ ] `go build` funciona
- [ ] Executável inicia sem Go instalado
- [ ] Configuração padrão funciona

---

## Task 19.2 — Build do KMP Desktop

### Acceptance Criteria

- [ ] Aplicativo desktop é empacotado
- [ ] Abre fora da IDE
- [ ] Recursos necessários estão incluídos

---

## Task 19.3 — Estratégia de inicialização conjunta

Definir como usuário inicia:

```text
Go Agent
+
KMP Desktop
```

Possíveis opções futuras:

- launcher;
- processo filho;
- serviço separado.

### Acceptance Criteria

- [ ] Estratégia documentada
- [ ] MVP possui forma simples e reproduzível de iniciar ambos
- [ ] Frontend detecta agente ausente

---

# Epic 20 — Documentação e Release

## Task 20.1 — README

Incluir:

- objetivo;
- arquitetura;
- requisitos;
- como executar backend;
- como executar frontend;
- endpoints;
- screenshots posteriormente.

### Acceptance Criteria

- [ ] Novo desenvolvedor consegue executar o projeto
- [ ] Comandos estão atualizados

---

## Task 20.2 — Documentar API

Criar documentação para todos os endpoints.

### Acceptance Criteria

- [ ] Método HTTP
- [ ] Path
- [ ] Exemplo de response
- [ ] Erros esperados

---

## Task 20.3 — Criar release MVP

Versão sugerida:

```text
v0.1.0
```

### Acceptance Criteria

- [ ] Backend compila
- [ ] Frontend compila
- [ ] CPU funciona
- [ ] RAM funciona
- [ ] Disco funciona
- [ ] Rede funciona
- [ ] Processos funcionam
- [ ] WebSocket funciona
- [ ] Reconexão funciona
- [ ] README finalizado
- [ ] Tag/release criada no Git

---

# Ordem Recomendada de Implementação

```text
01 Bootstrap
    ↓
02 HTTP Server
    ↓
03 Models + interfaces
    ↓
04 CPU
    ↓
05 RAM
    ↓
06 Metrics agregado
    ↓
07 System
    ↓
08 Disk
    ↓
09 Network
    ↓
10 Processes
    ↓
11 Robustez backend
    ↓
12 KMP bootstrap
    ↓
13 UI mockada
    ↓
14 REST integration
    ↓
15 WebSocket
    ↓
16 Histórico
    ↓
17 UX
    ↓
18 Testes
    ↓
19 Build
    ↓
20 Release
```

---

# Milestones

## Milestone 1 — Backend mínimo

Completo quando:

- [ ] `/health`
- [ ] `/api/v1/cpu`
- [ ] `/api/v1/memory`
- [ ] `/api/v1/metrics`

Resultado: métricas reais acessíveis pelo navegador/Postman.

---

## Milestone 2 — Backend completo do MVP

Completo quando:

- [ ] System
- [ ] Disk
- [ ] Network
- [ ] Processes
- [ ] Error handling
- [ ] Logging

Resultado: agente Go já funciona sozinho como produto técnico.

---

## Milestone 3 — Desktop conectado

Completo quando:

- [ ] KMP abre
- [ ] REST funciona
- [ ] Dashboard usa métricas reais
- [ ] Processos reais aparecem

Resultado: primeira versão visual utilizável.

---

## Milestone 4 — Realtime

Completo quando:

- [ ] WebSocket funciona
- [ ] StateFlow atualiza UI
- [ ] Reconexão funciona
- [ ] Histórico curto funciona

Resultado: monitoramento em tempo real.

---

## Milestone 5 — MVP v0.1.0

Completo quando:

- [ ] Backend buildado
- [ ] Frontend empacotado
- [ ] Testes principais passam
- [ ] README completo
- [ ] Release criada

Resultado: projeto pronto para GitHub e portfólio.

---

# Definition of Done Global

Uma task só pode ser marcada como concluída quando:

- [ ] Código compila.
- [ ] Resultado pode ser verificado.
- [ ] Não introduz erro conhecido nas tasks anteriores.
- [ ] Tratamento básico de erro existe quando aplicável.
- [ ] Contratos públicos permanecem documentados.
- [ ] Testes são adicionados quando a task envolve lógica relevante.
