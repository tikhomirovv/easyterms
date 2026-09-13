[English](README.md)

# EasyTerms

Self-hosted Telegram-бот: помогает **понять пользовательские соглашения** (Terms, Privacy, EULA и т.п.) до нажатия «Принять». Создаёте документ, отправляете текст или ссылку — получаете простое объяснение и подсветку рисков.

**Не юридическая консультация** — только информационная помощь.

## Стек

- Go 1.23+, **SQLite** (встроенная БД, без отдельного сервера)
- [go-telegram/bot](https://github.com/go-telegram/bot)
- LLM через **OpenAI-compatible HTTP API** (OpenAI, OpenRouter, LM Studio, …)
- CI: `go test` при каждом merge; Docker-образ — на тегах релиза

Подробнее: [`.docs/`](.docs/).

## Быстрый старт

```bash
git clone https://github.com/tikhomirovv/easyterms.git
cd easyterms

cp .env.example .env
# заполните .env — минимум TELEGRAM_BOT_TOKEN, LLM_API_KEY, ALLOWED_TELEGRAM_IDS

go run ./cmd/telegram
```

Запускайте из **корня репозитория** — `.env` подхватится автоматически.

Миграции применяются при старте бота. Вручную:

```bash
go run ./cmd/migrate -direction up
```

## Конфигурация

| Переменная | Назначение |
|------------|------------|
| `LOG_LEVEL` | `debug` / `info` / `warn` / `error` (по умолчанию `info`) |
| `DATABASE_PATH` | Путь к файлу SQLite (по умолчанию `data/easyterms.db`) |
| `TELEGRAM_BOT_TOKEN` | Токен от [@BotFather](https://t.me/BotFather) |
| `TELEGRAM_PROXY` | Опционально: прокси для Bot API (`http://`, `socks5h://` и т.д.); пусто = напрямую |
| `ALLOWED_TELEGRAM_IDS` | Telegram ID через запятую; **пусто = публичный бот** (warning в логах) |
| `LLM_BASE_URL` | URL API |
| `LLM_API_KEY` | Ключ API |
| `LLM_MODEL` | Имя модели |

### Примеры LLM

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

**LM Studio (локально)**

```env
LLM_BASE_URL=http://127.0.0.1:1234/v1
LLM_API_KEY=lm-studio
LLM_MODEL=google/gemma-3-12b-it
```

### Рекомендуемые модели (только примеры)

Это **ориентиры**, не гарантия — цены и качество меняются; проверяйте у своего провайдера:

- **GPT-5.6 Luna** — OpenAI, хорош для summary
- **DeepSeek V4 Flash** — дёшево через OpenRouter / DeepSeek
- **Gemma 4** (12B / E4B) — open weights, локально или в облаке

## Сценарий в боте

`/start` → **Новый документ** → текст или URL → **Готово к разбору** → **Объяснить просто** / **Подсветить риски**

`/demo` — статичный пример.

## Тесты

```bash
go test ./...
```

## Docker

### Готовый образ (рекомендуется)

На каждый тег версии (`0.1.0`, `1.2.3`, … — без префикса `v`) CI публикует публичный **multi-arch** образ (`amd64` + `arm64`) в [GitHub Container Registry](https://github.com/tikhomirovv/easyterms/pkgs/container/easyterms). Docker сам выбирает архитектуру (ПК, VPS, Orange Pi, Raspberry Pi).

```bash
cp .env.example .env
# заполните TELEGRAM_BOT_TOKEN, LLM_API_KEY, ALLOWED_TELEGRAM_IDS

docker compose up -d
docker compose logs -f
```

См. [`docker-compose.yml`](docker-compose.yml) — минимальный пример с комментариями. Зафиксируйте версию вместо `latest`:

```yaml
image: ghcr.io/tikhomirovv/easyterms:0.1.0
```

Первый образ появится после push тега в `main` (например `git tag 0.1.0 && git push origin 0.1.0`).

### Сборка локально

```bash
docker build -t easyterms:latest .
docker run --rm --env-file .env -v easyterms-data:/app/data easyterms:latest
```

## Лицензия

MIT — см. [LICENSE](LICENSE).
