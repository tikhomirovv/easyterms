# EasyTerms

Telegram-бот, который помогает **понять пользовательские соглашения** (Terms, Privacy, EULA и т.п.) до нажатия «Принять». Пользователь создаёт документ, отправляет текст или ссылку, получает простое объяснение и подсветку рисков.

**Не является юридической консультацией** — только информационная помощь.

## Стек

- Go 1.23, PostgreSQL
- [go-telegram/bot](https://github.com/go-telegram/bot)
- LLM через **OpenAI-compatible HTTP API** (OpenAI, LM Studio, OpenRouter и др.)
- CI: `go test` + сборка Docker-образа на GitHub Actions

Подробнее о продукте и архитектуре: [`.docs/`](.docs/).

## Требования

- **PostgreSQL** (локально, в облаке или в CI) — строка подключения в `DATABASE_URL`
- **Telegram Bot Token** — от [@BotFather](https://t.me/BotFather)
- **LLM API** — ключ OpenAI или локальный сервер (LM Studio)
- **Go 1.23+** — для локального запуска (опционально; тесты можно гонять в CI)

## Быстрый старт

```bash
git clone https://github.com/tikhomirovv/easyterms.git
cd easyterms

cp .env.example .env
# отредактируйте .env (см. ниже)

# миграции БД
go run ./cmd/migrate -direction up

# бот
go run ./cmd/telegram
```

Команды запускайте **из корня репозитория** — файл `.env` подхватывается автоматически.

## Конфигурация (`.env`)

Скопируйте [`.env.example`](.env.example) в `.env` и заполните:

| Переменная | Назначение |
|------------|------------|
| `DATABASE_URL` | PostgreSQL, напр. `postgres://user:pass@localhost:5432/easyterms?sslmode=disable` |
| `TELEGRAM_BOT_TOKEN` | Токен бота |
| `LOG_LEVEL` | `debug` / `info` / `warn` / `error` (по умолчанию `info`) |
| `LLM_PROVIDER` | `openai-compatible` (по умолчанию) |
| `LLM_BASE_URL` | URL API, напр. `https://api.openai.com/v1` |
| `LLM_API_KEY` | Ключ API (для OpenAI обязателен) |
| `LLM_MODEL` | Имя модели |
| `LLM_JSON_MODE` | `true` для OpenAI; часто `false` для LM Studio |
| `LLM_PROVIDER_LABEL` | Метка в логах |

### OpenAI (облако)

```env
LLM_BASE_URL=https://api.openai.com/v1
LLM_API_KEY=sk-...
LLM_MODEL=gpt-4o-mini
LLM_JSON_MODE=true
```

### LM Studio (локально)

В LM Studio включите сервер (OpenAI-compatible), загрузите модель.

```env
LLM_BASE_URL=http://127.0.0.1:1234/v1
LLM_API_KEY=lm-studio
LLM_MODEL=qwen/qwen3.6-27b
LLM_JSON_MODE=false
LLM_PROVIDER_LABEL=lm-studio
```

`LLM_MODEL` должен совпадать с именем модели в LM Studio.

## Команды

### Миграции БД

```bash
go run ./cmd/migrate -direction up    # применить
go run ./cmd/migrate -direction down  # откатить
```

Нужен `DATABASE_URL` в окружении или в `.env`.

### Telegram-бот

```bash
go run ./cmd/telegram
```

Обязательны: `TELEGRAM_BOT_TOKEN`, `DATABASE_URL`, настройки LLM.

**В боте:** `/start` → «Новый документ» → текст или URL → «Готово к разбору» → «Объяснить просто» / «Подсветить риски».  
`/demo` — пример без списания проверки.

### Начисление проверок (admin, MVP)

Ручное пополнение баланса (заглушка оплаты):

```bash
go run ./cmd/credit -telegram-id YOUR_TELEGRAM_ID -amount 3 -key admin-001
```

или `-user-id <uuid>` вместо `-telegram-id`.

## Тесты

```bash
go test ./...
```

Интеграционные тесты PostgreSQL (опционально, локально):

```bash
export DATABASE_URL=postgres://...
go run ./cmd/migrate -direction up
go test -tags=integration ./internal/storage/postgres/...
```

## Docker

Сборка и запуск образа бота:

```bash
docker build -t easyterms:latest .
docker run --rm --env-file .env easyterms:latest
```

Перед первым запуском примените миграции к вашей БД (`cmd/migrate` или отдельный job).

## Структура репозитория

```
cmd/
  telegram/   # бот (основной entrypoint)
  migrate/    # миграции SQL
  credit/     # admin: начисление проверок
internal/
  core/       # домен, сервисы, порты
  telegram/   # handlers, i18n, клавиатуры
  llm/        # OpenAI-compatible адаптер
  storage/    # PostgreSQL, миграции
.docs/        # PRD, технический дизайн
```

## Лицензия

Уточняется.
