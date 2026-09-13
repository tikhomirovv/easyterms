# Release workflow (EasyTerms)

## Context

- **Repo:** GitHub `tikhomirovv/easyterms`, branch `main`.
- **Tags:** plain semver `0.1.0` (no `v` prefix). CI listens on `push.tags: [0-9]+.[0-9]+.[0-9]+`.
- **Version in code:** `internal/version/version.go` → `const Version` — bump on every release.
- **Docker:** tag push → `.github/workflows/release.yml` → `ghcr.io/tikhomirovv/easyterms` (`latest` + semver tags).
- **CLI:** `gh` (GitHub CLI).
- **No local archive:** release notes exist only on the GitHub Release page.
- **No pre-releases, no drafts.**

## 1. Branch check

- If current branch is `main`: continue silently.
- Otherwise: **stop**, show current branch. Wait until the user switches to `main` and `git pull --ff-only`, or explicitly overrides.

## 2. Sync

```text
git checkout main
git pull --ff-only
```

1. If no version from the user → ask: «Какая версия релиза?» (e.g. `0.1.0`). Do not guess.
2. Normalize: strip leading `v` if present. Tag name = version string.

## 3. How CI fits in

Two pipelines, different jobs:

| Workflow | When | What |
|----------|------|------|
| `CI` | every push/PR to `main` | `go test` only |
| `Release` | push semver **tag** | Docker **build + push** to GHCR (no tests) |

**Before tagging**, verify the commit twice (no third test run on tag):

1. **Local:** `go test ./...`
2. **CI on that commit:** after pushing the version-bump commit to `main`, wait until CI is green on it:

```text
git rev-parse HEAD
gh run list --commit <sha> --workflow CI --limit 3
```

Tag only when the latest `CI` run for that commit is **success**. If still running, wait. If failed, fix before tagging.

The tag pipeline does **not** re-run tests — it only builds and publishes the image. A broken `Dockerfile` is caught here (rare if Dockerfile rarely changes).

## 4. Bump version in code

Edit `internal/version/version.go`:

```go
const Version = "<version>"
```

Stage and commit on `main`:

```text
git add internal/version/version.go
git commit -m "chore(release): <version>"
git push origin main
```

Then tag the **current `main` HEAD** (the commit that includes the version bump).

## 5. Commits since last tag

```text
git tag -l --sort=-version:refname
```

Keep tags that match `X.Y.Z` semver (no `v`). Ignore non-semver tag names.

**Previous tag** = latest semver tag **before** the target.

**First release** (no tags yet): use full history on `main`:

```text
git log --pretty=format:"%H|%s|%b|%ad" --date=iso
```

Otherwise:

```text
git log <previousTag>..HEAD --pretty=format:"%H|%s|%b|%ad" --date=iso
```

Optional: `gh issue view` when commits mention `#N`.

## 6. Verify

```text
go test ./...
```

Stop on failure.

## 7. Analyze and classify

See [user-draft.md](user-draft.md) for the optional user section.

## 8. Write release notes (ephemeral)

Temp file **outside the repo**. English only.

```markdown
## Technical changes

- …

## What's new for users

- …
```

Compare link when a previous tag exists:

```markdown
**Full changelog:** https://github.com/tikhomirovv/easyterms/compare/<previousTag>...<version>
```

## 9. Tag and GitHub Release

Publish immediately (not draft):

```text
gh release create <version> --target main --title "<version>" --notes-file <temp-notes-path>
```

This creates the annotated tag and the GitHub Release in one step.

Delete the temp notes file after success.

Verify:

```text
gh release view <version>
gh run list --workflow release.yml --limit 3
```

## 10. GHCR (Docker registry)

GitHub Container Registry hosts the bot image at `ghcr.io/tikhomirovv/easyterms`.

- Tag push triggers the **Release** workflow automatically.
- For **public repos**, the package is usually public after the first successful push.
- If `docker pull` returns 403/404, open **GitHub → Packages → easyterms → Package settings → Change visibility → Public** (one-time).

Tell the user the image tags: `<version>` and `latest`.

## 11. After publish

Self-host update:

```text
docker compose pull
docker compose up -d
```

## Do not

- Use `v` prefix on tags.
- Create pre-release or draft GitHub releases.
- Commit release notes into the repo.
- Tag when CI on `main` failed or `go test` failed.
- Guess the version.
