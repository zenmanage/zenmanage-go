// Package examples contains runnable samples demonstrating the Zenmanage
// Go SDK, mirroring the sample set in the JavaScript and PHP SDKs. Each
// file exposes a function that can be adapted into application code.
package examples

import (
	"time"

	"github.com/zenmanage/zenmanage-go"
)

func newClient(token string) (*zenmanage.Zenmanage, error) {
	cfg, err := zenmanage.NewConfigBuilder().
		WithEnvironmentToken(token).
		WithCacheTTL(time.Hour).
		Build()
	if err != nil {
		return nil, err
	}
	return zenmanage.New(cfg), nil
}
