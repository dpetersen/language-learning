package gpt

import "strings"

type Story struct {
	Title       string
	Description string
	Story       string
	Questions   []Question

	OriginalJSON string
	Thumbnail    string
}

type Question struct {
	Question string
	Answer   string
}

func (s Story) ToString() string {
	var result strings.Builder

	result.WriteString(s.Title)
	result.WriteString("\n\n")
	for _, paragraph := range strings.Split(s.Story, "\n") {
		result.WriteString(paragraph)
		result.WriteString("\n\n")
	}
	result.WriteString("Preguntas:\n\n")
	for _, question := range s.Questions {
		result.WriteString(question.Question)
		result.WriteString("\n\n")
		result.WriteString(question.Answer)
		result.WriteString("\n\n")
	}

	return result.String()
}
