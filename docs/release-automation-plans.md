# Binarius Release Automation Implementation Plans

This document provides three detailed implementation plans for enhancing Binarius release automation.

## Current State Analysis

### Existing Infrastructure

**Workflows:**
- `.github/workflows/release.yml` - Tag-triggered release workflow
- `.github/workflows/quality-and-build.yml` - Reusable build workflow
- `.github/workflows/pr-checks.yml` - PR validation

**Current Features:**
- Version injection via ldflags (`Version`, `BuildDate`, `GitCommit`)
- Cross-compilation: `linux/amd64`, `linux/arm64`
- SHA256 checksum generation
- Draft releases with manual publish
- Auto-generated release notes from GitHub

**Version Variables (main.go):**
```go
var (
    Version   = "dev"
    BuildDate = "unknown"
    GitCommit = "unknown"
)
```

---

## OPTION A: GoReleaser Migration

GoReleaser is the industry-standard tool for Go project releases. It consolidates build, archive, and release logic into a single declarative configuration.

### Configuration Files

#### `.goreleaser.yaml`

```yaml
# yaml-language-server: $schema=https://goreleaser.com/static/schema.json
# vim: set ts=2 sw=2 tw=0 fo=cnqoj

# GoReleaser configuration for Binarius
# Documentation: https://goreleaser.com

version: 2

# Project metadata
project_name: binarius

# Git configuration
git:
  # Tag sorting for changelog generation
  tag_sort: -version:refname
  # Prerelease handling
  prerelease_suffix: "-"

# Build configuration
before:
  hooks:
    # Ensure dependencies are up to date
    - go mod tidy
    # Verify module checksums
    - go mod verify
    # Run linter before build
    - golangci-lint run --timeout 5m

builds:
  - id: binarius
    # Main package location
    main: .
    # Output binary name
    binary: binarius
    # Build environment
    env:
      - CGO_ENABLED=0
    # Target platforms (Linux only per project requirements)
    goos:
      - linux
    goarch:
      - amd64
      - arm64
    # Ldflags for version injection - matches current main.go variables
    ldflags:
      - -s -w
      - -X main.Version={{.Version}}
      - -X main.BuildDate={{.Date}}
      - -X main.GitCommit={{.FullCommit}}
    # Reproducible builds
    mod_timestamp: "{{ .CommitTimestamp }}"
    # Build flags
    flags:
      - -trimpath

# Archive configuration
archives:
  - id: binarius-archives
    builds:
      - binarius
    # Archive naming template
    name_template: >-
      {{ .ProjectName }}-
      {{- .Version }}-
      {{- .Os }}-
      {{- .Arch }}
    # Use tar.gz for Linux
    format: tar.gz
    # Files to include in archive
    files:
      - LICENSE*
      - README*
      - CHANGELOG*
    # Strip parent directories
    wrap_in_directory: false

# Checksum configuration
checksum:
  name_template: "checksums.txt"
  algorithm: sha256
  # Include individual checksums per file
  extra_files:
    - glob: ./dist/*.tar.gz

# Snapshot configuration (for non-tag builds)
snapshot:
  version_template: "{{ incpatch .Version }}-next"

# Changelog configuration
changelog:
  # Sort by commit message
  sort: asc
  # Use conventional commits for categorization
  use: conventional
  # Filter commits for changelog
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^chore\\(deps\\):"
      - "^ci:"
      - Merge pull request
      - Merge branch
  # Group commits by type
  groups:
    - title: "Breaking Changes"
      regexp: '^.*?!:|^.*?BREAKING CHANGE:'
      order: 0
    - title: "Features"
      regexp: '^feat(\(.+\))?:'
      order: 1
    - title: "Bug Fixes"
      regexp: '^fix(\(.+\))?:'
      order: 2
    - title: "Performance Improvements"
      regexp: '^perf(\(.+\))?:'
      order: 3
    - title: "Refactoring"
      regexp: '^refactor(\(.+\))?:'
      order: 4
    - title: "Other Changes"
      order: 999

# GitHub release configuration
release:
  github:
    owner: nixknight
    name: binarius
  # Release name template
  name_template: "Release {{.Tag}}"
  # Draft mode for manual review before publishing
  draft: true
  # Prerelease detection (e.g., v1.0.0-rc1)
  prerelease: auto
  # Release header template
  header: |
    ## Binarius {{ .Tag }}

    Universal binary version manager for Linux.

    ### Installation

    Download the appropriate archive for your architecture:

    **Linux AMD64 (x86_64)**:
    ```bash
    curl -LO https://github.com/nixknight/binarius/releases/download/{{ .Tag }}/binarius-{{ .Version }}-linux-amd64.tar.gz
    tar -xzf binarius-{{ .Version }}-linux-amd64.tar.gz
    sudo mv binarius /usr/local/bin/
    ```

    **Linux ARM64 (aarch64)**:
    ```bash
    curl -LO https://github.com/nixknight/binarius/releases/download/{{ .Tag }}/binarius-{{ .Version }}-linux-arm64.tar.gz
    tar -xzf binarius-{{ .Version }}-linux-arm64.tar.gz
    sudo mv binarius /usr/local/bin/
    ```

    ### Verify Installation

    ```bash
    # Download checksum file
    curl -LO https://github.com/nixknight/binarius/releases/download/{{ .Tag }}/checksums.txt

    # Verify checksum (replace with your downloaded archive)
    sha256sum -c checksums.txt --ignore-missing

    # Check version
    binarius version
    ```

  footer: |
    ---
    **Full Changelog**: https://github.com/nixknight/binarius/compare/{{ .PreviousTag }}...{{ .Tag }}
  # Extra files to upload (beyond archives)
  extra_files:
    - glob: ./dist/checksums.txt

# Announce configuration (optional - can be extended)
announce:
  skip: true

# Milestones (close on release)
milestones:
  - close: true
    fail_on_error: false
    name_template: "{{ .Tag }}"

# SBOMs (Software Bill of Materials)
sboms:
  - artifacts: archive
    documents:
      - "${artifact}.sbom.json"

# Metadata for provenance
metadata:
  mod_timestamp: "{{ .CommitTimestamp }}"
```

#### `.github/workflows/release-goreleaser.yml`

```yaml
name: Release with GoReleaser

on:
  push:
    tags:
      - "v*.*.*"

permissions:
  contents: write
  packages: write
  id-token: write  # For SBOM signing

jobs:
  release:
    name: Release
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Required for changelog generation
          fetch-tags: true

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.22"
          cache: true

      - name: Install golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: v1.61
          # Only install, don't run (GoReleaser handles this)
          args: --version

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: "~> v2"
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Upload artifacts
        uses: actions/upload-artifact@v4
        if: always()
        with:
          name: dist
          path: dist/
          retention-days: 7
```

### Migration Guide

**Step 1: Install GoReleaser locally**
```bash
# Using Go
go install github.com/goreleaser/goreleaser/v2@latest

# Or using Homebrew (Linux)
brew install goreleaser
```

**Step 2: Add configuration file**
```bash
# Create .goreleaser.yaml in project root
# (use the configuration above)
```

**Step 3: Validate configuration**
```bash
goreleaser check
```

**Step 4: Test locally (dry run)**
```bash
# Create a test tag
git tag -a v0.1.0-test -m "Test release"

# Run in snapshot mode (no publishing)
goreleaser release --snapshot --clean

# Check output in dist/ directory
ls -la dist/
```

**Step 5: Update GitHub workflow**
```bash
# Replace .github/workflows/release.yml with release-goreleaser.yml
# Or rename files as needed
```

**Step 6: Clean up old workflow**
```bash
# After verifying GoReleaser works, remove old files:
# - .github/workflows/release.yml (old)
# - .github/workflows/quality-and-build.yml (if not used elsewhere)
```

**Step 7: Create first release**
```bash
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0
```

### Pros and Cons

**Pros:**
- Industry-standard tool with excellent documentation
- Single declarative configuration file
- Built-in changelog generation from conventional commits
- SBOM generation for security/compliance
- Supports signing, announcements, and many integrations
- Active community and maintenance
- Simplified workflow (one action instead of multiple)
- Built-in checksum generation and verification
- Reproducible builds with trimpath and timestamps

**Cons:**
- Additional dependency (GoReleaser)
- Learning curve for configuration
- Less flexible than custom scripts for edge cases
- Requires conventional commits for best changelog results
- Draft releases still require manual publishing

---

## OPTION B: GoReleaser + Release Please (Full Automation)

This option extends Option A with fully automated versioning and release PR creation using Google's Release Please.

### How It Works

1. **Developers commit** using conventional commits (`feat:`, `fix:`, etc.)
2. **Release Please monitors** the main branch and creates/updates a Release PR
3. **Release PR accumulates** changes and auto-updates version based on commits
4. **Maintainer merges** the Release PR when ready
5. **Merge triggers** tag creation automatically
6. **Tag triggers** GoReleaser to build and publish

### Configuration Files

#### `.github/workflows/release-please.yml`

```yaml
name: Release Please

on:
  push:
    branches:
      - main

permissions:
  contents: write
  pull-requests: write

jobs:
  release-please:
    name: Release Please
    runs-on: ubuntu-latest
    outputs:
      release_created: ${{ steps.release.outputs.release_created }}
      tag_name: ${{ steps.release.outputs.tag_name }}
      version: ${{ steps.release.outputs.version }}

    steps:
      - name: Run Release Please
        id: release
        uses: googleapis/release-please-action@v4
        with:
          # Release configuration
          release-type: go
          # Token for creating PRs and releases
          token: ${{ secrets.GITHUB_TOKEN }}

  # GoReleaser runs only when a release is created
  goreleaser:
    name: GoReleaser
    needs: release-please
    if: ${{ needs.release-please.outputs.release_created == 'true' }}
    runs-on: ubuntu-latest
    permissions:
      contents: write
      packages: write
      id-token: write

    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0
          fetch-tags: true
          # Checkout the tag created by Release Please
          ref: ${{ needs.release-please.outputs.tag_name }}

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.22"
          cache: true

      - name: Install golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: v1.61
          args: --version

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: "~> v2"
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Upload artifacts
        uses: actions/upload-artifact@v4
        if: always()
        with:
          name: dist
          path: dist/
          retention-days: 7
```

#### `.release-please-manifest.json`

```json
{
  ".": "0.1.0"
}
```

#### `release-please-config.json`

```json
{
  "$schema": "https://raw.githubusercontent.com/googleapis/release-please/main/schemas/config.json",
  "packages": {
    ".": {
      "release-type": "go",
      "bump-minor-pre-major": true,
      "bump-patch-for-minor-pre-major": true,
      "draft": false,
      "prerelease": false,
      "changelog-sections": [
        {
          "type": "feat",
          "section": "Features",
          "hidden": false
        },
        {
          "type": "fix",
          "section": "Bug Fixes",
          "hidden": false
        },
        {
          "type": "perf",
          "section": "Performance Improvements",
          "hidden": false
        },
        {
          "type": "refactor",
          "section": "Code Refactoring",
          "hidden": false
        },
        {
          "type": "docs",
          "section": "Documentation",
          "hidden": false
        },
        {
          "type": "chore",
          "section": "Miscellaneous Chores",
          "hidden": true
        },
        {
          "type": "ci",
          "section": "Continuous Integration",
          "hidden": true
        },
        {
          "type": "test",
          "section": "Tests",
          "hidden": true
        }
      ],
      "extra-files": [
        {
          "type": "json",
          "path": "version.json",
          "jsonpath": "$.version"
        }
      ]
    }
  },
  "versioning": "default",
  "include-component-in-tag": false,
  "include-v-in-tag": true,
  "tag-separator": "",
  "pull-request-title-pattern": "chore(release): release ${version}",
  "pull-request-header": "## Release ${version}\n\nThis PR was automatically generated by Release Please.",
  "sequential-calls": true
}
```

#### `version.json` (Optional - for programmatic version access)

```json
{
  "version": "0.1.0"
}
```

#### Updated `.goreleaser.yaml` for Release Please

```yaml
# yaml-language-server: $schema=https://goreleaser.com/static/schema.json
version: 2

project_name: binarius

git:
  tag_sort: -version:refname
  prerelease_suffix: "-"

before:
  hooks:
    - go mod tidy
    - go mod verify
    - golangci-lint run --timeout 5m

builds:
  - id: binarius
    main: .
    binary: binarius
    env:
      - CGO_ENABLED=0
    goos:
      - linux
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X main.Version={{.Version}}
      - -X main.BuildDate={{.Date}}
      - -X main.GitCommit={{.FullCommit}}
    mod_timestamp: "{{ .CommitTimestamp }}"
    flags:
      - -trimpath

archives:
  - id: binarius-archives
    builds:
      - binarius
    name_template: >-
      {{ .ProjectName }}-
      {{- .Version }}-
      {{- .Os }}-
      {{- .Arch }}
    format: tar.gz
    files:
      - LICENSE*
      - README*
      - CHANGELOG*
    wrap_in_directory: false

checksum:
  name_template: "checksums.txt"
  algorithm: sha256

snapshot:
  version_template: "{{ incpatch .Version }}-next"

# Minimal changelog since Release Please handles it
changelog:
  disable: true

release:
  github:
    owner: nixknight
    name: binarius
  name_template: "Release {{.Tag}}"
  # Not draft since Release Please already created the release
  draft: false
  prerelease: auto
  # Skip body since Release Please provides release notes
  skip_upload: false
  header: |
    ## Installation

    Download the appropriate archive for your architecture:

    **Linux AMD64 (x86_64)**:
    ```bash
    curl -LO https://github.com/nixknight/binarius/releases/download/{{ .Tag }}/binarius-{{ .Version }}-linux-amd64.tar.gz
    tar -xzf binarius-{{ .Version }}-linux-amd64.tar.gz
    sudo mv binarius /usr/local/bin/
    ```

    **Linux ARM64 (aarch64)**:
    ```bash
    curl -LO https://github.com/nixknight/binarius/releases/download/{{ .Tag }}/binarius-{{ .Version }}-linux-arm64.tar.gz
    tar -xzf binarius-{{ .Version }}-linux-arm64.tar.gz
    sudo mv binarius /usr/local/bin/
    ```

    ### Verify Installation

    ```bash
    curl -LO https://github.com/nixknight/binarius/releases/download/{{ .Tag }}/checksums.txt
    sha256sum -c checksums.txt --ignore-missing
    binarius version
    ```

  footer: ""
  extra_files:
    - glob: ./dist/checksums.txt

# Update release instead of creating new one
# Release Please already created the release, GoReleaser uploads assets
mode: replace

sboms:
  - artifacts: archive
    documents:
      - "${artifact}.sbom.json"

metadata:
  mod_timestamp: "{{ .CommitTimestamp }}"
```

### Migration Guide

**Step 1: Install and configure GoReleaser (as in Option A)**
```bash
go install github.com/goreleaser/goreleaser/v2@latest
```

**Step 2: Add Release Please configuration files**
```bash
# Create manifest file
echo '{ ".": "0.1.0" }' > .release-please-manifest.json

# Create config file (use configuration above)
# release-please-config.json

# Create version.json (optional)
echo '{ "version": "0.1.0" }' > version.json
```

**Step 3: Add GoReleaser configuration**
```bash
# Use the modified .goreleaser.yaml from above
```

**Step 4: Add the workflow file**
```bash
# Create .github/workflows/release-please.yml
```

**Step 5: Enable conventional commits**
```bash
# Install commitlint (see Option C for full setup)
npm install -g @commitlint/cli @commitlint/config-conventional

# Create commitlint config
echo "module.exports = {extends: ['@commitlint/config-conventional']}" > commitlint.config.js
```

**Step 6: Bootstrap Release Please**
```bash
# Merge the workflow to main branch
git add .
git commit -m "feat: add release automation with Release Please and GoReleaser"
git push origin main

# Release Please will create its first PR on next push
```

**Step 7: Workflow usage**
```bash
# Regular development
git commit -m "feat: add new tool support for kubectl"
git push origin main
# Release Please automatically updates or creates Release PR

# When ready to release
# Simply merge the Release PR
# Tags are created automatically
# GoReleaser runs automatically
```

### Pros and Cons

**Pros:**
- Fully automated release process
- No manual tag creation required
- CHANGELOG.md maintained automatically
- Semantic versioning from commits
- Release PRs provide review opportunity
- Consistent release notes format
- Version bumping follows semver rules automatically
- Integration with GoReleaser for builds

**Cons:**
- More complex setup with multiple tools
- Requires strict conventional commits discipline
- Two dependencies (Release Please + GoReleaser)
- More configuration files to maintain
- Slower feedback loop (PR-based releases)
- May be overkill for small projects
- Requires understanding of both tools

---

## OPTION C: Enhance Current Workflow

This option improves the existing workflow without replacing it, adding conventional commit enforcement and automated changelog generation.

### Configuration Files

#### `.github/workflows/commitlint.yml`

```yaml
name: Commit Lint

on:
  pull_request:
    branches:
      - main
  push:
    branches:
      - main

jobs:
  commitlint:
    name: Lint Commit Messages
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Fetch all history for commitlint

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: "20"

      - name: Install commitlint
        run: |
          npm install --save-dev @commitlint/cli @commitlint/config-conventional

      - name: Validate commits
        run: |
          # For PRs, check all commits in the PR
          if [ "${{ github.event_name }}" == "pull_request" ]; then
            npx commitlint --from ${{ github.event.pull_request.base.sha }} --to ${{ github.event.pull_request.head.sha }} --verbose
          else
            # For pushes, check the latest commit
            npx commitlint --from HEAD~1 --to HEAD --verbose
          fi
```

#### `commitlint.config.js`

```javascript
// Commitlint configuration for conventional commits
// Documentation: https://commitlint.js.org/

module.exports = {
  extends: ['@commitlint/config-conventional'],
  rules: {
    // Type must be one of the following
    'type-enum': [
      2, // Error level
      'always',
      [
        'feat',     // New feature
        'fix',      // Bug fix
        'docs',     // Documentation only
        'style',    // Code style (formatting, semicolons, etc)
        'refactor', // Code refactoring
        'perf',     // Performance improvement
        'test',     // Adding or updating tests
        'build',    // Build system or dependencies
        'ci',       // CI/CD changes
        'chore',    // Maintenance tasks
        'revert',   // Reverting previous commit
      ],
    ],
    // Type must be lowercase
    'type-case': [2, 'always', 'lower-case'],
    // Type cannot be empty
    'type-empty': [2, 'never'],
    // Subject cannot be empty
    'subject-empty': [2, 'never'],
    // Subject max length
    'subject-max-length': [2, 'always', 100],
    // Subject must not end with period
    'subject-full-stop': [2, 'never', '.'],
    // Body max line length
    'body-max-line-length': [2, 'always', 200],
    // Footer max line length
    'footer-max-line-length': [2, 'always', 200],
  },
  // Help message for failed commits
  helpUrl:
    'https://www.conventionalcommits.org/en/v1.0.0/',
};
```

#### `package.json` (minimal for commitlint)

```json
{
  "name": "binarius-dev",
  "private": true,
  "description": "Development dependencies for Binarius",
  "devDependencies": {
    "@commitlint/cli": "^19.0.0",
    "@commitlint/config-conventional": "^19.0.0",
    "husky": "^9.0.0"
  },
  "scripts": {
    "prepare": "husky"
  }
}
```

#### `.husky/commit-msg`

```bash
#!/usr/bin/env sh
. "$(dirname -- "$0")/_/husky.sh"

npx --no -- commitlint --edit ${1}
```

#### `.github/workflows/changelog.yml`

```yaml
name: Generate Changelog

on:
  workflow_dispatch:
    inputs:
      version:
        description: 'Version for changelog (e.g., v1.0.0)'
        required: true
        type: string

permissions:
  contents: write
  pull-requests: write

jobs:
  changelog:
    name: Generate Changelog
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0
          fetch-tags: true

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: "20"

      - name: Install changelog generator
        run: npm install -g conventional-changelog-cli

      - name: Generate changelog
        run: |
          # Generate changelog for the new version
          conventional-changelog -p conventionalcommits -i CHANGELOG.md -s -r 0

      - name: Create PR with changelog
        uses: peter-evans/create-pull-request@v6
        with:
          token: ${{ secrets.GITHUB_TOKEN }}
          commit-message: "docs: update CHANGELOG.md for ${{ inputs.version }}"
          title: "docs: update CHANGELOG.md for ${{ inputs.version }}"
          body: |
            This PR updates the CHANGELOG.md file for the upcoming release ${{ inputs.version }}.

            Please review the generated changelog and make any necessary adjustments before merging.
          branch: changelog-${{ inputs.version }}
          base: main
          labels: documentation
```

#### Updated `.github/workflows/release.yml` (Enhanced)

```yaml
name: Release

on:
  push:
    tags:
      - 'v*.*.*'

permissions:
  contents: write

jobs:
  # Validate tag follows semver
  validate-tag:
    name: Validate Tag
    runs-on: ubuntu-latest
    outputs:
      version: ${{ steps.extract.outputs.version }}
      is-prerelease: ${{ steps.extract.outputs.is-prerelease }}

    steps:
      - name: Extract and validate version
        id: extract
        run: |
          VERSION=${GITHUB_REF#refs/tags/}
          echo "version=${VERSION}" >> $GITHUB_OUTPUT

          # Check if prerelease (contains - like v1.0.0-rc1)
          if [[ "$VERSION" == *"-"* ]]; then
            echo "is-prerelease=true" >> $GITHUB_OUTPUT
          else
            echo "is-prerelease=false" >> $GITHUB_OUTPUT
          fi

          # Validate semver format
          if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?$ ]]; then
            echo "Error: Invalid version format: $VERSION"
            echo "Expected format: vX.Y.Z or vX.Y.Z-prerelease"
            exit 1
          fi

          echo "Valid version: $VERSION"

  # Extract version metadata from git tag
  extract-version:
    name: Extract Version Metadata
    needs: validate-tag
    runs-on: ubuntu-latest
    outputs:
      version: ${{ steps.extract.outputs.version }}
      build-date: ${{ steps.extract.outputs.build-date }}
      git-commit: ${{ steps.extract.outputs.git-commit }}

    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Extract metadata
        id: extract
        run: |
          VERSION=${GITHUB_REF#refs/tags/}
          echo "version=${VERSION}" >> $GITHUB_OUTPUT

          BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
          echo "build-date=${BUILD_DATE}" >> $GITHUB_OUTPUT

          GIT_COMMIT=$(git rev-parse HEAD)
          echo "git-commit=${GIT_COMMIT}" >> $GITHUB_OUTPUT

          echo "Building version ${VERSION} (commit: ${GIT_COMMIT:0:7}) at ${BUILD_DATE}"

  # Run quality checks and build binaries
  quality-and-build:
    name: Quality Checks & Build
    needs: [validate-tag, extract-version]
    uses: ./.github/workflows/quality-and-build.yml
    with:
      inject-version: true
      version: ${{ needs.extract-version.outputs.version }}
      build-date: ${{ needs.extract-version.outputs.build-date }}
      git-commit: ${{ needs.extract-version.outputs.git-commit }}
      generate-checksums: true
      upload-artifacts: true
      artifact-retention-days: 30
      go-version: '1.22'
      golangci-lint-version: 'v1.61'

  # Generate changelog from commits
  generate-changelog:
    name: Generate Changelog
    needs: validate-tag
    runs-on: ubuntu-latest
    outputs:
      changelog: ${{ steps.changelog.outputs.changelog }}

    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0
          fetch-tags: true

      - name: Get previous tag
        id: prev-tag
        run: |
          PREV_TAG=$(git describe --tags --abbrev=0 HEAD^ 2>/dev/null || echo "")
          echo "tag=${PREV_TAG}" >> $GITHUB_OUTPUT
          echo "Previous tag: ${PREV_TAG:-none}"

      - name: Generate changelog
        id: changelog
        run: |
          VERSION=${GITHUB_REF#refs/tags/}
          PREV_TAG="${{ steps.prev-tag.outputs.tag }}"

          # Generate changelog from commits
          if [ -n "$PREV_TAG" ]; then
            RANGE="${PREV_TAG}..HEAD"
          else
            RANGE="HEAD"
          fi

          echo "Generating changelog for range: $RANGE"

          # Create changelog with categories
          CHANGELOG=""

          # Features
          FEATURES=$(git log ${RANGE} --pretty=format:"- %s (%h)" --grep="^feat" 2>/dev/null || true)
          if [ -n "$FEATURES" ]; then
            CHANGELOG="${CHANGELOG}### Features\n\n${FEATURES}\n\n"
          fi

          # Bug Fixes
          FIXES=$(git log ${RANGE} --pretty=format:"- %s (%h)" --grep="^fix" 2>/dev/null || true)
          if [ -n "$FIXES" ]; then
            CHANGELOG="${CHANGELOG}### Bug Fixes\n\n${FIXES}\n\n"
          fi

          # Performance
          PERF=$(git log ${RANGE} --pretty=format:"- %s (%h)" --grep="^perf" 2>/dev/null || true)
          if [ -n "$PERF" ]; then
            CHANGELOG="${CHANGELOG}### Performance Improvements\n\n${PERF}\n\n"
          fi

          # Other changes
          OTHER=$(git log ${RANGE} --pretty=format:"- %s (%h)" --grep="^refactor\|^build\|^ci" 2>/dev/null || true)
          if [ -n "$OTHER" ]; then
            CHANGELOG="${CHANGELOG}### Other Changes\n\n${OTHER}\n\n"
          fi

          # Save changelog (handle multiline)
          echo "changelog<<EOF" >> $GITHUB_OUTPUT
          echo -e "$CHANGELOG" >> $GITHUB_OUTPUT
          echo "EOF" >> $GITHUB_OUTPUT

  # Create GitHub release with artifacts
  create-release:
    name: Create GitHub Release
    needs: [validate-tag, extract-version, quality-and-build, generate-changelog]
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Download build artifacts
        uses: actions/download-artifact@v4
        with:
          name: ${{ needs.quality-and-build.outputs.artifact-name }}

      - name: Verify artifacts downloaded
        run: |
          echo "Downloaded artifacts:"
          ls -lh

          for file in binarius-linux-amd64 binarius-linux-arm64 SHA256SUMS; do
            if [ ! -f "$file" ]; then
              echo "Error: Missing required file: $file"
              exit 1
            fi
          done

          echo "Verifying checksums..."
          sha256sum -c SHA256SUMS

      - name: Create release
        uses: softprops/action-gh-release@v2
        with:
          draft: true
          prerelease: ${{ needs.validate-tag.outputs.is-prerelease }}
          name: "Release ${{ needs.extract-version.outputs.version }}"
          body: |
            ## Binarius ${{ needs.extract-version.outputs.version }}

            Universal binary version manager for Linux.

            ${{ needs.generate-changelog.outputs.changelog }}

            ### Installation

            Download the appropriate binary for your architecture:

            **Linux AMD64 (x86_64)**:
            ```bash
            curl -LO https://github.com/nixknight/binarius/releases/download/${{ needs.extract-version.outputs.version }}/binarius-linux-amd64
            chmod +x binarius-linux-amd64
            sudo mv binarius-linux-amd64 /usr/local/bin/binarius
            ```

            **Linux ARM64 (aarch64)**:
            ```bash
            curl -LO https://github.com/nixknight/binarius/releases/download/${{ needs.extract-version.outputs.version }}/binarius-linux-arm64
            chmod +x binarius-linux-arm64
            sudo mv binarius-linux-arm64 /usr/local/bin/binarius
            ```

            ### Verify Installation

            ```bash
            curl -LO https://github.com/nixknight/binarius/releases/download/${{ needs.extract-version.outputs.version }}/SHA256SUMS
            sha256sum -c SHA256SUMS --ignore-missing
            binarius version
            ```

          files: |
            binarius-linux-amd64
            binarius-linux-arm64
            SHA256SUMS
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

#### `CHANGELOG.md` (Template)

```markdown
# Changelog

All notable changes to Binarius will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

### Changed

### Deprecated

### Removed

### Fixed

### Security

---

## [0.1.0] - YYYY-MM-DD

### Added
- Initial release
- Support for terraform, opentofu (tofu), and terragrunt
- Symlink-based zero-overhead version management
- Commands: init, install, use, list, uninstall, info

[Unreleased]: https://github.com/nixknight/binarius/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/nixknight/binarius/releases/tag/v0.1.0
```

### Migration Guide

**Step 1: Initialize npm (for commitlint)**
```bash
# Initialize npm in project root (if not exists)
npm init -y

# Install commitlint dependencies
npm install --save-dev @commitlint/cli @commitlint/config-conventional husky
```

**Step 2: Configure Husky for pre-commit hooks**
```bash
# Initialize Husky
npx husky init

# Create commit-msg hook
echo '#!/usr/bin/env sh
. "$(dirname -- "$0")/_/husky.sh"
npx --no -- commitlint --edit ${1}' > .husky/commit-msg

chmod +x .husky/commit-msg
```

**Step 3: Add commitlint configuration**
```bash
# Create commitlint.config.js (use configuration above)
```

**Step 4: Add workflow files**
```bash
# Add .github/workflows/commitlint.yml
# Update .github/workflows/release.yml
# Add .github/workflows/changelog.yml
```

**Step 5: Create initial CHANGELOG.md**
```bash
# Create CHANGELOG.md with template
```

**Step 6: Update .gitignore**
```bash
# Add Node.js artifacts to .gitignore
echo "node_modules/" >> .gitignore
echo "package-lock.json" >> .gitignore  # Optional
```

**Step 7: Test the setup**
```bash
# Test commitlint locally
echo "test" | npx commitlint  # Should fail
echo "feat: test feature" | npx commitlint  # Should pass

# Try a commit
git commit -m "invalid message"  # Should be rejected
git commit -m "feat: add commitlint configuration"  # Should succeed
```

### Pros and Cons

**Pros:**
- Minimal changes to existing workflow
- Enforces commit message standards
- Preserves existing build process
- Gradual improvement path
- No new build tools required
- Familiar workflow for team
- Can migrate to Option A or B later

**Cons:**
- Still requires manual tag creation
- Changelog generation is semi-automated
- Less sophisticated than dedicated tools
- Node.js dependency for commitlint only
- Pre-commit hooks can slow down workflow
- Manual version management

---

## Comparison Summary

| Feature | Option A (GoReleaser) | Option B (GoReleaser + Release Please) | Option C (Enhanced Current) |
|---------|----------------------|----------------------------------------|----------------------------|
| **Complexity** | Medium | High | Low |
| **Setup Time** | 1-2 hours | 2-4 hours | 30 min - 1 hour |
| **New Dependencies** | GoReleaser | GoReleaser, Release Please | commitlint, husky |
| **Automation Level** | High | Full | Medium |
| **Version Bumping** | Manual tags | Automatic | Manual tags |
| **Changelog** | Auto from commits | Auto with PR | Semi-auto |
| **Release Flow** | Tag -> Build -> Draft | PR merge -> Tag -> Build -> Publish | Tag -> Build -> Draft |
| **Learning Curve** | Medium | High | Low |
| **Maintenance** | Low | Medium | Low |
| **SBOM Support** | Yes | Yes | No |
| **Best For** | Most projects | Frequent releases, teams | Simple projects, gradual adoption |

## Recommendation

**For Binarius specifically, I recommend Option A (GoReleaser Migration)** for the following reasons:

1. **Right balance**: Provides significant automation without excessive complexity
2. **Industry standard**: GoReleaser is widely adopted in the Go ecosystem
3. **Maintainability**: Single configuration file, well-documented
4. **Growth path**: Can add Release Please later if needed
5. **Features**: SBOM, checksums, archives, reproducible builds
6. **Current state alignment**: Matches current draft release workflow

**Option B** is excellent but may be overkill for a project in early stages. Consider migrating to it once release frequency increases and the team is comfortable with conventional commits.

**Option C** is a good stepping stone if you want to start with conventional commits first before adopting GoReleaser.

---

## File Locations Summary

### Option A Files
- `/.goreleaser.yaml` - GoReleaser configuration
- `/.github/workflows/release-goreleaser.yml` - GitHub Actions workflow

### Option B Files
- `/.goreleaser.yaml` - GoReleaser configuration (modified)
- `/.github/workflows/release-please.yml` - Combined Release Please + GoReleaser workflow
- `/.release-please-manifest.json` - Version manifest
- `/release-please-config.json` - Release Please configuration
- `/version.json` - Optional version file

### Option C Files
- `/commitlint.config.js` - Commitlint configuration
- `/package.json` - NPM dependencies
- `/.husky/commit-msg` - Git hook
- `/.github/workflows/commitlint.yml` - Commit validation workflow
- `/.github/workflows/changelog.yml` - Changelog generation workflow
- `/.github/workflows/release.yml` - Enhanced release workflow
- `/CHANGELOG.md` - Changelog file
