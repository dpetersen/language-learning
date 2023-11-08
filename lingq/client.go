package lingq

import (
	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
)

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		client: resty.New().SetDebug(viper.GetBool("lingq.http_debug")),
	}
}

func (c *Client) newAPIRequest() *resty.Request {
	return c.client.R().
		SetHeader("Authorization", "Token "+c.apiKey).
		SetHeader("accept", "application/json")
}
