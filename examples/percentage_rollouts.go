package examples

import (
	"context"

	"github.com/zenmanage/zenmanage-go"
)

// PercentageRollouts demonstrates deterministic rollout behavior.
func PercentageRollouts(token, userID string) (bool, error) {
	client, err := newClient(token)
	if err != nil {
		return false, err
	}

	ctx := zenmanage.SingleContext("user", userID, "")
	flag, err := client.Flags().WithContext(ctx).Single(context.Background(), "new-checkout-flow", false)
	if err != nil {
		return false, err
	}
	return flag.IsEnabled(), nil
}
