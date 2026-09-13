---
name: release
description: >
  Prepare and publish an EasyTerms release on main — ask the target version,
  bump internal/version, collect commits since the previous tag, write English
  GitHub Release notes (technical + optional user-facing), create annotated tag
  X.Y.Z (no v prefix), and publish via gh release create. Tag push triggers GHCR
  Docker image build. Use when the user asks to release, cut a version, publish
  a tag, release notes, or ship a new Docker image version.
---

# Release

EasyTerms **self-hosted Telegram bot** (GitHub `tikhomirovv/easyterms`). Default branch: `main`. Tags are **plain semver** (`0.1.0`) — **no `v` prefix**.

Read references in order:

| File | When |
|------|------|
| [references/workflow.md](references/workflow.md) | Always — full procedure |
| [references/user-draft.md](references/user-draft.md) | When writing or skipping the user-facing section |

## Output — GitHub Release only

**No local release-note files.** All history lives in GitHub Releases.

Release notes are **English only**, markdown, with:

- `## Technical changes` — always
- `## What's new for users` — only when warranted ([user-draft.md](references/user-draft.md))

Write notes to a **temp file outside the repo** (or pass inline to `gh`). Never commit `RELEASE*.md` or `.release-notes/` into the repository.

## Hard rules

- Ask target **version** if not given; do not guess semver.
- Tag format: `<version>` exactly (e.g. `0.1.0`). Strip a leading `v` if the user types `v0.1.0`.
- Update `internal/version/version.go` (`const Version = "…"`) to match the tag, commit on `main`, push, then tag.
- Collect **all commits since the previous release tag** on `main` (first release = entire `main` history).
- Run `go test ./...` locally before tagging; stop if tests fail.
- Tag only a commit on `main` (after version bump is pushed).
- If branch ≠ `main`: stop (see workflow).
- **Release CI** (on tag push) runs `go test` on the **tagged commit**; Docker image is published only if tests pass. No separate “check main CI” step required.
- **No pre-releases**, **no draft releases** — publish immediately via `gh release create`.
- Once the version is known: bump version, analyze commits, write notes, tag, `gh release create` — **no pause** for approval. Only stop if branch ≠ `main`, tests fail, CI failed, or `gh` is missing.
- Pushing the tag starts `.github/workflows/release.yml` → image on `ghcr.io/tikhomirovv/easyterms`.
