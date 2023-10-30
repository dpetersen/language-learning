package audio

import (
	"context"
	"log"
	"math/rand"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
)

var voiceNames = []string{
	"es-US-Studio-B",
	"es-US-Neural2-A",
	"es-US-Neural2-B",
	"es-US-Neural2-C",
}

type AudioClient struct{}

func NewAudioClient() *AudioClient {
	return &AudioClient{}
}

func (c *AudioClient) TextToSpeech(text string) ([]byte, error) {
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
				InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
			},
			Voice: &texttospeechpb.VoiceSelectionParams{
				Name:         "es-US-Neural2-B",
				LanguageCode: "es-US",
			},
			AudioConfig: &texttospeechpb.AudioConfig{
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
