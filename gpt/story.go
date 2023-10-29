package gpt

import (
	"encoding/json"
	"fmt"

	"github.com/dpetersen/language-learning/lingq"
	"github.com/go-resty/resty/v2"
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

func (c *StoryClient) Create(words []lingq.Word, threshold int) (string, error) {
	var promptWords string
	for _, word := range words {
		if word.Status >= threshold {
			promptWords += fmt.Sprintf("%s,%d\n", word.Term, word.Status)
		}
	}
	requestObject := completionRequest{
		Model: c.model,
		Messages: []completionMessage{
			{
				Role:    "system",
				Content: "You are a Spanish tutor and you are teaching a student how to write a story in Spanish. The student is writing a story in Spanish using the following words with their familiarity levels. A higher number means the word is better known. Make sure the resulting story that it is roughly 95% comprehensible based on my known words:\n" + promptWords,
			},
			{
				Role:    "user",
				Content: "Please write me a story that I can understand. I am a beginner.",
			},
		},
		MaxTokens:   300,
		N:           1,
		Temperature: 0.7,
		User:        "Language Learning",
	}

	requestBody, err := json.Marshal(requestObject)
	if err != nil {
		return "", fmt.Errorf("serializing request to JSON: %v", err)
	}

	fmt.Printf("Sending data: %+v\n", string(requestBody))
	response, err := c.client.R().
		SetAuthToken(c.apiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(requestBody).
		Post(completionsAPI)
	if err != nil {
		return "", fmt.Errorf("making HTTP request: %v", err)
	}

	fmt.Printf("Got data: %s\n", response.Body())

	// var data map[string]interface{}
	// if err = json.Unmarshal(response.Body(), &data); err != nil {
	//   return "", fmt.Errorf("decoding JSON response: %v", err)
	// }

	// content := data["choices"].([]interface{})[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)

	return "", nil
}
