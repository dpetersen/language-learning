package lingq

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/go-resty/resty/v2"
)

// This API is totally undocumented. The v3 API in particular has very little
// info around it, but you can use developer tools in Chrome to see what
// requests it makes. There is some cobbling together in the forums here, for
// useful things like authentication:
// https://forum.lingq.com/t/url-and-docs-for-30-api/75276/4

type Word struct {
	Term   string
	Status int
}

type VocabularyClient struct {
	apiKey string
	client *resty.Client
}

type APIResponse struct {
	Count   int
	Next    *string
	Results []struct {
		Term   string
		Status int
	}
}

// Get your API key at:
// https://www.lingq.com/en/accounts/apikey/
const (
	apiRoot   = "https://www.lingq.com/api/"
	v3Path    = apiRoot + "v3"
	cardsPath = v3Path + "/es/cards/"
)

func NewVocabularyClient(apiKey string) *VocabularyClient {
	return &VocabularyClient{
		apiKey: apiKey,
		client: resty.New(),
	}
}

func (c *VocabularyClient) GetNonNewWords() ([]Word, error) {
	var words []Word
	next := strPtr(cardsPath + "?page=1&page_size=200&sort=alpha&status=2&status=3&status=4")

	for next != nil {
		resp, err := c.newAPIRequest().Get(*next)
		if err != nil {
			return nil, fmt.Errorf("making HTTP request: %v", err)
		}

		var apiResponse APIResponse
		if err := json.Unmarshal(resp.Body(), &apiResponse); err != nil {
			return nil, fmt.Errorf("deserializing API response: %v", err)
		}

		for _, result := range apiResponse.Results {
			words = append(words, Word{Term: result.Term, Status: result.Status})
		}

		next = apiResponse.Next
		log.Printf("Next is now: %v", strFromPtr(next))
	}

	return words, nil
}

func (c *VocabularyClient) newAPIRequest() *resty.Request {
	return c.client.R().
		SetHeader("Authorization", "Token "+c.apiKey).
		SetHeader("accept", "application/json")
}

func strPtr(s string) *string {
	return &s
}

func strFromPtr(s *string) string {
	if s != nil {
		return *s
	}

	return ""
}
