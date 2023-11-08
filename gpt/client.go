package gpt

import (
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const apiUserName = "Language Learning"

type Client struct {
	client *resty.Client
	model  string
	apiKey string
}

func NewClient(apiKey, model string) *Client {
	return &Client{
		client: resty.New().SetDebug(viper.GetBool("openai.http_debug")),
		model:  model,
		apiKey: apiKey,
	}
}

func (c *Client) makeAPICall(requestObject interface{}, url string, responseObject interface{}) error {
	requestBody, err := json.Marshal(requestObject)
	if err != nil {
		return fmt.Errorf("serializing request to JSON: %v", err)
	}

	logrus.WithField("requestBody", string(requestBody)).Debug("Sending request to OpenAI API")
	response, err := c.client.R().
		SetAuthToken(c.apiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(requestBody).
		SetResult(responseObject).
		Post(url)
	if err != nil {
		return fmt.Errorf("making HTTP request: %v", err)
	}

	logrus.WithField("response", string(response.Body())).Debug("Got response from OpenAI API")

	return nil
}
