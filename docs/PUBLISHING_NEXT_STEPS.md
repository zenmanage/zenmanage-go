# Publishing Next Steps for zenmanage-go

This guide covers publishing and discoverability for public Go package consumers.

## 1. Repository Readiness

1. Ensure the module path in go.mod matches the final repository URL.
2. Ensure LICENSE, README, CHANGELOG, and CONTRIBUTING are present.
3. Confirm package-level and exported symbol documentation exists.
4. Confirm tests pass and coverage is healthy.

~~~bash
go test ./... -coverprofile=coverage.out
~~~

## 2. Semantic Version Tagging

Go package distribution is tag-driven.

1. Update CHANGELOG.md with release notes.
2. Commit release changes.
3. Create an annotated tag.

~~~bash
git tag -a v1.0.0 -m "v1.0.0"
git push origin main --tags
~~~

## 3. Publish on GitHub

1. Create a GitHub release from the tag.
2. Include highlights and migration notes.
3. Attach generated artifacts if needed.

Optional automation:

- Use GitHub Actions to run tests and create releases on tag push.

## 4. Publish to Go Ecosystem Indexes

Go does not use a central upload workflow like npm or PyPI. Publication happens via VCS tags.

### pkg.go.dev

- pkg.go.dev auto-indexes public tagged modules.
- To request immediate indexing, open:
  https://pkg.go.dev/github.com/zenmanage/zenmanage-go

### Go Proxy and Checksum DB

- proxy.golang.org and sum.golang.org fetch tags automatically when first requested.
- Verify install flow from a clean environment:

~~~bash
go clean -modcache
go list -m -versions github.com/zenmanage/zenmanage-go
go get github.com/zenmanage/zenmanage-go@v1.0.0
~~~

## 5. Backward Compatibility Policy

1. Keep non-breaking additions in minor releases.
2. Use major version suffixes for breaking changes.

Examples:

- v1.x.y uses module path github.com/zenmanage/zenmanage-go
- v2.x.y uses module path github.com/zenmanage/zenmanage-go/v2

## 6. Optional Additional Distribution

1. Homebrew formula for CLI tools (only if CLI binaries are added later).
2. Internal mirrors such as Artifactory or Athens for enterprise environments.
3. Vanity import path with go-import meta tags, if branding requires it.

## 7. Release Checklist

1. Tests green on supported Go versions.
2. Coverage reviewed and no critical gaps.
3. README examples compile and match public APIs.
4. Changelog updated.
5. Version tag pushed.
6. GitHub release published.
7. pkg.go.dev page indexed and rendering docs.
8. Announcement posted (README badge updates, release notes, team comms).
