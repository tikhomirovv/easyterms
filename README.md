[Русский](README.ru.md)

# EasyTerms

Self-hosted Telegram bot that helps you **understand terms of service** (Terms, Privacy, EULA, etc.) before you click Accept. Create a document, send text or a URL, get a plain-language summary and risk highlights.

**Not legal advice** — informational help only.

## Stack

- Go 1.23+, **SQLite** (embedded, no separate DB server)
- [go-telegram/bot](https://github.com/go-telegram/bot)
- LLM via **OpenAI-compatible HTTP API** (OpenAI, OpenRouter, LM Studio, …)
- CI: `go test` + Docker build on GitHub Actions

More context: [`.docs/`](.docs/).

## Quick start

```bash
git clone https://github.com/tikhomirovv/easyterms.git
cd easyterms

cp .env.example .env
# edit .env — at minimum TELEGRAM_BOT_TOKEN, LLM_API_KEY, ALLOWED_TELEGRAM_IDS

go run ./cmd/telegram
```

Run from the **repo root** so `.env` is found.

Migrations run automatically on bot startup. To apply manually:

```bash
go run ./cmd/migrate -direction up
```

## Configuration

| Variable | Description |
|----------|-------------|
| `LOG_LEVEL` | `debug` / `info` / `warn` / `error` (default `info`) |
| `DATABASE_PATH` | SQLite file path (default `data/easyterms.db`) |
| `TELEGRAM_BOT_TOKEN` | Bot token from [@BotFather](https://t.me/BotFather) |
| `ALLOWED_TELEGRAM_IDS` | Comma-separated Telegram user IDs; **empty = public bot** (warns at startup) |
| `LLM_BASE_URL` | API base URL, e.g. `https://api.openai.com/v1` or OpenRouter / LM Studio |
| `LLM_API_KEY` | API key (placeholder OK for local servers) |
| `LLM_MODEL` | Model name for your provider |

### Example LLM setups

**OpenAI**

```env
LLM_BASE_URL=https://api.openai.com/v1
LLM_API_KEY=sk-...
LLM_MODEL=gpt-5.6-luna
```

**OpenRouter**

```env
LLM_BASE_URL=https://openrouter.ai/api/v1
LLM_API_KEY=...
LLM_MODEL=deepseek/deepseek-v4-flash-0731
```

**LM Studio (local)**

```env
LLM_BASE_URL=http://127.0.0.1:1234/v1
LLM_API_KEY=lm-studio
LLM_MODEL=google/gemma-3-12b-it
```

### Recommended models (examples only)

These are **starting points**, not guarantees — pricing and quality change; test on your provider:

- **GPT-5.6 Luna** (`gpt-5.6-luna`) — OpenAI, summarization-oriented
- **DeepSeek V4 Flash** — low-cost API via OpenRouter or DeepSeek
- **Gemma 4** (e.g. 12B / E4B) — open weights, local or hosted

## Bot flow

`/start` → **New document** → paste text or URL → **Ready to analyze** → **Explain simply** / **Highlight risks**

`/demo` — static example text.

## Tests

```bash
go test ./...
```

## Docker

### Pre-built image (recommended)

On every version tag (`v*`), CI publishes a public image to [GitHub Container Registry](https://github.com/tikhomirovv/easyterms/pkgs/container/easyterms).

```bash
cp .env.example .env
# edit .env — TELEGRAM_BOT_TOKEN, LLM_API_KEY, ALLOWED_TELEGRAM_IDS

docker compose up -d
docker compose logs -f
```

See [`docker-compose.yml`](docker-compose.yml) — minimal example with comments. Pin a release instead of `latest`:

```yaml
image: ghcr.io/tikhomirovv/easyterms:0.1.0
```

The first image appears after you push a tag to `main` (e.g. `git tag v0.1.0 && git push origin v0.1.0`).

### Build locally

```bash
docker build -t easyterms:latest .
docker run --rm --env-file .env -v easyterms-data:/app/data easyterms:latest
```

Mount a volume for `data/` if you use the default `DATABASE_PATH`.

## License

MIT — see [LICENSE](LICENSE).
