# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Context

This is a **fork of [it-beard/evo-bot-go](https://github.com/it-beard/evo-bot-go)** — a Telegram bot for managing an AI & Programming learning community. All user-facing strings have been translated from Russian to English and the bot is rebranded as "AI & Programming Course Bot" with Mini App course integration.

**Mini App course URL**: `https://antikriza.github.io/BBD-evolution-code-clone/telegram-archive/course/twa/index.html`

## Common Commands

### Development
- **Run the bot**: `go run main.go`
- **Build**: `go build -o bot`
- **Update deps**: `go mod tidy`
- **Cross-compile**: `GOOS=linux GOARCH=amd64 go build -o bot`

### Testing
- **All tests**: `go test ./...`
- **Verbose**: `go test -v ./...`
- **Specific package**: `go test evo-bot-go/internal/handlers/privatehandlers`
- **Specific test**: `go test -run TestHelpHandler_Name evo-bot-go/internal/handlers/privatehandlers`
- **Coverage**: `go test -cover ./...`
- **Race detection**: `go test -race ./...`
- **Colored output**: `go run gotest.tools/gotestsum@latest --format pkgname --format-icons hivis`

## Architecture

Go 1.23+ Telegram bot with layered architecture and dependency injection.

### Entry Point
`main.go` → LoadConfig → NewOpenAiClient → NewTgBotClient → Start (with graceful shutdown)

### Layer Structure

| Layer | Path | Responsibility |
|-------|------|----------------|
| Bot | `internal/bot/` | Handler registration, DI, scheduled tasks |
| Config | `internal/config/` | Env var loading (`TG_EVO_BOT_` prefix) |
| Handlers | `internal/handlers/` | Telegram update processing |
| Services | `internal/services/` | Business logic |
| Repositories | `internal/database/repositories/` | PostgreSQL data access |
| Database | `internal/database/` | Connection + auto-migrations |
| Prompts | `internal/database/prompts/` | AI prompt templates (English) |
| Formatters | `internal/formatters/` | Message formatting |
| Buttons | `internal/buttons/` | Inline keyboard layouts |
| Tasks | `internal/tasks/` | Scheduled background jobs |
| Constants | `internal/constants/` | Command names, callback keys |
| Utils | `internal/utils/` | Permissions, chat ID helpers |

### Handlers by context
- `adminhandlers/` — event CRUD, profile management, test triggers
- `grouphandlers/` — thread moderation, message tracking, join/leave cleanup
- `privatehandlers/` — AI search (`/tools`, `/content`, `/intro`), profiles, topics

### Key services
- `SummarizationService` — daily AI chat summaries
- `RandomCoffeeService` — weekly poll + smart pairing with history
- `ProfileService` — user profile management and publishing
- `PermissionsService` — admin/member authorization
- `MessageSenderService` — centralized message sending with HTML/Markdown support

### Important patterns
- All handlers receive `HandlerDependencies` struct (services + repositories)
- Config requires env vars: bot token, chat ID, OpenAI key, DB connection, topic IDs
- DB schema auto-initializes via migrations on startup (11 migrations)
- AI prompts stored as Go constants in `internal/database/prompts/` with `fmt.Sprintf` format verbs