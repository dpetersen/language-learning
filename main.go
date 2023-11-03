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

Additional Stuff:
	- Get GPT to write you a title and short description of the story
	- Make the prompt add variety to the stories
	- Use DALL-E to generate images for the stories (3 isn't available in the API yet, but whatever)
	- Make it possible to tweak the speed of the speech. It's a little fast right now.
	- Use the SSML <voice> tag to help with dialogue: https://cloud.google.com/text-to-speech/docs/ssml#voice
	- Consider having an English speaker (with the <voice> tag) go over the vocabulary from the story that is new after each?
	- Work in high frequency words from the target language that aren't already
	in your vocabulary. Might be able to just tell ChatGPT this instead of having
	to provide the list:
	  - https://strommeninc.com/1000-most-common-spanish-words-frequency-vocabulary/

*/

type Config struct {
	LingQAPIKey       string
	LingQDatabasePath string
	OpenAIAPIKey      string
	GPTModel          string
	LoadStoryFile     string
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

	words := LoadWords(config)
	story := LoadStory(config, words)

	transcriptFile, err := os.Create("output.json")
	if err != nil {
		logrus.WithError(err).Fatal("failed to create transcript")
	}
	defer transcriptFile.Close()

	if _, err = transcriptFile.WriteString(story.OriginalJSON); err != nil {
		logrus.WithError(err).Fatal("Failed to write to file")
	}

	audio, err := audio.NewAudioClient().TextToSpeech(*story)
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
	loadStoryFile := os.Getenv("LOAD_STORY_FILE")

	return Config{
		LingQAPIKey:       lingqAPIKey,
		LingQDatabasePath: lingqDatabasePath,
		OpenAIAPIKey:      openAIAPIKey,
		GPTModel:          gptModel,
		LoadStoryFile:     loadStoryFile,
	}
}

func LoadWords(config Config) []lingq.Word {
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
	return words
}

func LoadStory(config Config, words []lingq.Word) *gpt.Story {
	storyClient := gpt.NewStoryClient(config.OpenAIAPIKey, config.GPTModel)

	if config.LoadStoryFile == "" {
		story, err := storyClient.Create(words, 3)
		if err != nil {
			logrus.WithError(err).Fatal("Creating story")
		}

		logrus.WithField("storyCharacters", len(story.Description)).Info("Got story")

		return story
	} else {
		logrus.Info("Skipping story generation, loading from file...")
		story, err := storyClient.Load(config.LoadStoryFile)
		if err != nil {
			logrus.WithError(err).Fatal("Loading story from file")
		}
		return story
	}
}
