# Zenmanage Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/zenmanage/zenmanage-go.svg)](https://pkg.go.dev/github.com/zenmanage/zenmanage-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/zenmanage/zenmanage-go)](https://goreportcard.com/report/github.com/zenmanage/zenmanage-go)

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

- Go 1.22+
- Server token prefixed with srv_

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

## Testing

~~~bash
go test ./... -coverprofile=coverage.out
~~~

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Publishing

See [docs/PUBLISHING_NEXT_STEPS.md](docs/PUBLISHING_NEXT_STEPS.md) for a complete release and publication checklist.

## License

MIT. See [LICENSE](LICENSE).
