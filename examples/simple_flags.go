package examples

import "context"

// SimpleFlags demonstrates basic single flag evaluation and fallbacks.
func SimpleFlags(token string) (bool, error) {
	client, err := newClient(token)
	if err != nil {
		return false, err
	}
	flag, err := client.Flags().Single(context.Background(), "new-dashboard", false)
	if err != nil {
		return false, err
	}
	return flag.IsEnabled(), nil
}
