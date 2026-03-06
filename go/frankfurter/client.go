package frankfurter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type FXClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewFXClient() *FXClient {
	return &FXClient{
		baseURL:    "https://api.frankfurter.dev/v1",
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type response struct {
	Rates map[string]float64 `json:"rates"`
}

func (c *FXClient) GetConversionRate(ctx context.Context, base, quote string, date time.Time) (float64, error) {
	url := fmt.Sprintf("%s/%s?base=%s&symbols=%s", c.baseURL, date.Format("2006-01-02"), base, quote)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("frankfurter API status %d", resp.StatusCode)
	}
	var r response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return 0, err
	}
	rate, ok := r.Rates[quote]
	if !ok || rate == 0 {
		return 0, fmt.Errorf("no rate for %s in frankfurter response", quote)
	}
	return rate, nil
}
