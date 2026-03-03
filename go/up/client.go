package up

import (
	"context"
	"fmt"
	"net/http"
)

type UpClient struct {
	gen *ClientWithResponses
}

func NewUpClient(token string) (*UpClient, error) {
	c, err := NewClientWithResponses(ServerUrlHttpsapiUpComAuapiv1, WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	}))
	if err != nil {
		return nil, err
	}
	return &UpClient{gen: c}, nil
}

func (c *UpClient) Ping(ctx context.Context) error {
	resp, err := c.gen.GetUtilPingWithResponse(ctx)
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("up api returned status %d", resp.StatusCode())
	}
	return nil
}

func (c *UpClient) doGet(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	inner := c.gen.ClientInterface.(*Client)
	for _, editor := range inner.RequestEditors {
		if err := editor(ctx, req); err != nil {
			return nil, err
		}
	}
	return inner.Client.Do(req)
}

func fetchPages[T any](ctx context.Context, c *UpClient, firstPageURL string, parse func(*http.Response) ([]T, *string, error)) ([]T, error) {
	var all []T
	nextURL := &firstPageURL

	for nextURL != nil {
		resp, err := c.doGet(ctx, *nextURL)
		if err != nil {
			return nil, err
		}
		data, next, err := parse(resp)
		if err != nil {
			return nil, err
		}
		all = append(all, data...)
		nextURL = next
	}

	return all, nil
}

func (c *UpClient) ListAccounts(ctx context.Context) ([]AccountResource, error) {
	return fetchPages(ctx, c, ServerUrlHttpsapiUpComAuapiv1+"/accounts", func(resp *http.Response) ([]AccountResource, *string, error) {
		parsed, err := ParseGetAccountsResponse(resp)
		if err != nil {
			return nil, nil, err
		}
		if parsed.JSON200 == nil {
			return nil, nil, fmt.Errorf("up api returned status %d", parsed.StatusCode())
		}
		return parsed.JSON200.Data, parsed.JSON200.Links.Next, nil
	})
}

func (c *UpClient) ListTransactions(ctx context.Context, accountID string) ([]TransactionResource, error) {
	url := ServerUrlHttpsapiUpComAuapiv1 + "/transactions"
	if accountID != "" {
		url = ServerUrlHttpsapiUpComAuapiv1 + "/accounts/" + accountID + "/transactions?page[size]=100"
	}

	return fetchPages(ctx, c, url, func(resp *http.Response) ([]TransactionResource, *string, error) {
		parsed, err := ParseGetTransactionsResponse(resp)
		if err != nil {
			return nil, nil, err
		}
		if parsed.JSON200 == nil {
			return nil, nil, fmt.Errorf("up api returned status %d", parsed.StatusCode())
		}
		return parsed.JSON200.Data, parsed.JSON200.Links.Next, nil
	})
}

func (c *UpClient) GetTransaction(ctx context.Context, id string) (*TransactionResource, error) {
	resp, err := c.gen.GetTransactionsIdWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("up api returned status %d", resp.StatusCode())
	}
	return &resp.JSON200.Data, nil
}

func (c *UpClient) RegisterWebhook(ctx context.Context, url string) (id string, secretKey string, err error) {
	body := PostWebhooksJSONRequestBody{}
	body.Data.Attributes.Url = url

	resp, err := c.gen.PostWebhooksWithResponse(ctx, body)
	if err != nil {
		return "", "", err
	}
	if resp.JSON201 == nil {
		return "", "", fmt.Errorf("up api returned status %d: %s", resp.StatusCode(), string(resp.Body))
	}

	return resp.JSON201.Data.Id, *resp.JSON201.Data.Attributes.SecretKey, nil
}

func (c *UpClient) DeleteWebhook(ctx context.Context, id string) error {
	resp, err := c.gen.DeleteWebhooksIdWithResponse(ctx, id)
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusNoContent {
		return fmt.Errorf("up api returned status %d", resp.StatusCode())
	}
	return nil
}

func (c *UpClient) ListWebhooks(ctx context.Context) ([]WebhookResource, error) {
	return fetchPages(ctx, c, ServerUrlHttpsapiUpComAuapiv1+"/webhooks", func(resp *http.Response) ([]WebhookResource, *string, error) {
		parsed, err := ParseGetWebhooksResponse(resp)
		if err != nil {
			return nil, nil, err
		}
		if parsed.JSON200 == nil {
			return nil, nil, fmt.Errorf("up api returned status %d", parsed.StatusCode())
		}
		return parsed.JSON200.Data, parsed.JSON200.Links.Next, nil
	})
}
