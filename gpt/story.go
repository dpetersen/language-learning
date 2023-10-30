package gpt

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/dpetersen/language-learning/lingq"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

const (
	completionsAPI = "https://api.openai.com/v1/chat/completions"
)

type StoryClient struct {
	client *resty.Client
	model  string
	apiKey string
}

func NewStoryClient(apiKey, model string) *StoryClient {
	return &StoryClient{
		client: resty.New(),
		model:  model,
		apiKey: apiKey,
	}
}

type completionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type completionRequest struct {
	Model       string              `json:"model"`
	Messages    []completionMessage `json:"messages"`
	MaxTokens   int                 `json:"max_tokens"`
	N           int                 `json:"n"`
	Temperature float64             `json:"temperature"`
	User        string              `json:"user"`
}

type completionResponse struct {
	Choices []struct {
		Message struct {
			Content string
		}
		FinishReason string `json:"finish_reason"`
	}
}

func (c *StoryClient) Create(words []lingq.Word, threshold int) (string, error) {
	requestObject := completionRequest{
		Model: c.model,
		Messages: []completionMessage{
			{
				Role: "system",
				Content: `
				You are a Spanish tutor who teaches by telling stories using the theory of Comprehensible Input. You believe the student learns best when they understand 95% of the words they read or hear. Keep the story around 500-700 words.

				After each story, ask the student 5 questions in Spanish about the story afterwards, but giving them the answer as part of the question (e.g. Juan wants to travel to France. Where does Juan want to travel?)

				Here is a vocabulary list for the student:\n` + wordsByStatus(words, threshold),
			},
			{
				Role:    "user",
				Content: "Please write me a story that I can understand. I am a beginner.",
			},
		},
		MaxTokens:   1000,
		N:           1,
		Temperature: 0.7,
		User:        "Language Learning",
	}

	requestBody, err := json.Marshal(requestObject)
	if err != nil {
		return "", fmt.Errorf("serializing request to JSON: %v", err)
	}

	logrus.WithField("requestBody", string(requestBody)).Debug("Sending request to OpenAI API")
	response, err := c.client.R().
		SetAuthToken(c.apiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(requestBody).
		Post(completionsAPI)
	if err != nil {
		return "", fmt.Errorf("making HTTP request: %v", err)
	}

	logrus.WithField("response", string(response.Body())).Debug("Got response from OpenAI API")

	var responseObject completionResponse
	if err = json.Unmarshal(response.Body(), &responseObject); err != nil {
		return "", fmt.Errorf("decoding JSON response: %v", err)
	}

	if len(responseObject.Choices) == 0 {
		return "", errors.New("no choices in response")
	}

	if responseObject.Choices[0].FinishReason != "stop" {
		return "", fmt.Errorf("unexpected finish reason: %v", responseObject.Choices[0].FinishReason)
	}

	return responseObject.Choices[0].Message.Content, nil
}

func wordsByStatus(words []lingq.Word, threshold int) string {
	statusMap := make(map[int][]string)

	for _, word := range words {
		if word.Status >= threshold {
			statusMap[word.Status] = append(statusMap[word.Status], word.Term)
		}
	}

	var result strings.Builder
	for level := threshold; level <= lingq.MaxWordStatus; level++ {
		if terms, exists := statusMap[level]; exists {
			result.WriteString(lingq.WordStatusMeanings[level] + "\n")
			result.WriteString(strings.Join(terms, ","))
			result.WriteString("\n\n")
		}
	}

	return result.String()
}
