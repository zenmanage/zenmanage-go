# Contributing

Thanks for your interest in contributing.

## Development Setup

1. Install Go 1.25+
2. Clone the repository
3. Run tests

~~~bash
go test ./... -cover
~~~

4. Run the linter

~~~bash
golangci-lint run ./...
~~~

## Quality Bar

- Keep public APIs documented
- Add tests for all behavior changes
- Preserve backward compatibility unless explicitly changing major versions
- Keep examples and README aligned with actual API behavior

## Pull Requests

- Keep PRs focused and reviewable
- Include rationale in commit messages and PR description
- Update CHANGELOG.md for user-visible changes
