# AI & Programming Course Bot

![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)

A Telegram bot for managing an AI & Programming learning community. Moderates group discussions, provides AI-powered search across curated content, generates daily chat summaries, and integrates with a [Mini App course](https://antikriza.github.io/BBD-evolution-code-clone/telegram-archive/course/twa/index.html) covering 42 topics across 5 levels.

> **Fork of [it-beard/evo-bot-go](https://github.com/it-beard/evo-bot-go)** — all user-facing strings translated from Russian to English, rebranded, and integrated with the AI course Mini App.

## Quick Start

```bash
# 1. Clone
git clone https://github.com/antikriza/evo-bot-go.git && cd evo-bot-go

# 2. Configure
cp .env.example .env   # then fill in your values

# 3. Run
go run main.go
```

Prerequisites: Go 1.23+, PostgreSQL (tables auto-created on first run).

## Features

### Moderation
- **Thread Management** — deletes non-admin messages in read-only threads
- **Message Forwarding** — forwards replies from closed threads to the general topic
- **Join/Leave Cleanup** — removes join/leave messages for cleaner conversations
- **Message Tracking** — stores group messages for AI summarization
- **Topic Management** — tracks forum topics and metadata

### AI-Powered Search
- `/tools` — find AI tools from the Tools channel (fast / deep modes)
- `/content` — find content from the Video Content channel
- `/intro` — find member info from the Intro channel (smart profile search)
- **Daily Summarization** — AI-generated chat summaries posted on schedule
  - Manual trigger: `/trySummarize` (admin-only)
  - Send course link: `/tryLinkToLearn` (admin-only)

### Random Coffee
- Weekly automated polls (configurable day/time, default: Friday 14:00 UTC)
- Smart pairing algorithm considering pairing history (default: Monday 12:00 UTC)
- Manual pairing: `/tryGenerateCoffeePairs` (admin-only)

### Profiles & Events
- `/profile` — create, edit, and publish your profile to the Intro topic
- `/events` — view upcoming events
- `/topics` / `/topicAdd` — browse or suggest event topics
- `/profilesManager` — admin tool for managing member profiles
- `/eventSetup`, `/eventEdit`, `/eventStart`, `/eventDelete` — admin event management

### Course Integration
- `/start` shows an "Open AI Course" button linking to the [Mini App](https://antikriza.github.io/BBD-evolution-code-clone/telegram-archive/course/twa/index.html)
- `/help` includes a course link at the bottom
- `/tryLinkToLearn` sends the course link in a DM

Use `/help` in the bot chat for the full command list.

## Technology Stack

- **Language**: Go 1.23+
- **Framework**: [gotgbot](https://github.com/PaulSonOfLars/gotgbot) for Telegram Bot API
- **Database**: PostgreSQL with automated migrations
- **AI Integration**: OpenAI API for content analysis and search
- **Architecture**: Clean layered architecture with dependency injection
- **Testing**: Unit tests with gotestsum support

## Architecture

```
internal/
├── bot/           # Bot setup, handler registration, dependency injection
├── buttons/       # Inline keyboard button layouts
├── clients/       # OpenAI API client
├── config/        # Environment variable loading
├── constants/     # Command names, callback keys
├── database/
│   ├── migrations/    # PostgreSQL schema migrations (auto-run)
│   ├── prompts/       # AI prompt templates
│   └── repositories/  # Data access layer
├── formatters/    # Message formatting (help, profiles, events)
├── handlers/
│   ├── adminhandlers/     # Admin commands (events, profiles, test triggers)
│   ├── grouphandlers/     # Group moderation (threads, join/leave cleanup)
│   └── privatehandlers/   # User commands (AI search, profile, topics)
├── services/      # Business logic (coffee, summarization, permissions)
├── tasks/         # Scheduled jobs (daily summary, weekly coffee)
└── utils/         # Helpers (permissions, chat ID conversion)
```

## Required Bot Permissions

The bot needs these admin permissions in the Telegram supergroup:

- **Pin messages** — for event announcements
- **Delete messages** — for thread moderation and join/leave cleanup

## Database

PostgreSQL with auto-initialized tables (11 migrations run on first startup):

| Table | Purpose |
|-------|---------|
| `group_messages` | Group messages stored for AI summarization |
| `group_topics` | Forum topic names and metadata |
| `prompting_templates` | Customizable AI prompt templates |
| `users` | User info, karma score, coffee ban status |
| `profiles` | User bios and published intro message IDs |
| `events` | Community events (type, status, start time) |
| `topics` | Event discussion topics and questions |
| `random_coffee_polls` | Weekly coffee poll tracking |
| `random_coffee_participants` | Poll participation responses |
| `random_coffee_pairs` | Pairing history for smart matching |
| `migrations` | Schema migration tracking |

## Building

```bash
# Development
go run main.go

# Build for current platform
go build -o bot

# Cross-compile
GOOS=linux  GOARCH=amd64 go build -o bot        # Linux
GOOS=darwin GOARCH=arm64 go build -o bot        # macOS (Apple Silicon)
GOOS=windows GOARCH=amd64 go build -o bot.exe   # Windows

# Update dependencies
go mod tidy
```

## Configuration

Copy `.env.example` to `.env` and fill in your values. All variables use the `TG_EVO_BOT_` prefix.

### Required

| Variable | Description |
|----------|-------------|
| `TG_EVO_BOT_TOKEN` | Bot token from [@BotFather](https://t.me/BotFather) |
| `TG_EVO_BOT_SUPERGROUP_CHAT_ID` | Supergroup chat ID (negative number) |
| `TG_EVO_BOT_OPENAI_API_KEY` | OpenAI API key |
| `TG_EVO_BOT_DB_CONNECTION` | PostgreSQL connection string |
| `TG_EVO_BOT_ADMIN_USER_ID` | Your Telegram user ID |

### Topic IDs

Create these forum topics in your supergroup, then set their thread IDs:

| Variable | Topic |
|----------|-------|
| `TG_EVO_BOT_TOOL_TOPIC_ID` | Tools — AI tools database for `/tools` |
| `TG_EVO_BOT_CONTENT_TOPIC_ID` | Content — video/article content for `/content` |
| `TG_EVO_BOT_INTRO_TOPIC_ID` | Introductions — member profiles for `/intro` |
| `TG_EVO_BOT_ANNOUNCEMENT_TOPIC_ID` | Announcements |
| `TG_EVO_BOT_SUMMARY_TOPIC_ID` | Daily Summary — where AI summaries are posted |
| `TG_EVO_BOT_RANDOM_COFFEE_TOPIC_ID` | Random Coffee — polls and pair announcements |
| `TG_EVO_BOT_MONITORED_TOPICS_IDS` | Comma-separated IDs to include in daily summaries |

### Optional

| Variable | Default | Description |
|----------|---------|-------------|
| `TG_EVO_BOT_CLOSED_TOPICS_IDS` | — | Comma-separated read-only topic IDs |
| `TG_EVO_BOT_FORWARDING_TOPIC_ID` | `0` | Topic for forwarded replies (0 = General) |
| `TG_EVO_BOT_SUMMARY_TIME` | `03:00` | Daily summary time (24h UTC) |
| `TG_EVO_BOT_SUMMARIZATION_TASK_ENABLED` | `true` | Enable daily summaries |
| `TG_EVO_BOT_RANDOM_COFFEE_POLL_TASK_ENABLED` | `false` | Enable weekly coffee polls |
| `TG_EVO_BOT_RANDOM_COFFEE_POLL_TIME` | `14:00` | Poll creation time (24h UTC) |
| `TG_EVO_BOT_RANDOM_COFFEE_POLL_DAY` | `friday` | Day to create poll |
| `TG_EVO_BOT_RANDOM_COFFEE_PAIRS_TASK_ENABLED` | `false` | Enable auto pair generation |
| `TG_EVO_BOT_RANDOM_COFFEE_PAIRS_TIME` | `12:00` | Pair announcement time (24h UTC) |
| `TG_EVO_BOT_RANDOM_COFFEE_PAIRS_DAY` | `monday` | Day to announce pairs |

## Testing

```bash
go test ./...                    # Run all tests
go test -v ./...                 # Verbose output
go test -cover ./...             # Coverage summary
go test -race ./...              # Race condition detection
go test -run TestName ./...      # Run specific test

# Coverage HTML report
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

# Colored output with gotestsum
go run gotest.tools/gotestsum@latest --format pkgname --format-icons hivis
```

## Upstream

Forked from [it-beard/evo-bot-go](https://github.com/it-beard/evo-bot-go). Key changes in this fork:

- All user-facing strings translated from Russian to English
- Removed "Evocoders" / "Evolution of Code" branding
- Integrated [AI & Programming Course Mini App](https://antikriza.github.io/BBD-evolution-code-clone/telegram-archive/course/twa/index.html) (42 topics, 5 levels)
- Added `.env.example` for easier setup
- Renamed `GetTypeInRussian()` to `GetTypeName()` in event formatters
