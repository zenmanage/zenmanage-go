# Zenmanage Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/zenmanage/zenmanage-go.svg)](https://pkg.go.dev/github.com/zenmanage/zenmanage-go)
[![CI](https://github.com/zenmanage/zenmanage-go/actions/workflows/ci.yml/badge.svg)](https://github.com/zenmanage/zenmanage-go/actions/workflows/ci.yml)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/e45d035523964569bb95d00e5e2c0a82)](https://app.codacy.com/gh/zenmanage/zenmanage-go/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade)

Add feature flags to your Go services in minutes. Control feature rollouts, run A/B tests, and manage configuration without redeploying.

## Why Zenmanage?

- Fast local evaluation with cached rules
- Context-aware targeting by user, org, or any attributes
- Deterministic percentage rollouts via CRC32B bucketing
- Safe defaults and defensive error handling
- Testable interfaces and high unit test coverage

## Installation

~~~bash
go get github.com/zenmanage/zenmanage-go
~~~

Requirements:

- Go 1.25+
- Server token prefixed with srv_

## Key Compatibility

- Server runtime only: environment tokens prefixed with `srv_`.
- Client keys (`cli_`) and mobile keys (`mob_`) are rejected by this SDK at
  configuration time via `ConfigurationError` — this is a server-side SDK,
  matching the PHP core SDK's key requirements. Use `zenmanage-javascript` (or
  another browser/mobile SDK) for client- or mobile-key runtimes.

## Quick Start

~~~go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/zenmanage/zenmanage-go"
)

func main() {
    cfg, err := zenmanage.NewConfigBuilder().
        WithEnvironmentToken("srv_your_server_key_here").
        Build()
    if err != nil {
        log.Fatal(err)
    }

    client := zenmanage.New(cfg)

    // Simple one-liner — evaluate a flag for a user ID.
    enabled, err := client.IsEnabled(context.Background(), "new-dashboard", "user-123")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("enabled:", enabled)
}
~~~

## Common Use Cases

### Simple API (single-call evaluation)

~~~go
// Boolean flag
enabled, err := client.IsEnabled(ctx, "beta-feature", "user-123")

// String flag
color, err := client.GetString(ctx, "button-color", "user-123", "blue")

// Numeric flag
limit, err := client.GetNumber(ctx, "rate-limit", "user-123", 100)
~~~

Pass an empty string for `userID` to evaluate without user context (e.g. kill-switch flags).

### Advanced API — fluent FlagManager

~~~go
flag, err := client.Flags().
    WithContext(ctx).
    Single(context.Background(), "new-dashboard", false)
if err != nil {
    log.Fatal(err)
}
fmt.Println("enabled:", flag.IsEnabled())
~~~

### Context-Based Targeting

~~~go
ctx := zenmanage.NewContext(
    "user",
    "user-123",
    "Taylor",
    []zenmanage.Attribute{
        zenmanage.NewAttribute("country", []string{"US"}),
        zenmanage.NewAttribute("plan", []string{"pro"}),
    },
)

flag, err := client.Flags().
    WithContext(ctx).
    Single(context.Background(), "beta-program", false)
~~~

### Percentage Rollouts

~~~go
ctx := zenmanage.SingleContext("user", "user-123", "")
flag, err := client.Flags().
    WithContext(ctx).
    Single(context.Background(), "new-checkout-flow", false)
~~~

### Defaults Collection

~~~go
defaults := zenmanage.DefaultsFromMap(map[string]any{
    "new-ui": true,
    "api-version": "v2",
})

flag, err := client.Flags().
    WithDefaults(defaults).
    Single(context.Background(), "new-ui")
~~~

### Fetch All Flags

~~~go
flags, err := client.Flags().All(context.Background())
if err != nil {
    log.Fatal(err)
}
for _, flag := range flags {
    fmt.Println(flag.Key(), flag.Value())
}
~~~

`All` evaluates every flag in the environment against the manager's current
context in one call. Unlike `Single`, it does not report per-flag usage —
usage reporting is a signal tied to a specific evaluation decision, not a bulk
retrieval.

### Manual Usage Reporting

`Single` reports usage automatically on every evaluation. Call `ReportUsage`
directly only when you need to record usage outside of a `Single`/`All`
evaluation path (for example, after evaluating a flag some other way):

~~~go
err := client.Flags().ReportUsage(context.Background(), "new-dashboard", false)
~~~

## Configuration

Build configuration with fluent helpers:

- WithEnvironmentToken
- WithCacheTTL
- WithCacheBackend: memory, filesystem, null
- WithCacheDirectory
- WithUsageReporting
- WithAPIEndpoint
- WithLogger
- WithCache
- WithHTTPClient

Or load from environment with ConfigFromEnvironment using:

- ZENMANAGE_ENVIRONMENT_TOKEN
- ZENMANAGE_CACHE_TTL
- ZENMANAGE_CACHE_BACKEND
- ZENMANAGE_CACHE_DIR
- ZENMANAGE_ENABLE_USAGE_REPORTING
- ZENMANAGE_API_ENDPOINT

## Examples

See [examples/README.md](examples/README.md) for parity samples:

- simple_flags.go
- context_based_flags.go
- percentage_rollouts.go
- ab_testing.go
- caching.go
- defaults.go
- middleware.go

## Middleware

The `middleware` sub-package provides framework integrations so flag evaluation
fits naturally into HTTP request handling.

### net/http (standard library)

~~~go
import "github.com/zenmanage/zenmanage-go/middleware"

mux := http.NewServeMux()
mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    enabled, err := middleware.IsEnabled(r.Context(), "new-feature")
    ...
})
http.ListenAndServe(":8080", middleware.InjectFlags(zmClient, mux))
~~~

If the request contains an `X-User-ID` header, the middleware automatically
wires up a user context so rollout and targeting rules apply per-request.

### Gin

~~~bash
go get github.com/zenmanage/zenmanage-go/middleware/gin
~~~

~~~go
import ginmw "github.com/zenmanage/zenmanage-go/middleware/gin"

r := gin.Default()
r.Use(ginmw.InjectFlags(zmClient))

r.GET("/feature", func(c *gin.Context) {
    enabled, err := ginmw.IsEnabled(c, "new-feature")
    ...
})
~~~

### Echo

~~~bash
go get github.com/zenmanage/zenmanage-go/middleware/echo
~~~

~~~go
import echomw "github.com/zenmanage/zenmanage-go/middleware/echo"

e := echo.New()
e.Use(echomw.InjectFlags(zmClient))

e.GET("/feature", func(c echo.Context) error {
    enabled, err := echomw.IsEnabled(c, "new-feature")
    ...
})
~~~

## Error Handling

Every error this SDK returns satisfies the shared `zenmanage.Error` interface,
in addition to the standard `error` interface, so you can distinguish SDK
errors from other errors with `errors.As`:

~~~go
import (
    "errors"

    "github.com/zenmanage/zenmanage-go"
)

flag, err := client.Flags().Single(ctx, "unknown-flag")
if err != nil {
    var evalErr *zenmanage.EvaluationError
    var fetchErr *zenmanage.FetchRulesError
    switch {
    case errors.As(err, &evalErr):
        log.Println("flag not found:", evalErr.Message)
    case errors.As(err, &fetchErr):
        log.Println("failed to fetch rules:", fetchErr.Message, fetchErr.StatusCode)
    default:
        var zmErr zenmanage.Error
        if errors.As(err, &zmErr) {
            log.Println("SDK error:", zmErr)
        }
    }
}

// Or pass an inline default to avoid the "not found" error entirely.
flag, err = client.Flags().Single(ctx, "unknown-flag", false)
~~~

The concrete error types are `ConfigurationError` (invalid SDK setup),
`EvaluationError` (rule/flag evaluation failure), `FetchRulesError` (failure
loading rules from the API, with an optional `StatusCode`), and
`InvalidRulesError` (malformed rules payload).

## Testing

~~~bash
go test ./... -coverprofile=coverage.out
~~~

## Linting

CI runs [golangci-lint](https://golangci-lint.run/) against the root module and
each middleware submodule (`.golangci.yml`). Run it locally with:

~~~bash
golangci-lint run ./...
~~~

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT. See [LICENSE](LICENSE).
