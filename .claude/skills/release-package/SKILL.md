---
name: release-package
description: Release @corvina/corvina-seed by creating the GitHub release and its tag, which the GitHub Actions workflows publish. Use when asked to release, publish, tag or bump the version of a package of this repo.
---

# Release a package of corvina-seed

The procedure is shared by the Corvina repos and lives in
[corvina-github-public-actions](https://github.com/Corvina-R-D/corvina-github-public-actions/blob/master/skills/release-package/SKILL.md).
**Read it first** and follow it with the table below:

```bash
gh api repos/Corvina-R-D/corvina-github-public-actions/contents/skills/release-package/SKILL.md \
  -H 'Accept: application/vnd.github.raw'
```

If it can't be read, the essentials: create release and tag together on `master` (prereleases may target a
branch), never bump versions by hand (the workflows do it from the tag), no leading `v` in tags, title = tag, ask the
user to confirm before creating, `gh release create` or else the web UI.

## Packages

Default branch: `master`.

| Package | Tag | Folder | Notes / Latest | Workflow |
|---|---|---|---|---|
| `@corvina/corvina-seed` (main) | `X.Y.Z` | repo root | bare version, generated notes, **Latest** | `publish.yaml` — Publish @corvina/corvina-seed |
