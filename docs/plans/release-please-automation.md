# Release Please Automation - Implementation Plan

**Status**: SAVED FOR LATER
**Created**: 2025-12-07
**Context**: Automate tag generation for Binarius releases using conventional commits

---

## Overview

Release Please (by Google) automates version management and changelog generation based on conventional commits. It creates Release PRs that, when merged, trigger the existing tag-based release workflow.

## Why Release Please

- Project strictly follows conventional commits
- Provides automation with human control (PR review before release)
- Auto-generates changelogs organized by commit type
- No changes needed to existing release workflow
- Industry standard for Go projects

## How It Works

```
Push commits → Release Please creates Release PR → Merge PR → Tag created → Existing release.yml triggers
```

### Version Determination

| Commit Type | Version Bump | Example |
|-------------|--------------|---------|
| `feat:` | Minor | 1.0.0 → 1.1.0 |
| `fix:` | Patch | 1.0.0 → 1.0.1 |
| `feat!:` or `BREAKING CHANGE:` | Major | 1.0.0 → 2.0.0 |
| `chore:`, `docs:`, `ci:` | Patch | 1.0.0 → 1.0.1 |

---

## Files to Create

### 1. `.github/workflows/release-please.yml`

```yaml
name: Release Please

on:
  push:
    branches: [main]

permissions:
  contents: write
  pull-requests: write

jobs:
  release-please:
    name: Release Please
    runs-on: ubuntu-latest
    steps:
      - name: Run Release Please
        uses: googleapis/release-please-action@v4
        with:
          release-type: go
          package-name: binarius
```

### 2. `.release-please-manifest.json`

```json
{
  ".": "0.1.0"
}
```

### 3. `release-please-config.json`

```json
{
  "packages": {
    ".": {
      "release-type": "go",
      "package-name": "binarius",
      "changelog-sections": [
        {"type": "feat", "section": "Features", "hidden": false},
        {"type": "fix", "section": "Bug Fixes", "hidden": false},
        {"type": "perf", "section": "Performance Improvements", "hidden": false},
        {"type": "chore", "section": "Miscellaneous", "hidden": false},
        {"type": "docs", "section": "Documentation", "hidden": false},
        {"type": "test", "section": "Tests", "hidden": false},
        {"type": "ci", "section": "CI/CD", "hidden": false},
        {"type": "refactor", "section": "Code Refactoring", "hidden": false}
      ],
      "bump-minor-pre-major": true,
      "bump-patch-for-minor-pre-major": false
    }
  }
}
```

---

## Implementation Steps

1. Create `.github/workflows/release-please.yml`
2. Create `.release-please-manifest.json` with starting version
3. Create `release-please-config.json` with changelog sections
4. Commit: `chore: add Release Please automation`
5. Push to main
6. Release Please creates first Release PR
7. Review and merge when ready

## No Changes Required to Existing Workflows

- `.github/workflows/release.yml` - Unchanged
- `.github/workflows/quality-and-build.yml` - Unchanged
- `.github/workflows/pr-checks.yml` - Unchanged

---

## New Workflow (After Implementation)

```bash
# Just commit as usual
git commit -m "feat: add kubectl support"
git push origin main

# When ready to release:
# 1. Review the auto-created Release PR
# 2. Merge it
# 3. Done - release happens automatically
```

## Manual Backup Still Works

```bash
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3
```

---

## Agent Delegation (Per Constitution)

- **devops-engineer**: Create workflow and config files
- **git-workflow-manager**: Verify integration
