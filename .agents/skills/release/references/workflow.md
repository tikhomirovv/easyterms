# Release workflow (EasyTerms)

## Context

- **Repo:** GitHub `tikhomirovv/easyterms`, branch `main`.
- **Tags:** `v*` prefix (`v0.1.0`). CI release workflow listens on `push.tags: v*`.
- **Docker:** tag push → `.github/workflows/release.yml` → `ghcr.io/tikhomirovv/easyterms` (`latest` + semver tags).
- **CLI:** `gh` (GitHub CLI), not `glab`.
- **No local archive:** release notes exist only on the GitHub Release page.

## 1. Branch check

- If current branch is `main`: continue silently.
- Otherwise: **stop**, show current branch. Wait until the user switches to `main` and `git pull --ff-only`, or explicitly overrides.
- Do not warn when already on `main`.

## 2. Sync

```text
git checkout main
git pull --ff-only
```

1. If no version from the user → ask: «Какая версия релиза?» (accept `0.1.0` or `v0.1.0`; normalize to tag `v0.1.0`).
2. Do not guess semver.

## 3. Commits since last tag

```text
git tag -l "v*" --sort=-version:refname
```

**Previous tag** = latest semver tag **before** the target (ignore pre-release tags like `v0.1.0-rc.1` unless the user says otherwise).

**First release** (no tags yet): use full history on `main`:

```text
git log --pretty=format:"%H|%s|%b|%ad" --date=iso
```

Otherwise:

```text
git log <previousTag>..HEAD --pretty=format:"%H|%s|%b|%ad" --date=iso
```

Optional per commit:

```text
git diff-tree --no-commit-id --name-only -r <hash>
```

Optional: `gh issue view` when commits mention `#N` or PR numbers.

This set is the sole source for both note sections.

## 4. Verify

```text
go test ./...
```

Stop on failure. Do not tag a broken `main`.

Confirm latest `main` CI is green when practical (`gh run list --branch main --limit 1`).

## 5. Analyze and classify

### Technical (always)

Stack changes: Go core, SQLite, Telegram bot, LLM adapter, CI, Docker, migrations, agent skills, docs. English, precise, grouped by area.

### User-facing (only when honest)

Read [user-draft.md](user-draft.md). Include only if a **self-hoster or bot user** would notice.

Internal-only → omit user section. Do not ask «internal or not?» when commits make it obvious.

## 6. Write release notes (ephemeral)

Use a temp path **outside the repo**, e.g. `%TEMP%\easyterms-release-v0.1.0.md` (Windows) or `/tmp/easyterms-release-v0.1.0.md`.

```markdown
## Technical changes

- …

## What's new for users

- …
```

Omit the user heading when step 5 has no user bullets.

Add a compare link when a previous tag exists:

```markdown
**Full changelog:** https://github.com/tikhomirovv/easyterms/compare/<previousTag>...v<version>
```

Do **not** wait for the user to approve the text. Proceed to publish.

## 7. Tag and GitHub Release

Prefer one step — `gh release create` creates the tag if missing:

```text
gh release create v<version> --target main --title "v<version>" --notes-file <temp-notes-path>
```

Or annotated tag first, then release:

```text
git tag -a v<version> -m "Release v<version>"
git push origin v<version>
gh release create v<version> --title "v<version>" --notes-file <temp-notes-path>
```

Delete the temp notes file after success.

Verify:

```text
gh release view v<version>
gh run list --workflow release.yml --limit 3
```

If the Release workflow did not start, check Actions tab and tag name (`v*`).

## 8. After publish

Remind the user:

- Docker image: `ghcr.io/tikhomirovv/easyterms:<version>` and `:latest`
- Self-host: `docker compose pull && docker compose up -d` (see root `docker-compose.yml`)
- First GHCR publish may require setting package visibility to **public** in GitHub UI

## Do not

- Commit release notes into the repo (no `.release-notes/`, no `CHANGELOG.md` unless the user asks).
- Guess the version.
- Tag when not on `main` (unless user overrides).
- Tag when `go test ./...` fails.
- Invent user-visible changes; put SQLite/CI internals in the user section.
- Pause for notes/tag/release approval after the version is known.
