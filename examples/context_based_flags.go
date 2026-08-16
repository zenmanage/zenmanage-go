package examples

import (
	"context"

	"github.com/zenmanage/zenmanage-go"
)

// ContextBasedFlags demonstrates attribute-driven targeting.
func ContextBasedFlags(token, userID, country, plan string) (bool, error) {
	client, err := newClient(token)
	if err != nil {
		return false, err
	}

	ctx := zenmanage.NewContext("user", userID, "", []zenmanage.Attribute{
		zenmanage.NewAttribute("country", []string{country}),
		zenmanage.NewAttribute("plan", []string{plan}),
	})

	flag, err := client.Flags().WithContext(ctx).Single(context.Background(), "pro-dashboard", false)
	if err != nil {
		return false, err
	}
	return flag.IsEnabled(), nil
}
