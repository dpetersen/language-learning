package gpt

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

const (
	imageGenerationAPI = "https://api.openai.com/v1/images/generations"
	imagePrompt        = `
Create an eye-catching thumbnail in the style of an Audiobook cover for the story that follows. Match the style and intended audience of the image to that of the story:
`
)

type generationRequest struct {
	Model          string `json:"model"`
	Prompt         string `json:"prompt"`
	Size           string `json:"size"`
	Quality        string `json:"quality"`
	Style          string `json:"style"`
	ResponseFormat string `json:"response_format"`
	User           string `json:"user"`
}

type generationResponse struct {
	Data []struct {
		RevisedPrompt string `json:"revised_prompt"`
		B64JSON       string `json:"b64_json"`
	} `json:"data"`
}

func (c *Client) CreateImage(story string) (string, error) {
	requestObject := generationRequest{
		Model:          "dall-e-3",
		Prompt:         imagePrompt + "\n" + firstN(story, 2000),
		Size:           "1024x1024",
		Quality:        "standard",
		Style:          "vivid",
		ResponseFormat: "b64_json",
		User:           apiUserName,
	}

	var responseObject generationResponse
	if err := c.makeAPICall(requestObject, imageGenerationAPI, &responseObject); err != nil {
		return "", fmt.Errorf("making Image Generation API call: %v", err)
	}

	logrus.WithField("responseObject", responseObject).Debug("Got response from Image Generation API")

	if len(responseObject.Data) != 1 {
		return "", fmt.Errorf("unexpected number of data elements in response: %v", len(responseObject.Data))
	}

	return responseObject.Data[0].B64JSON, nil
}

func firstN(s string, n int) string {
	i := 0
	for j := range s {
		if i == n {
			return s[:j]
		}
		i++
	}
	return s
}
