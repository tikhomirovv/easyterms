# Technical Design

Стек и архитектура EasyTerms (OSS personal bot). Обновляется по мере принятия решений.

## Стек

| Слой | Выбор | Комментарий |
|------|-------|-------------|
| Язык | **Go** | Backend и Telegram entrypoint |
| Telegram | **[go-telegram/bot](https://github.com/go-telegram/bot)** | Thin client поверх core |
| Ядро | `internal/core` | Domain + application services, без transport |
| БД | **SQLite** (`modernc.org/sqlite`, pure Go) | Файл по `DATABASE_PATH`, default `data/easyterms.db` |
| LLM | Port + OpenAI-compatible adapter | `LLM_BASE_URL`, `LLM_API_KEY`, `LLM_MODEL` |
| i18n | JSON-каталоги в `internal/telegram/i18n` | UI по locale Telegram |
| Лицензия | MIT | Open source, self-hosted |

**Убрано из scope:** PostgreSQL, billing, payment providers, check balance / ledger.

## Архитектура

```
┌─────────────────┐
│  cmd/telegram   │  allowlist, handlers, migrations on startup
└────────┬────────┘
         ▼
┌───────────────────────┐
│   internal/core       │  documents, analysis, ingest orchestration
└─────────┬─────────────┘
          ▼
┌───────────────────────┐     ┌──────────────────┐
│ internal/storage/     │     │ internal/llm/    │
│ sqlite + migrate      │     │ openai-compatible│
└───────────────────────┘     └──────────────────┘
```

Принципы:

- **Telegram** — transport only: updates → core, i18n, keyboards
- **Core** — use cases без знания SQLite/Telegram/HTTP
- **Migrations** — SQL в `internal/storage/migrations/`, runner в `internal/storage/migrate/`

## Ключевые решения

| Решение | Выбор |
|---------|-------|
| БД | SQLite, один файл, без отдельного сервера |
| Миграции | Custom runner + `schema_migrations`; bot auto-migrates on start |
| Доступ | `ALLOWED_TELEGRAM_IDS`; пусто = public + WARN |
| LLM JSON | Prompt-only JSON; invalid response → **1 retry** |
| LLM config | No `LLM_JSON_MODE` / `LLM_PROVIDER_LABEL`; host from `LLM_BASE_URL` in logs |
| Монетизация | Нет — владелец платит LLM API сам |
| Обратная совместимость | Не требуется (pre-production) |

## База данных

### SQLite schema (упрощённо)

- `users` — telegram id, locale
- `documents` — status (`draft` / `ingested`), timestamps
- `document_sources` — text paste, URL
- `analysis_results` — type (`plain`, `highlights`), JSON payload, cache

Нет таблиц purchases, check_ledger, balance.

### Migrations

- `go run ./cmd/migrate -direction up|down`
- `cmd/telegram` вызывает migrate up при старте

## LLM

```go
type LLMClient interface {
    ExtractCleanText(ctx context.Context, req ExtractRequest) (ExtractResponse, error)
    Analyze(ctx context.Context, req AnalyzeRequest) (AnalyzeResponse, error)
}
```

Adapter: `internal/llm/openai` — chat completions, no `response_format`; retry on invalid JSON for `Analyze`.

Env:

- `LLM_BASE_URL`
- `LLM_API_KEY`
- `LLM_MODEL`

## Telegram access control

`internal/telegram/allowlist.go` — parse `ALLOWED_TELEGRAM_IDS`. Denied users get localized message; no core calls.

## Поток данных

### Ingest

1. User adds source(s) to document
2. URL fetch (`internal/ingest/urlfetch`) or direct text
3. LLM → clean text → save document (`ingested`)
4. Bot shows analysis buttons

### Analysis

1. User picks mode
2. Cache hit → return from DB
3. Miss → LLM → save `analysis_results` → reply

## Структура проекта

```
/
├── cmd/
│   ├── telegram/       # bot entrypoint
│   └── migrate/        # CLI migrations
├── internal/
│   ├── core/           # domain, services, ports, config, prompts
│   ├── telegram/       # bot, allowlist, i18n
│   ├── llm/            # factory + openai adapter
│   ├── storage/
│   │   ├── sqlite/     # repositories
│   │   ├── migrate/    # migration runner
│   │   └── migrations/ # SQL files
│   └── ingest/         # URL fetch
├── .docs/
├── Dockerfile
└── LICENSE
```

## Docker

Multi-stage: build `cmd/telegram` → minimal alpine runtime.

**Release:** `.github/workflows/release.yml` — on git tag `v*`, push to `ghcr.io/tikhomirovv/easyterms` (`latest` + semver tags).

**Local / compose:** root `docker-compose.yml` — `env_file: .env`, volume `easyterms-data:/app/data`.

```bash
docker compose up -d
```

## Тестирование

| Слой | Тип |
|------|-----|
| `internal/core` | unit (mocks) |
| `internal/llm/openai` | unit + httptest |
| `internal/storage/sqlite` | integration (temp DB) |
| `internal/storage/migrate` | integration |
| `internal/telegram` | unit (allowlist, helpers) |

CI (`.github/workflows/ci.yml`): `go test ./...` + `docker build`. No Postgres service.

## Инженерные правила

- Core не импортирует Telegram SDK и concrete LLM HTTP client
- UI strings — только через i18n
- Disclaimer «не юридическая консультация» в ответах бота
- Тесты в том же PR, что и фича
