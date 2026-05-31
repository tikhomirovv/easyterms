# Technical Design

Стек и архитектура для EasyTerms. Обновляется по мере принятия решений.

## Стек

| Слой | Выбор | Комментарий |
|------|-------|-------------|
| Язык | **Go** | Основной язык backend и клиентов |
| Telegram-клиент | **[go-telegram/bot](https://github.com/go-telegram/bot)** | Активно поддерживается (2026), в [официальных Go samples](https://core.telegram.org/bots/samples#go), Bot API 10.0, middleware, zero deps |
| Ядро | Go packages (domain + application services) | Без привязки к Telegram |
| БД | **PostgreSQL** | Реляционные сущности + JSONB для гибких результатов анализа |
| LLM | **Port + adapters** | Domain-интерфейс в core; OpenAI-compatible — первая реализация |
| Payments | **Port + adapters** | MVP: stub/manual; v1: cost research + цены + ЮKassa |
| i18n | TBD (например `go-i18n` / JSON-каталоги) | Все UI-сообщения бота через переводы |

### Альтернативы Telegram (не выбраны)

- [gotgbot/v2](https://github.com/PaulSonOfLars/gotgbot) — codegen-обёртка, dispatcher; v2 пока RC
- [go-telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api) — устаревший вариант для нового проекта

## Архитектура

Monorepo: **ядро отдельно от клиентов**. Отдельный HTTP API на старте не нужен — клиенты вызывают core как Go-библиотеку. Архитектурно заложена возможность позже добавить web-клиент или REST/gRPC без переписывания логики.

```
┌─────────────────┐     ┌─────────────────┐
│  cmd/telegram   │     │  cmd/web (later) │
│  Telegram bot   │     │  Web UI / API    │
└────────┬────────┘     └────────┬─────────┘
         │                       │
         └───────────┬───────────┘
                     ▼
         ┌───────────────────────┐
         │   internal/core       │
         │   domain + services   │
         │   (documents, billing,│
         │    analysis, i18n)    │
         └───────────┬───────────┘
                     ▼
         ┌───────────────────────┐
         │   PostgreSQL          │
         └───────────────────────┘
```

Принципы:

- **Telegram-бот** — thin client: парсинг update, кнопки, locale, вызов core
- **Core** — use cases: создать документ, добавить источник, запустить анализ, списать проверку, история
- **Клиент не содержит бизнес-логики** — только адаптация transport → core
- Позже **web-клиент** подключается к тому же core (напрямую или через новый transport-слой)

## Ключевые решения

| Решение | Выбор | Обоснование |
|---------|-------|-------------|
| Язык | Go | Единый стек для core и клиентов |
| Форма продукта | Простой UX бота, document-centric backend (вариант C) | Быстрый UX для момента «перед Accept»; запас на рост |
| Monorepo + core/clients | Core package + `cmd/telegram` | Web позже без дублирования логики |
| Telegram library | go-telegram/bot | Актуальность API, community, официальный sample |
| Границы документа | Пользователь явно создаёт новый документ | Нет смешивания разных соглашений |
| Модель сессии | Один документ принимает несколько входов (текст, URL) | Многочастные загрузки |
| UX анализа | Кнопки после загрузки | Контроль пользователя; расширяемые режимы |
| Lazy analysis | Ingest через LLM → clean text; analysis modes по запросу + кэш | Ingest — первый LLM; analysis — без доп. списания |
| Списание проверки | При первом **успешном** LLM-вызове в документе | Обычно ingest; billing привязан к LLM, не к кнопке |
| LLM | Port `LLMClient` + adapters | Core не знает HTTP/вендоров; OpenAI-compatible — первая реализация |
| Payments | Port `PaymentProvider` + adapters | MVP stub/manual; v1: pricing + ЮKassa |
| i18n | UI бота переводится; анализ на языке locale Telegram | Один источник языка для UX и LLM |
| БД | PostgreSQL + JSONB | ACID для billing/баланса; гибкость для новых типов анализа |
| Единица монетизации | 1 проверка = 1 документ + все базовые кнопки | Понятная ценность |
| Free tier | Только demos и примеры | Стоимость inference |
| Учёт задач | GitHub через `gh` CLI | Отдельно от `.docs/` |

## База данных

### Выбор: PostgreSQL

Рекомендация — **PostgreSQL**, не MongoDB.

**Почему не отдельная document-DB:** опасение «Postgres разрастётся из-за новых типов ответов» решается **JSONB-колонками** для payload результатов анализа. Сущности с чёткой структурой (пользователь, документ, баланс, покупка) остаются в нормальных таблицах с FK и транзакциями.

**Почему не Mongo как основная БД:**

- billing, баланс проверок, списания — нужны **ACID и транзакции**
- связи User → Document → AnalysisResult → Purchase проще и надёжнее в SQL
- JSONB даёт гибкость схемы там, где она нужна (новые режимы анализа, версии промптов, metadata)

**Гибкая часть (JSONB):**

- `analysis_results.payload` — структура ответа LLM по типу анализа
- `analysis_results.meta` — модель, версия промпта, tokens, cost
- при новом типе анализа — новый `analysis_type` + свой JSON schema, без миграции десятка колонок

**Жёсткая часть (таблицы):**

- users, documents, document_sources, analysis_results (shell), purchases, check_ledger

Mongo имеет смысл только если позже появится отдельный тяжёлый document/log pipeline — для MVP это избыточно.

## LLM

### Принцип абстракции

Core **не вызывает HTTP и не знает про OpenAI/OpenRouter**. Он зависит только от domain-level port:

```go
// internal/core/ports/llm.go — пример, не финальный API

type LLMClient interface {
    ExtractCleanText(ctx context.Context, req ExtractRequest) (ExtractResponse, error)
    Analyze(ctx context.Context, req AnalyzeRequest) (AnalyzeResponse, error)
}
```

**ExtractRequest** — вход (URL, raw text, image bytes), locale пользователя, метаданные документа.  
**AnalyzeRequest** — clean text, режим (`plain`, `highlights`, …), locale, версия промпта.

Core не знает: `base URL`, `model`, headers, JSON schema API провайдера.

### Adapters (реализации)

Adapters живут **вне core** (`internal/llm/...`) и реализуют port:

| Adapter | Когда | Примечание |
|---------|-------|------------|
| `OpenAICompatibleProvider` | MVP / v1 | OpenAI, OpenRouter — один HTTP-клиент, разный конфиг |
| (future) другие провайдеры | по demand | Новый файл, core не меняется |

Конфиг adapter (env): `LLM_BASE_URL`, `LLM_API_KEY`, `LLM_MODEL`, `LLM_PROVIDER_LABEL` (логи, `AnalysisResult.meta`).

### Правила для разработки

- Core импортирует только `ports.LLMClient`, never concrete adapter
- Prompt templates — в core или отдельном пакете, не в adapter
- Adapter отвечает только за transport + mapping request/response
- Смена модели или провайдера — конфиг + (при необходимости) новый adapter, **без правок business logic**
- Тесты core — через mock `LLMClient`

### Где используется LLM

1. **Ingest** — `ExtractCleanText`: вход → clean text документа
2. **Analysis** — `Analyze`: clean text + режим → структурированный результат

### Billing и LLM

- Проверка списывается при **первом успешном** ответе LLM в рамках документа (типично — ingest)
- До успешного ingest пользователь не теряет проверку (ошибки, timeout — без списания)
- Последующие LLM-вызовы analysis modes в том же документе — без повторного списания
- Idempotency: повтор ingest/analysis того же типа — из кэша, без LLM

## Payments

### Принцип абстракции

Аналогично LLM: core работает с **port**, не с ЮKassa/Telegram напрямую.

```go
// internal/core/ports/payment.go — пример

type PaymentProvider interface {
    ID() string
    DisplayName() string
    SupportedPackages() []PackageOffer
    CreatePayment(ctx context.Context, req CreatePaymentRequest) (PaymentSession, error)
}

type PaymentSession struct {
    PaymentID string
    Kind      PaymentSessionKind // link | telegram_invoice | manual_pending
    URL       string             // для внешней оплаты
}
```

**BillingService** в core:

- `ListPackages()` — пакеты (1 / 3 / 10 проверок)
- `ListPaymentMethods(user)` — доступные providers для пользователя
- `StartPurchase(user, package, providerID)` → `PaymentSession`
- `ConfirmPayment(paymentID)` → начислить проверки (idempotent)
- `CreditChecksManual(user, n, reason)` — ручное пополнение (admin / stub)

Webhook handler (HTTP) вызывает `ConfirmPayment`; Telegram handler только рисует UI и вызывает core.

### UX оплаты (целевой)

По образцу удачных ботов (несколько способов, разные ссылки):

1. **«Купить проверки»**
2. Выбор пакета: `1 / 3 / 10` с ценой
3. Выбор **способа оплаты** — кнопки из зарегистрированных providers
4. Provider возвращает session:
   - **link** — URL + кнопка «Открыть оплату» + дублирование ссылки текстом (если in-app browser завис)
   - **telegram_invoice** — нативная оплата в Telegram (later)
   - **manual** — инструкция / ожидание ручного зачисления (MVP stub)
5. После подтверждения — начисление проверок, уведомление

### Roadmap providers

| Этап | Provider | Назначение |
|------|----------|------------|
| **MVP** | `ManualProvider` / `StubProvider` | Заглушка; admin вручную начисляет проверки |
| **v1** | Cost research + фиксация цен пакетов + `YooKassaProvider` | Замеры на готовой реализации → цены в ₽ → реальная оплата |

Дальнейшие providers (Telegram Payments, foreign card, crypto, Stars) — **не планируем заранее**; добавим при необходимости.

### ЮKassa (v1)

- Cost research на **готовой MVP-реализации** → фиксация цен пакетов (1 / 3 / 10) в PRD
- Payment link после выбора пакета
- Webhook на успешную оплату → `ConfirmPayment`
- Таблица `purchases`: external id, статус, сумма, пакет, provider
- Idempotency по external payment id
- Чеки / 54-ФЗ — при интеграции, отдельно от core

### MVP stub

- Provider `manual` или `dev_stub` в конфиге
- Admin: начислить N проверок по userId
- Таблицы balance / purchases — как в prod, чтобы не переписывать при подключении ЮKassa

### Правила для разработки

- Core импортирует только `ports.PaymentProvider`
- Registry включённых providers — конфиг
- Начисление проверок **только** через `BillingService`
- Telegram handler не содержит логику ЮKassa

## Основные сущности

- **User** — Telegram user id, locale, баланс проверок
- **Document** — одна проверка; статус (`draft` / `ingested` / `paid`), timestamps; флаг `check_consumed`
- **DocumentSource** — фрагмент входа: text paste, URL, (later) image
- **OriginalText** — assembled plain text документа (может быть полем или отдельной таблицей)
- **AnalysisResult** — тип анализа (`plain`, `highlights`, …), payload (JSONB), locale, cached flag
- **Purchase** — попытка оплаты: provider, package, сумма, статус, external payment id
- **CheckLedger** — движения баланса проверок (начисление, списание)

## Поток данных

### Purchase

1. Пользователь: купить → пакет → способ оплаты
2. `BillingService.StartPurchase` → `PaymentSession`
3. Telegram показывает link / invoice / manual instruction
4. Webhook или admin → `ConfirmPayment` → +N проверок на баланс

### Ingest

1. Пользователь добавляет источник(и) в документ
2. Core проверяет баланс проверок (до LLM)
3. LLM преобразует контент → **clean text** (или сбор простого текста без LLM, если применимо)
4. При **первом успешном** LLM-ответе — **списать 1 проверку**, сохранить original + clean text
5. Бот подтверждает готовность и показывает кнопки analysis

### Analysis

1. Пользователь выбирает режим
2. Core проверяет, что документ ingested и проверка уже списана
3. Если результат этого типа есть → вернуть из БД
4. Иначе → LLM → сохранить **AnalysisResult** → вернуть клиенту

## Структура проекта (целевая)

```
/
├── .docs/
├── .github/
│   └── workflows/         # CI: go test + docker build
├── cmd/
│   └── telegram/          # entrypoint Telegram-бота
├── internal/
│   ├── core/              # domain, services, ports (llm, payment)
│   ├── telegram/          # handlers, keyboards, i18n wiring
│   ├── llm/               # adapters: OpenAI-compatible, ...
│   ├── payment/           # adapters: manual, yookassa, ...
│   ├── storage/           # PostgreSQL repositories
│   └── ingest/            # URL fetch, preprocessing
├── locales/               # переводы UI бота
├── Dockerfile             # production-образ бота (multi-stage)
├── .dockerignore
├── go.mod
└── go.sum
```

`cmd/web/` — позже. Webhook endpoint для ЮKassa — `cmd/telegram` или отдельный `cmd/webhook` (TBD).

## Docker

Приложение **обязательно** упаковывается в Docker-образ для деплоя. Образ собирается на каждом PR в CI — локальная сборка опциональна.

### Образ

- **Multi-stage Dockerfile**: stage `builder` (Go compile) → stage `runtime` (minimal distroless или alpine + бинарник)
- Entrypoint: бинарник из `cmd/telegram`
- Конфиг — только через env (см. config loading в core/cmd)
- Образ не содержит секретов; только runtime-зависимости

### Правила

- `docker build` должен проходить в CI без дополнительных флагов (context — корень репозитория)
- Изменения, ломающие сборку образа, блокируют merge так же, как падающие тесты
- `.dockerignore` исключает `.git`, `.docs`, артефакты и лишний контекст — быстрая сборка
- `docker-compose.yml` — опционально позже (локальный dev с Postgres); для MVP достаточно Dockerfile + CI

## Среда разработки

Локальная машина разработчика **не обязана** иметь Go или Docker. Основной цикл:

1. Ветка → код + тесты → PR
2. **GitHub Actions** запускает `go test ./...` и `docker build`
3. Merge только при зелёных checks

Локальный `go test` / `go build` — по желанию, если инструменты установлены. Источник истины для «собирается / тесты проходят» — CI.

## Тестирование

Тесты **обязательны**. Архитектура (ports/adapters) проектируется так, чтобы core и сервисы можно было тестировать без Telegram, реального LLM и реальных платежей.

### Принципы

- **Каждая фича** на этапе, где это возможно, поставляется **вместе с тестами** — не «потом допишем»
- **Core / domain / services** — unit-тесты с mock ports (`LLMClient`, `PaymentProvider`, repositories)
- **Adapters** — integration-тесты там, где есть смысл (HTTP-клиент LLM — с mock server; storage — test DB или testcontainers)
- **Telegram handlers** — thin layer; ключевая логика тестируется в core, не через live Bot API
- LLM и payment **не вызываются в unit-тестах** — только mocks/fakes

### Что тестировать по слоям

| Слой | Тип | Примеры |
|------|-----|---------|
| `internal/core` | unit | ingest, consume check, analysis cache, billing ledger, idempotency |
| `internal/llm` | integration | mapping request/response, OpenAI-compatible client vs mock HTTP |
| `internal/payment` | unit + integration | stub provider; YooKassa webhook parsing (v1) |
| `internal/storage` | integration | migrations, repositories, transactions |
| `internal/telegram` | unit (optional) | mapping update → core call, keyboard builders |

### CI (GitHub Actions)

Workflow в `.github/workflows/ci.yml` — **обязательный gate** перед merge. Запускается на `push` и `pull_request` в `main` (GitHub Flow: feature branch → PR → merge).

| Job | Команда | Назначение |
|-----|---------|------------|
| **test** | `go test ./...` | unit/integration-тесты |
| **docker** | `docker build -t easyterms:ci .` | образ собирается на каждый PR |

Правила:

- Оба job должны быть зелёными для merge
- Новый PR без тестов на изменённую business logic — не принимается
- Агент/implementer после push **ждёт результат CI** и сообщает статус checks в PR
- Postgres service container в CI — добавляется с issue #4 (storage integration tests); до этого `go test` без live DB

### Инженерные правила (тесты)

- Ports (`LLMClient`, `PaymentProvider`, repositories) — всегда через интерфейсы, чтобы core был testable
- Тесты пишутся **в том же PR**, что и фича
- Для регрессий по billing и ledger — тесты обязательны (деньги и баланс)

## Инженерные правила

- Источник истины по продукту — `.docs/` (`project-overview.md`, `prd.md`, этот файл)
- **Docker-образ и CI checks** — обязательная часть каждого PR с кодом; merge без зелёного CI запрещён
- Локальные Go/Docker не требуются; проверка сборки и тестов — в GitHub Actions
- Core не импортирует Telegram SDK, vendor LLM SDK и SDK платёжных систем
- Core зависит только от ports (`LLMClient`, `PaymentProvider`); adapters — в `internal/llm`, `internal/payment`
- Все строки UI бота — через i18n, без хардкода в handlers
- LLM-промпты и шаблоны — версионировать; в `AnalysisResult.meta` писать версию
- Измерить себестоимость inference и зафиксировать цены пакетов — **в v1**, на готовой MVP-реализации
- Ответы бота содержат disclaimer «не юридическая консультация»
- JSONB для evolving analysis shapes; SQL migrations для стабильных таблиц
- Тесты на каждом этапе, где возможно; business logic без тестов не мёржится
