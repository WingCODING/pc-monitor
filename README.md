# PC Monitor

Aplicativo desktop para monitoramento em tempo real dos recursos do computador.

- **Backend / Agent:** Go (`agent/`)
- **Frontend Desktop:** Kotlin Multiplatform + Compose Multiplatform (`desktop/`)

Veja [`plan.md`](plan.md) e [`spec-driven-tasks.md`](spec-driven-tasks.md) para a especificação completa e o roadmap.

## Desenvolvimento

### Backend (Go)

```bash
cd agent
go run ./cmd/server
```

### Frontend (Kotlin Multiplatform / Compose Desktop)

```bash
cd desktop
./gradlew :composeApp:run
```
