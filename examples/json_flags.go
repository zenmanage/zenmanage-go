package examples

import "context"

// JSONFlags demonstrates evaluating a structured (JSON object/array) flag.
func JSONFlags(token string) (map[string]any, error) {
	client, err := newClient(token)
	if err != nil {
		return nil, err
	}
	flag, err := client.Flags().Single(context.Background(), "theme-config", map[string]any{
		"mode": "light",
	})
	if err != nil {
		return nil, err
	}
	theme, _ := flag.AsJSON().(map[string]any)
	return theme, nil
}
