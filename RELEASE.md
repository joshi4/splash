# Release Process

This document describes how to create releases for splash using the automated GitHub Actions workflow.

## Overview

The project uses [GoReleaser](https://goreleaser.com/) to build multi-platform binaries and publish them to:
- GitHub Releases with downloadable assets
- Homebrew tap repository (`joshi4/homebrew-tap`)

## Prerequisites

Before creating releases, ensure you have:

1. **Homebrew Tap Repository**: Create a repository named `homebrew-tap` under your GitHub account
2. **GitHub Token**: Create a personal access token with repository permissions for the Homebrew tap
3. **Repository Secret**: Add the token as `HOMEBREW_TAP_GITHUB_TOKEN` in repository secrets

## Release Process

### Automatic Releases

Releases are automatically created when you push a semantic version tag:

```bash
# Create and push a new tag
git tag v1.0.0
git push origin v1.0.0
```

This triggers the GitHub Actions workflow that:
1. Runs all tests
2. Builds binaries for multiple platforms (Linux, macOS, Windows)
3. Creates checksums and signs releases
4. Publishes to GitHub Releases
5. Updates the Homebrew tap repository

### Manual Release Testing

To test the release process locally:

```bash
# Install GoReleaser
go install github.com/goreleaser/goreleaser@latest

# Test configuration
goreleaser check

# Build snapshot (no publishing)
goreleaser build --snapshot --clean

# Full release dry-run
goreleaser release --snapshot --clean
```

## Semantic Versioning

Follow [Semantic Versioning](https://semver.org/) principles:

- `v1.0.0` - Major release (breaking changes)
- `v1.1.0` - Minor release (new features, backward compatible)
- `v1.0.1` - Patch release (bug fixes, backward compatible)

## Supported Platforms

The release process builds binaries for:

- **Linux**: amd64, arm64, 386
- **macOS**: amd64, arm64 (with universal binary)
- **Windows**: amd64, 386

## Installation Methods

After release, users can install splash via:

### Homebrew (Recommended)
```bash
brew install joshi4/tap/splash
```

### Direct Download
Download from [GitHub Releases](https://github.com/joshi4/splash/releases)

### Go Install
```bash
go install github.com/joshi4/splash@latest
```

### Upgrade Command
```bash
splash upgrade
```

## Release Assets

Each release includes:
- Compiled binaries for all supported platforms
- `checksums.txt` file for verification
- Source code archives
- Automated changelog

## Troubleshooting

### Homebrew Formula Issues
If Homebrew installation fails:
1. Check the `homebrew-tap` repository was updated
2. Verify the formula syntax is correct
3. Test installation manually: `brew install --HEAD joshi4/tap/splash`

### Binary Download Issues
If direct downloads don't work:
1. Verify checksums match
2. Check platform compatibility
3. Ensure binary has execute permissions: `chmod +x splash`

### Version Issues
If version information is incorrect:
1. Check that ldflags are properly set in `.goreleaser.yaml`
2. Verify the git tag format matches semantic versioning
3. Ensure the tag was pushed to the repository

## GitHub Actions Workflows

The repository includes several workflows:

- **`release.yml`**: Triggered by version tags, handles full release process
- **`test.yml`**: Runs on PRs and pushes, ensures code quality
- **`draft-release.yml`**: Creates draft releases from main branch
- **`dependabot-auto-merge.yml`**: Automatically merges dependency updates

## Configuration Files

- **`.goreleaser.yaml`**: GoReleaser configuration for builds and releases
- **`.github/workflows/`**: GitHub Actions workflow definitions
- **`.github/dependabot.yml`**: Dependency update configuration
