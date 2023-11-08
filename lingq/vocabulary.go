package lingq

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// This API is totally undocumented. The v3 API in particular has very little
// info around it, but you can use developer tools in Chrome to see what
// requests it makes. There is some cobbling together in the forums here, for
// useful things like authentication:
// https://forum.lingq.com/t/url-and-docs-for-30-api/75276/4

const MaxWordStatus = 5

var WordStatusMeanings = map[int]string{
	1: "New: not learned",
	2: "Recognized: not learned but I have seen this before",
	3: "Familiar: I usually remember seeing this word and often recall its meaning, though sometimes I forget",
	4: "Learned: I know this word and almost always recognize it",
	5: "Known: I have mastered this word and can use it in a sentence",
}

type Word struct {
	Term   string
	Status int
}

type Client struct {
	apiKey string
	client *resty.Client
}

type APIResult struct {
	Term           string
	Status         int
	ExtendedStatus *int `json:"extended_status"`
}

type APIResponse struct {
	Count   int
	Next    *string
	Results []APIResult
}

// Get your API key at:
// https://www.lingq.com/en/accounts/apikey/
const (
	apiRoot   = "https://www.lingq.com/api/"
	v3Path    = apiRoot + "v3"
	cardsPath = v3Path + "/es/cards/"
)

func (c *Client) GetNonNewWords() ([]Word, error) {
	var words []Word
	next := strPtr(cardsPath + "?page=1&page_size=200&sort=alpha&status=2&status=3&status=4&status=5")

	logrus.Debug("fetching initial page from Lingq API")
	for next != nil {
		var apiResponse APIResponse
		resp, err := c.newAPIRequest().SetResult(apiResponse).Get(*next)
		if err != nil {
			return nil, fmt.Errorf("making HTTP request: %v", err)
		}

		logrus.WithField("response", string(resp.Body())).Debug("Got response from Lingq API")
		for _, result := range apiResponse.Results {
			words = append(
				words,
				Word{Term: result.Term, Status: actualStatusFromInsaneStatus(result)},
			)
		}

		next = apiResponse.Next
		logrus.WithField("next", strFromPtr(next)).Debug("Got next page")
	}

	return words, nil
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

// Apparently LingQ's status field only goes from 1 to 3? And then there's an
// extended_status field has to be added to the status in some cases? I found
// somebody else who'd parsed this insanity. Here are the rules:
//
// 1: status: 0, extended_status: null
// 2: status: 1, extended_status: 0
// 3: status: 2, extended_status: 0
// 4: status: 3, extended_status: 0 || 1 (that's right, some of them are 0 and some are 1)
// 5: status: 3, extended_status: 3 || null (yup, in a handful of cases the extended_status is null)
//
// I cannot imagine what somebody was thinking with this.
//
// I found somebody else who dealt with this insanity, too, so might need to
// look here:
// https://github.com/thags/lingqAnkiSync/blob/main/LingqAnkiSync/LingqApi.py#L64C23-L64C23
func actualStatusFromInsaneStatus(result APIResult) int {
	if result.Status == 0 {
		return 1
	} else if result.Status == 1 {
		return 2
	} else if result.Status == 2 {
		return 3
	} else if result.Status == 3 && result.ExtendedStatus != nil && *result.ExtendedStatus < 2 {
		return 4
	} else if (result.Status == 3 && result.ExtendedStatus != nil && *result.ExtendedStatus == 3) || (result.Status == 3 && result.ExtendedStatus == nil) {
		return 5
	} else {
		logrus.WithField("result", result).Warn("Unknown status")
		return 0
	}
}
