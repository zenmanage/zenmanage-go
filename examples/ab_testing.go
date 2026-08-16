package examples

import (
	"context"

	"github.com/zenmanage/zenmanage-go"
)

// ABTesting demonstrates variant selection via string flags.
func ABTesting(token, userID, country string) (string, error) {
	client, err := newClient(token)
	if err != nil {
		return "", err
	}

	ctx := zenmanage.NewContext("user", userID, "", []zenmanage.Attribute{
		zenmanage.NewAttribute("country", []string{country}),
	})

	variant, err := client.Flags().WithContext(ctx).Single(context.Background(), "checkout-flow", "control")
	if err != nil {
		return "", err
	}
	return variant.AsString(), nil
}
