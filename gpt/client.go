package gpt

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

const apiUserName = "Language Learning"

type Client struct {
	client *resty.Client
	model  string
	apiKey string
}

func NewClient(apiKey, model string) *Client {
	return &Client{
		client: resty.New(),
		model:  model,
		apiKey: apiKey,
	}
}

func (c *Client) makeAPICall(requestObject interface{}, url string, responseObject interface{}) error {
	if os.Getenv("GPT_HTTP_DEBUG") == "true" {
		c.client.SetDebug(true)
	}

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
