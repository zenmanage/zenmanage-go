package zenmanage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// APIClient handles network communication with Zenmanage APIs.
type APIClient struct {
	httpClient  *http.Client
	envToken    string
	apiEndpoint string
	clientAgent string
	sdkVersion  string
	enableUsage bool
	logger      Logger
}

// NewAPIClient creates an API client.
func NewAPIClient(cfg Config) *APIClient {
	return &APIClient{
		httpClient:  cfg.HTTPClient,
		envToken:    cfg.EnvironmentToken,
		apiEndpoint: strings.TrimRight(cfg.APIEndpoint, "/"),
		clientAgent: cfg.ClientAgent,
		sdkVersion:  cfg.SDKVersion,
		enableUsage: cfg.EnableUsageReporting,
		logger:      cfg.Logger,
	}
}

type metadataResponse struct {
	Data struct {
		CDN  string `json:"cdn"`
		Path string `json:"path"`
	} `json:"data"`
}

// FetchRules fetches the full rules payload from API/CDN.
func (c *APIClient) FetchRules(ctx context.Context) (RulesResponse, error) {
	metaURL := c.apiEndpoint + "/v1/flag-json"
	body, status, err := c.doRequestWithRetry(ctx, http.MethodGet, metaURL, nil, nil, true)
	if err != nil {
		return RulesResponse{}, err
	}
	if status < 200 || status >= 300 {
		return RulesResponse{}, &FetchRulesError{Message: "failed to fetch metadata", StatusCode: status}
	}

	var meta metadataResponse
	if err := json.Unmarshal(body, &meta); err != nil {
		return RulesResponse{}, &InvalidRulesError{Message: "invalid metadata response"}
	}
	if meta.Data.CDN == "" || meta.Data.Path == "" {
		return RulesResponse{}, &InvalidRulesError{Message: "metadata missing cdn/path"}
	}

	rulesURL := strings.TrimRight(meta.Data.CDN, "/") + "/" + strings.TrimLeft(meta.Data.Path, "/")
	rulesBody, status, err := c.doRequestWithRetry(ctx, http.MethodGet, rulesURL, nil, nil, false)
	if err != nil {
		return RulesResponse{}, err
	}
	if status < 200 || status >= 300 {
		return RulesResponse{}, &FetchRulesError{Message: "failed to fetch rules", StatusCode: status}
	}

	var rules RulesResponse
	if err := json.Unmarshal(rulesBody, &rules); err != nil {
		return RulesResponse{}, &InvalidRulesError{Message: "invalid rules response"}
	}
	if rules.Version == "" {
		return RulesResponse{}, &InvalidRulesError{Message: "rules missing version"}
	}
	return rules, nil
}

// ReportUsage emits usage telemetry for a flag key. defaultValue, when
// non-nil, is the fallback value the caller used because the flag couldn't
// be resolved from rules; it's sent as the X-Default-Value header so it can
// be persisted and shown on the flag detail page.
func (c *APIClient) ReportUsage(ctx context.Context, key string, contextData *Context, defaultValue any) error {
	if !c.enableUsage {
		return nil
	}
	url := fmt.Sprintf("%s/v1/flags/%s/usage", c.apiEndpoint, key)

	headers := map[string]string{}
	if contextData != nil && !contextData.IsEmpty() {
		payload, err := contextData.JSON()
		if err == nil {
			headers["X-ZENMANAGE-CONTEXT"] = string(payload)
		}
	}
	if defaultValue != nil {
		payload, err := json.Marshal(map[string]any{key: defaultValue})
		if err == nil {
			headers["X-Default-Value"] = string(payload)
		}
	}

	_, status, err := c.doRequestWithRetry(ctx, http.MethodPost, url, []byte("{}"), headers, true)
	if err != nil {
		return err
	}
	if status < 200 || status >= 300 {
		return &FetchRulesError{Message: "failed to report usage", StatusCode: status}
	}
	return nil
}

func (c *APIClient) doRequestWithRetry(ctx context.Context, method, url string, body []byte, headers map[string]string, withAuth bool) ([]byte, int, error) {
	var lastErr error
	var lastStatus int

	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(100*(1<<(attempt-1))) * time.Millisecond
			select {
			case <-ctx.Done():
				return nil, 0, ctx.Err()
			case <-time.After(backoff):
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
		if err != nil {
			return nil, 0, err
		}
		req.Header.Set("Accept", "application/json")
		if method == http.MethodPost {
			req.Header.Set("Content-Type", "application/json")
		}
		if withAuth {
			req.Header.Set("X-API-Key", c.envToken)
			req.Header.Set("X-ZEN-CLIENT-AGENT", fmt.Sprintf("%s/%s", c.clientAgent, c.sdkVersion))
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = &FetchRulesError{Message: "request failed: " + err.Error()}
			continue
		}

		respBody, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			lastErr = &FetchRulesError{Message: "failed reading response"}
			continue
		}

		lastStatus = resp.StatusCode
		if resp.StatusCode >= 500 {
			lastErr = &FetchRulesError{Message: "server error", StatusCode: resp.StatusCode}
			continue
		}
		return respBody, resp.StatusCode, nil
	}

	if lastErr != nil {
		return nil, lastStatus, lastErr
	}
	return nil, lastStatus, &FetchRulesError{Message: "request failed", StatusCode: lastStatus}
}
