package examples

import (
	"context"

	"github.com/zenmanage/zenmanage-go"
)

// Defaults demonstrates inline and collection defaults.
func Defaults(token string) (bool, bool, error) {
	client, err := newClient(token)
	if err != nil {
		return false, false, err
	}

	inlineFlag, err := client.Flags().Single(context.Background(), "inline-default-flag", true)
	if err != nil {
		return false, false, err
	}

	defaults := zenmanage.DefaultsFromMap(map[string]any{"collection-default-flag": true})
	collectionFlag, err := client.Flags().WithDefaults(defaults).Single(context.Background(), "collection-default-flag")
	if err != nil {
		return false, false, err
	}

	return inlineFlag.IsEnabled(), collectionFlag.IsEnabled(), nil
}
