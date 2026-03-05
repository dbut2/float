package mastercard

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"time"

	oauth "github.com/mastercard/oauth1-signer-go"
	"github.com/mastercard/oauth1-signer-go/utils"
)

// FXClient is a Mastercard API client with OAuth 1.0a request signing.
type FXClient struct {
	client *ClientWithResponses
}

// NewFXClientFromP12 loads an RSA signing key from a PKCS#12 file.
// Set sandbox=true when using a Mastercard sandbox key.
func NewFXClientFromP12(consumerKey, p12Path, p12Password string, sandbox bool) (*FXClient, error) {
	if consumerKey == "" || p12Path == "" {
		return nil, fmt.Errorf("consumer key and key path are required")
	}
	key, err := utils.LoadSigningKey(p12Path, p12Password)
	if err != nil {
		return nil, fmt.Errorf("loading p12 key: %w", err)
	}
	return newFXClient(consumerKey, key, sandbox)
}

// NewFXClientFromPEM parses a PEM-encoded RSA private key.
// Set sandbox=true when using a Mastercard sandbox key.
func NewFXClientFromPEM(consumerKey, privateKeyPEM string, sandbox bool) (*FXClient, error) {
	if consumerKey == "" || privateKeyPEM == "" {
		return nil, fmt.Errorf("consumer key and private key are required")
	}
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("private key is not valid PEM")
	}
	var rsaKey *rsa.PrivateKey
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		var ok bool
		rsaKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("private key is not RSA")
		}
	} else if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		rsaKey = key
	} else {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	return newFXClient(consumerKey, rsaKey, sandbox)
}

func newFXClient(consumerKey string, key *rsa.PrivateKey, sandbox bool) (*FXClient, error) {
	server := ServerUrlProductionServer
	if sandbox {
		server = ServerUrlSandboxServer
	}
	c, err := NewClientWithResponses(server, WithRequestEditorFn(oauthSigner(consumerKey, key)))
	if err != nil {
		return nil, err
	}
	return &FXClient{client: c}, nil
}

func oauthSigner(consumerKey string, key *rsa.PrivateKey) RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		authHeader, err := oauth.GetAuthorizationHeader(req.URL, req.Method, nil, consumerKey, key)
		if err != nil {
			return err
		}
		req.Header.Set(oauth.AuthorizationHeaderName, authHeader)
		return nil
	}
}

// GetConversionRate returns how many units of quote currency equal 1 unit of base currency.
// e.g. GetConversionRate(ctx, "AUD", "CNY", today) → 4.72 means 1 AUD = 4.72 CNY.
//
// Mastercard models this as trans_curr (foreign/quote) being billed to crdhld_bill_curr (home/base).
// The returned rate is quote→base, so we invert it to get base→quote.
func (c *FXClient) GetConversionRate(ctx context.Context, base, quote string, date time.Time) (float64, error) {
	resp, err := c.client.GetEnhancedConversionDetailsUsingGETWithResponse(ctx, &GetEnhancedConversionDetailsUsingGETParams{
		RateDate:       date.Format("2006-01-02"),
		TransCurr:      quote, // foreign currency
		TransAmt:       1,
		CrdhldBillCurr: base, // home card currency
	})
	if err != nil {
		return 0, err
	}
	if resp.JSON200 == nil {
		return 0, fmt.Errorf("Mastercard API status %d: %s", resp.HTTPResponse.StatusCode, resp.Body) //nolint:staticcheck
	}
	mc := resp.JSON200.Data.Mastercard
	if mc == nil || mc.MastercardConvRateExclAllFees == nil {
		return 0, fmt.Errorf("no conversion rate in Mastercard response")
	}
	rate := float64(*mc.MastercardConvRateExclAllFees)
	if rate == 0 {
		return 0, fmt.Errorf("received zero conversion rate from Mastercard")
	}
	// rate is quote→base (e.g. CNY→AUD), invert to get base→quote (AUD→CNY)
	return 1 / rate, nil
}
