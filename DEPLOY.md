# Deploy to VPS

## Requirements

| Resource | Minimum |
|----------|---------|
| OS | Ubuntu 22.04+ / Debian 12+ |
| RAM | 512 MB |
| Disk | 5 GB |
| CPU | 1 vCPU |
| Network | Outbound HTTPS (no inbound ports needed) |

The bot uses **long polling** (not webhooks), so no domain, SSL, or open ports are required.

## Files to transfer

```
evo-bot-go/
├── Dockerfile              # Multi-stage Go build → Alpine runtime
├── docker-compose.yml      # Bot + PostgreSQL services
├── .env                    # Your secrets (NEVER commit this)
├── .dockerignore           # Excludes .env, .git, docs from build
├── go.mod                  # Go module definition
├── go.sum                  # Dependency checksums
├── main.go                 # Entry point
└── internal/               # All application code
    ├── bot/
    ├── buttons/
    ├── clients/
    ├── config/
    ├── constants/
    ├── database/
    ├── formatters/
    ├── handlers/
    ├── services/
    ├── tasks/
    └── utils/
```

**Do NOT transfer**: `.env` (create it on the server), `.git/`, `.qoder/`, `*.md` docs.

## Step-by-step deployment

### 1. Install Docker on VPS

```bash
# SSH into your VPS
ssh user@your-vps-ip

# Install Docker (Ubuntu/Debian)
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER
# Log out and back in for group change to take effect
exit
ssh user@your-vps-ip

# Verify
docker --version
docker compose version
```

### 2. Clone the repository

```bash
git clone https://github.com/antikriza/evo-bot-go.git
cd evo-bot-go
```

### 3. Create .env file

```bash
cp .env.example .env
nano .env
```

Fill in all required values:

```env
TG_EVO_BOT_TOKEN=your_bot_token
TG_EVO_BOT_SUPERGROUP_CHAT_ID=your_chat_id
TG_EVO_BOT_OPENAI_API_KEY=sk-your-key
TG_EVO_BOT_ADMIN_USER_ID=your_telegram_user_id
TG_EVO_BOT_TOOL_TOPIC_ID=18
TG_EVO_BOT_CONTENT_TOPIC_ID=19
TG_EVO_BOT_INTRO_TOPIC_ID=20
TG_EVO_BOT_ANNOUNCEMENT_TOPIC_ID=21
TG_EVO_BOT_SUMMARY_TOPIC_ID=1
TG_EVO_BOT_RANDOM_COFFEE_TOPIC_ID=1
TG_EVO_BOT_MONITORED_TOPICS_IDS=1,18,19
TG_EVO_BOT_SUMMARIZATION_TASK_ENABLED=false
TG_EVO_BOT_RANDOM_COFFEE_POLL_TASK_ENABLED=false
TG_EVO_BOT_RANDOM_COFFEE_PAIRS_TASK_ENABLED=false
```

Secure the file:
```bash
chmod 600 .env
```

### 4. Start the bot

```bash
docker compose up -d
```

First run takes ~2 minutes (downloads Go image, compiles, downloads PostgreSQL).
Subsequent starts are instant.

### 5. Verify

```bash
# Check containers are running
docker compose ps

# Check bot logs
docker compose logs bot --tail 20

# Expected output:
# Successfully applied migration: ...  (x11)
# DailySummarizationTask: Daily summarization task is disabled
# Bot Runner: Bot @pm_ai_club_bot has been started successfully
```

### 6. Test

Send `/start` to your bot in Telegram DM. You should see the welcome message.

## Operations

### View logs

```bash
# Last 50 lines
docker compose logs bot --tail 50

# Follow live
docker compose logs bot -f

# All logs since last restart
docker compose logs bot
```

### Restart

```bash
docker compose restart bot
```

### Update (after pushing code changes)

```bash
cd ~/evo-bot-go
git pull
docker compose build bot
docker compose up -d bot
```

### Stop

```bash
docker compose down        # Stop containers (keeps data)
docker compose down -v     # Stop and DELETE database (fresh start)
```

### Database backup

```bash
# Backup
docker compose exec db pg_dump -U evobot evobot > backup_$(date +%Y%m%d).sql

# Restore
docker compose exec -T db psql -U evobot evobot < backup_20260207.sql
```

### Change .env variables

```bash
nano .env
docker compose up -d bot   # Recreates bot container with new env
```

## Auto-restart on reboot

Docker containers with `restart: unless-stopped` auto-start when Docker starts.
Ensure Docker starts on boot:

```bash
sudo systemctl enable docker
```

## Monitoring (optional)

### Simple health check script

Create `~/check-bot.sh`:
```bash
#!/bin/bash
if ! docker compose -f ~/evo-bot-go/docker-compose.yml ps bot | grep -q "Up"; then
    echo "$(date): Bot is down, restarting..." >> ~/bot-monitor.log
    cd ~/evo-bot-go && docker compose up -d bot
fi
```

Add to crontab (checks every 5 minutes):
```bash
chmod +x ~/check-bot.sh
(crontab -l 2>/dev/null; echo "*/5 * * * * ~/check-bot.sh") | crontab -
```

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| `Bot has been started successfully` then crash loop | Check `.env` values, especially topic IDs |
| `failed to ping database` | PostgreSQL not ready — wait 10s or `docker compose restart` |
| `TG_EVO_BOT_TOKEN environment variable is not set` | `.env` file missing or not in the same directory |
| `conflict: port is already allocated` | Another service uses port 5432 — the bot uses internal Docker networking, no host ports needed |
| Bot doesn't respond to commands | Verify token with `curl https://api.telegram.org/bot<TOKEN>/getMe` |
| Bot responds in DM but not in group | Expected — commands work in private DM only |

## Resource usage

| Container | RAM | Disk |
|-----------|-----|------|
| bot | ~15 MB | ~20 MB (binary) |
| PostgreSQL | ~30 MB | ~50 MB (empty DB) |
| **Total** | **~50 MB** | **~100 MB** |

Cost: Any $4-5/month VPS (Hetzner, DigitalOcean, Vultr) is more than enough.
