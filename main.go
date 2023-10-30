package main

import (
	"os"

	"github.com/dpetersen/language-learning/audio"
	"github.com/dpetersen/language-learning/gpt"
	"github.com/dpetersen/language-learning/lingq"
	"github.com/sirupsen/logrus"
)

/*

Todo List:

- syncs vocabulary from LingQ to local
- sends vocabulary and prompt to GPT to build comprehensible input
  - completely invented fictional stories
  - summaries of Wikipedia pages or news articles?
- generates text-to-speech audio from script
- pushes back text and audio to LingQ as new import

*/

type Config struct {
	LingQAPIKey       string
	LingQDatabasePath string
	OpenAIAPIKey      string
	GPTModel          string
}

func init() {
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		if parsed, err := logrus.ParseLevel(level); err == nil {
			logrus.SetLevel(parsed)
		}
	}
}

func main() {
	config := GetVarsOrDieTrying()

	logrus.Info("Checking local database...")
	database := lingq.NewWordDatabase(config.LingQDatabasePath)
	words, err := database.FetchIfFresh()
	if err != nil {
		logrus.WithError(err).Fatal("Fetching LingQ words from database")
	}

	if words == nil {
		logrus.Info("Database not fresh, fetching new words...")
		fetchedWords, err := lingq.NewVocabularyClient(config.LingQAPIKey).GetNonNewWords()
		if err != nil {
			logrus.WithError(err).Fatal("Getting non-new words from LingQ")
		}
		words = fetchedWords

		logrus.Info("Storing fetched words in database...")
		if err := database.Store(words); err != nil {
			logrus.WithError(err).Fatal("Storing LingQ words into database")
		}
	}

	logrus.WithField("count", len(words)).Info("Loaded words")

	storyClient := gpt.NewStoryClient(config.OpenAIAPIKey, config.GPTModel)
	story, err := storyClient.Create(words, 3)
	if err != nil {
		logrus.WithError(err).Fatal("Creating story")
	}

	logrus.WithField("story", story).Info("Got story")

	transcriptFile, err := os.Create("output.txt")
	if err != nil {
		logrus.WithError(err).Fatal("failed to create transcript")
	}
	defer transcriptFile.Close()

	if _, err = transcriptFile.WriteString(story); err != nil {
		logrus.WithError(err).Fatal("Failed to write to file")
	}

	audio, err := audio.NewAudioClient().TextToSpeech(story)
	if err != nil {
		logrus.WithError(err).Fatal("Generating audio")
	}

	audioFile, err := os.Create("output.mp3")
	if err != nil {
		logrus.WithError(err).Fatal("failed to create audio file")
	}
	defer audioFile.Close()

	if _, err = audioFile.Write(audio); err != nil {
		logrus.WithError(err).Fatal("Failed to write audio to file")
	}
}

func GetVarsOrDieTrying() Config {
	lingqAPIKey := os.Getenv("LINGQ_API_KEY")
	if lingqAPIKey == "" {
		logrus.Fatal("LINGQ_API_KEY must be set!")
	}
	lingqDatabasePath := os.Getenv("LINGQ_DATABASE_PATH")
	if lingqDatabasePath == "" {
		logrus.Fatal("LINGQ_DATABASE_PATH must be set!")
	}
	openAIAPIKey := os.Getenv("OPENAI_API_KEY")
	if openAIAPIKey == "" {
		logrus.Fatal("OPENAI_API_KEY must be set!")
	}
	gptModel := os.Getenv("GPT_MODEL")
	if gptModel == "" {
		logrus.Fatal("GPT_MODEL must be set!")
	}

	return Config{
		LingQAPIKey:       lingqAPIKey,
		LingQDatabasePath: lingqDatabasePath,
		OpenAIAPIKey:      openAIAPIKey,
		GPTModel:          gptModel,
	}
}
