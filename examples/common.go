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
