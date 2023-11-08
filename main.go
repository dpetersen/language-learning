package main

import (
	"encoding/base64"
	"os"

	"github.com/dpetersen/language-learning/audio"
	"github.com/dpetersen/language-learning/gpt"
	"github.com/dpetersen/language-learning/lingq"
	"github.com/sirupsen/logrus"
)

/*

Todo List:
	- Make the Info logging more consistent about what's happening
		- Probably make it the default level? There's not much to watch if it's on
		a higher level.
	- Make the prompt add variety to the stories
	  - Summarize Wikipedia pages or news articles
		- Use prompts to generate ideas from best seller lists, etc
		- Maybe just prompt on the CLI for what you want the story to be about?
		- Generate fake "podcasts" (using the voice tag in SSML for this?) about whatever topics you want
	- Move your hardcoded stuff into configuration, maybe with Viper

Additional Ideas:
	- Use the SSML <voice> tag to help with dialogue: https://cloud.google.com/text-to-speech/docs/ssml#voice
	- Consider having an English speaker (with the <voice> tag) go over the vocabulary from the story that is new after each?
	  - This doesn't really jive with how LingQ works, unless there's a separate field for "Lesson Notes" or something?
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

	words, lingqClient := LoadWords(config)
	story := LoadStory(config, words)

	// Write Story to JSON
	jsonFile, err := os.Create("output.json")
	if err != nil {
		logrus.WithError(err).Fatal("failed to create transcript")
	}
	defer jsonFile.Close()
	if _, err = jsonFile.WriteString(story.OriginalJSON); err != nil {
		logrus.WithError(err).Fatal("Failed to write to file")
	}

	// Write Story to plain text
	textFile, err := os.Create("output.txt")
	if err != nil {
		logrus.WithError(err).Fatal("failed to create transcript")
	}
	defer textFile.Close()
	if _, err = textFile.WriteString(story.ToString()); err != nil {
		logrus.WithError(err).Fatal("Failed to write to text file")
	}

	// Write thumbnail to PNG
	imageFile, err := os.Create("output.png")
	if err != nil {
		logrus.WithError(err).Fatal("failed to create transcript")
	}
	defer imageFile.Close()
	decodedBytes, err := base64.StdEncoding.DecodeString(story.Thumbnail)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to base64 decode thumbnail")
	}
	if _, err = imageFile.Write(decodedBytes); err != nil {
		logrus.WithError(err).Fatal("Failed to write to file")
	}

	// Write audio to MP3
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

	if err := lingqClient.ImportLesson(
		textFile.Name(),
		audioFile.Name(),
		imageFile.Name(),
		story.Description,
		story.Title,
	); err != nil {
		logrus.WithError(err).Fatal("Importing lesson to LingQ")
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

func LoadWords(config Config) ([]lingq.Word, *lingq.Client) {
	logrus.Info("Checking local database...")
	database := lingq.NewWordDatabase(config.LingQDatabasePath)
	words, err := database.FetchIfFresh()
	if err != nil {
		logrus.WithError(err).Fatal("Fetching LingQ words from database")
	}

	client := lingq.NewClient(config.LingQAPIKey)
	if words == nil {
		logrus.Info("Database not fresh, fetching new words...")
		fetchedWords, err := client.GetNonNewWords()
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
	return words, client
}

func LoadStory(config Config, words []lingq.Word) *gpt.Story {
	client := gpt.NewClient(config.OpenAIAPIKey, config.GPTModel)

	if config.LoadStoryFile == "" {
		story, err := client.CreateStory(words, 3)
		if err != nil {
			logrus.WithError(err).Fatal("Creating story")
		}

		logrus.WithField("storyCharacters", len(story.Description)).Info("Got story")

		data, err := client.CreateImage(story.Story)
		if err != nil {
			logrus.WithError(err).Fatal("Creating thumbnail image")
		}

		story.Thumbnail = data
		return story
	} else {
		logrus.Info("Skipping story generation, loading from file...")
		story, err := client.LoadStory(config.LoadStoryFile)
		if err != nil {
			logrus.WithError(err).Fatal("Loading story from file")
		}
		return story
	}
}
