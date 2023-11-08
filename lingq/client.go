package lingq

import (
	"os"

	"github.com/go-resty/resty/v2"
)

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		client: resty.New(),
	}
}

func (c *Client) newAPIRequest() *resty.Request {
	if os.Getenv("LINGQ_HTTP_DEBUG") == "true" {
		c.client.SetDebug(true)
	}

	return c.client.R().
		SetHeader("Authorization", "Token "+c.apiKey).
		SetHeader("accept", "application/json")
}
