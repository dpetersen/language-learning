package audio

import (
	"context"
	"log"
	"math/rand"
	"strings"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"github.com/dpetersen/language-learning/gpt"
	"github.com/sirupsen/logrus"
)

const speakingRate = 0.8

var voiceNames = []string{
	"es-US-Studio-B",
	"es-US-Neural2-A",
	"es-US-Neural2-B",
	"es-US-Neural2-C",
}

type AudioClient struct{}

func storyToSSML(story gpt.Story) string {
	var result strings.Builder

	result.WriteString("<speak>")
	result.WriteString("<p>")
	result.WriteString(story.Title)
	result.WriteString("</p>")
	result.WriteString(`<break time="2s"/>`)
	logrus.WithField("paragraphs", len(strings.Split(story.Story, "\n"))).Debug("How many paragraphs?")
	for _, paragraph := range strings.Split(story.Story, "\n") {
		result.WriteString("<p>")
		result.WriteString(paragraph)
		result.WriteString("</p>")
	}
	result.WriteString(`<break time="3s"/>`)
	result.WriteString("<p>")
	result.WriteString("Preguntas:")
	result.WriteString("</p>")
	result.WriteString(`<break time="1s"/>`)
	for _, question := range story.Questions {
		result.WriteString("<p>")
		result.WriteString(question.Question)
		result.WriteString("</p>")
		result.WriteString(`<break time="3s"/>`)
		result.WriteString("<p>")
		result.WriteString(question.Answer)
		result.WriteString("</p>")
		result.WriteString(`<break time="1s"/>`)
	}

	result.WriteString("</speak>")

	logrus.WithField("ssml", result.String()).Debug("Generated SSML")

	return result.String()
}

func NewAudioClient() *AudioClient {
	return &AudioClient{}
}

func (c *AudioClient) TextToSpeech(story gpt.Story) ([]byte, error) {
	ctx := context.Background()

	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	resp, err := client.SynthesizeSpeech(
		ctx,
		&texttospeechpb.SynthesizeSpeechRequest{
			Input: &texttospeechpb.SynthesisInput{
				InputSource: &texttospeechpb.SynthesisInput_Ssml{Ssml: storyToSSML(story)},
			},
			Voice: &texttospeechpb.VoiceSelectionParams{
				Name:         randomVoiceName(),
				LanguageCode: "es-US",
			},
			AudioConfig: &texttospeechpb.AudioConfig{
				SpeakingRate:  speakingRate,
				AudioEncoding: texttospeechpb.AudioEncoding_MP3,
			},
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	return resp.AudioContent, nil
}

func randomVoiceName() string {
	return voiceNames[rand.Intn(len(voiceNames))]
}
