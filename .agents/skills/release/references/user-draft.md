# User-facing release section

Section **What's new for users** in the GitHub Release notes.

## Audience

A **self-hoster** running the Telegram bot, or a **Telegram user** talking to the bot. Not a Go developer reading the repo.

**Test each bullet:** would this make sense without knowing SQLite, GHCR, or internal package names?

## When to include the section

Include **only if** the commit range has **honest, specific** user value, for example:

- New bot commands, buttons, or flows.
- Clearer error messages or i18n for Telegram UI.
- Allowlist / config that changes who can use the bot (with plain explanation).
- Docker / compose changes that simplify setup for self-hosters.
- Fixes users could hit (crashes, broken URL ingest, analysis failures).

**Omit the entire section** when the release is internal-only: refactor, CI, tests, agent skills, docs-only, dependency bumps with no behavior change.

Do not add filler («improved stability») unless the user asks.

## Language

- **English**, plain sentences.
- Outcomes: «The bot now…», «Fixed an error when…», «Docker Compose example added for faster setup…»

## Never in user bullets

- Commit subjects, file paths, Go package names.
- Issue/PR numbers, GitHub Actions, GHCR URLs (unless explaining a one-line pull command).
- «Refactor», internal skill names, migration file names.

## Good vs bad

| Bad | Good |
|-----|------|
| Migrate storage from PostgreSQL to SQLite | Self-hosting no longer needs a separate database server — single SQLite file |
| Bump actions/checkout to v6 | (omit — CI only) |
| Add Telegram allowlist via ALLOWED_TELEGRAM_IDS | Restrict bot access to specific Telegram user IDs via `ALLOWED_TELEGRAM_IDS` |
