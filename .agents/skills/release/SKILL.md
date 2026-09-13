---
name: release
description: >
  Prepare and publish an EasyTerms release on main — ask the target version,
  collect commits since the previous tag, write English GitHub Release notes
  (technical + optional user-facing), create annotated tag vX.Y.Z, and publish
  via gh release create. Tag push triggers GHCR Docker image build. Use when
  the user asks to release, cut a version, publish a tag, release notes, or
  ship a new Docker image version.
---

# Release

EasyTerms **self-hosted Telegram bot** (GitHub `tikhomirovv/easyterms`). Default branch: `main`. Tags use **`v` prefix** (`v0.1.0`). Not GitLab.

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
- Tag format: `v<version>` (e.g. user says `0.1.0` → tag `v0.1.0`). Match existing tags.
- Collect **all commits since the previous release tag** on `main`.
- Run `go test ./...` before tagging; stop if tests fail.
- If branch ≠ `main`: stop (see workflow).
- Once the version is known: analyze commits, write notes, tag, `gh release create` — **no pause** for notes/tag approval. Only stop if branch ≠ `main`, tests fail, or `gh` is missing.
- Pushing the tag starts `.github/workflows/release.yml` → public image on `ghcr.io/tikhomirovv/easyterms`.

## No version file

This Go repo has **no** `package.json` / app version constant to bump. The tag **is** the version. Do not invent a version bump commit unless the user explicitly asks to embed version in code later.
