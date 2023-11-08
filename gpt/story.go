package gpt

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/dpetersen/language-learning/lingq"
	"github.com/spf13/viper"
)

const (
	completionsAPI     = "https://api.openai.com/v1/chat/completions"
	formatInstructions = `
After each story, ask the student 5 questions in Spanish about the story. The
point is to reinforce the vocabulary from the story.

I want the response in the form of a valid JSON object. Here is an example:

{
	"title": "Juan's Trip to France",
	"description": "Juan takes a trip to France and learns the true meaning of friendship.",
	"story": "Once upon a time there was a boy named Juan. He wanted to travel to France. He thought it was a beautiful country.\nHe had a friend named Maria. She wanted to travel to France too. They decided to travel to France together. They had a great time. They learned a lot about French culture. They learned a lot about each other.\nThey became best friends. The end.",
	"questions": [
		{
			"question": "Where does Juan want to travel to?",
			"answer": "Juan wants to travel to France. He thinks it is a beautiful country."
		},
		{
			"question": "Is Maria Juan's sister?",
			"answer": "No, Maria is Juan's friend."
		}
	]
}
Here is a vocabulary list for the student:
`
)

type completionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type completionRequest struct {
	Model          string              `json:"model"`
	Messages       []completionMessage `json:"messages"`
	MaxTokens      int                 `json:"max_tokens"`
	N              int                 `json:"n"`
	Temperature    float64             `json:"temperature"`
	User           string              `json:"user"`
	ResponseFormat struct {
		Type string `json:"type"`
	} `json:"response_format"`
}

type completionResponse struct {
	Choices []struct {
		Message struct {
			Content string
		}
		FinishReason string `json:"finish_reason"`
	}
}

func (c *Client) LoadStory(path string) (*Story, error) {
	s, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %v", err)
	}

	story, err := contentJSONToStory(string(s))
	if err != nil {
		return nil, fmt.Errorf("decoding story: %v", err)
	}

	return story, nil
}

func (c *Client) CreateStory(words []lingq.Word, threshold int) (*Story, error) {
	requestObject := completionRequest{
		Model: c.model,
		Messages: []completionMessage{
			{
				Role: "system",
				Content: viper.GetString("openai.story_instructions") +
					"\n\n" +
					fmt.Sprintf("Please make the story in the neighborhood of %d words.\n\n", viper.GetInt("openai.story_length")) +
					formatInstructions +
					wordsByStatus(words, threshold),
			},
			{
				Role:    "user",
				Content: viper.GetString("openai.story_prompt"),
			},
		},
		// TODO could count the length of the prompt and do this intelligently,
		// instead of just adding 500
		MaxTokens:   viper.GetInt("openai.story_length") + 500,
		N:           1,
		Temperature: 0.7,
		User:        apiUserName,
	}
	requestObject.ResponseFormat.Type = "json_object"

	var responseObject completionResponse
	if err := c.makeAPICall(requestObject, completionsAPI, &responseObject); err != nil {
		return nil, fmt.Errorf("calling completions API: %v", err)
	}

	if len(responseObject.Choices) == 0 {
		return nil, errors.New("no choices in response")
	}

	if responseObject.Choices[0].FinishReason != "stop" {
		return nil, fmt.Errorf("unexpected finish reason: %v", responseObject.Choices[0].FinishReason)
	}

	story, err := contentJSONToStory(responseObject.Choices[0].Message.Content)
	if err != nil {
		return nil, fmt.Errorf("decoding story: %v", err)
	}

	return story, nil
}

func contentJSONToStory(s string) (*Story, error) {
	var story Story
	if err := json.Unmarshal([]byte(s), &story); err != nil {
		return nil, fmt.Errorf("decoding JSON: %v", err)
	}
	story.OriginalJSON = s

	return &story, nil
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
