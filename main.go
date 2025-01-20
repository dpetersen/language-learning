package main

import (
	"encoding/base64"
	"os"
	"strings"

	"github.com/dpetersen/language-learning/audio"
	"github.com/dpetersen/language-learning/gpt"
	"github.com/dpetersen/language-learning/lingq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

/*

Todo List:
  - Deal with this error and longer stories:
     INFO[0000] Checking local database...
		 INFO[0000] Database not fresh, fetching new words...
		 INFO[0008] Loaded words                                  count=1435
		 INFO[0008] Generating story...
		 INFO[0075] Generated Story                               storyCharacters=3935
		 INFO[0091] Generating audio...
		 2024/03/01 17:40:50 rpc error: code = InvalidArgument desc = Either `input.text` or `input.ssml` is longer than the limit of 5000 bytes. This limit is different from quotas. To fix, reduce the byte length of the characters in this request, or consider using the Long Audio API: https://cloud.google.com/text-to-speech/docs/create-audio-text-long-audio-synthesis.

	- Maybe use a higher temperature to write the story, then use a lower
	temperature to rewrite it using the correct vocabulary. What I'm getting is
	30-40% unknown words, which is too high. Could even re-rinse the story over
	and over until the unknown words are low enough, although that might run up
	the API costs.
	- Make the prompt add variety to the stories
	  - Summarize Wikipedia pages or news articles
		- Use prompts to generate ideas from best seller lists, etc
		- Maybe just prompt on the CLI for what you want the story to be about?
		- Generate fake "podcasts" (using the voice tag in SSML for this?) about whatever topics you want
		- Look at generate multi-chapter stories that are the right size for a
		lesson. Can have ChatGPT summarize each chapter so you can use that as a
		prompt for the next without blowing through your tokens.

Additional Ideas:
	- Use the SSML <voice> tag to help with dialogue: https://cloud.google.com/text-to-speech/docs/ssml#voice
	- Consider having an English speaker (with the <voice> tag) go over the vocabulary from the story that is new after each?
	  - This doesn't really jive with how LingQ works, unless there's a separate field for "Lesson Notes" or something?
*/

var loadStoryFile string

func init() {
	viper.SetEnvPrefix("LL")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	viper.SetDefault("log_level", "info")
	viper.SetDefault("lingq.database_path", "lingq-data.json")
}

func main() {
	if err := viper.ReadInConfig(); err != nil {
		logrus.WithError(err).Fatal("Error reading config file")
	}

	if parsed, err := logrus.ParseLevel(viper.GetString("log_level")); err == nil {
		logrus.SetLevel(parsed)
	}

	words, lingqClient := LoadWords()
	story := LoadStory(words)

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
	logrus.Info("Generating audio...")
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

	logrus.Info("Importing lesson to LingQ...")
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

func GetVarsOrDieTrying() {
	required := []string{
		"lingq.api_key",
		"openai.api_key",
		"openai.chat_model",
	}

	for _, key := range required {
		if viper.GetString(key) == "" {
			logrus.WithField("key", key).Fatal("Must be set!")
		}
	}
	loadStoryFile = os.Getenv("LOAD_STORY_FILE")
}

func LoadWords() ([]lingq.Word, *lingq.Client) {
	logrus.Info("Checking local database...")
	database := lingq.NewWordDatabase(viper.GetString("lingq.database_path"))
	words, err := database.FetchIfFresh()
	if err != nil {
		logrus.WithError(err).Fatal("Fetching LingQ words from database")
	}

	client := lingq.NewClient(viper.GetString("lingq.api_key"))
	if words == nil {
		logrus.Info("Database not fresh, fetching new words...")
		fetchedWords, err := client.GetNonNewWords()
		if err != nil {
			logrus.WithError(err).Fatal("Getting non-new words from LingQ")
		}
		words = fetchedWords

		if len(words) == 0 {
			logrus.Fatal("No vocabulary found!")
		}
		if err := database.Store(words); err != nil {
			logrus.WithError(err).Fatal("Storing LingQ words into database")
		}
	}

	logrus.WithField("count", len(words)).Info("Loaded words")
	return words, client
}

func LoadStory(words []lingq.Word) *gpt.Story {
	client := gpt.NewClient(
		viper.GetString("openai.api_key"),
		viper.GetString("openai.chat_model"),
	)

	if loadStoryFile == "" {
		logrus.Info("Generating story...")

		story, err := client.CreateStory(words, 3)
		if err != nil {
			logrus.WithError(err).Fatal("Creating story")
		}

		logrus.WithField("storyCharacters", len(story.Story)).Info("Generated Story")

		data, err := client.CreateImage(story.Story)
		if err != nil {
			logrus.WithError(err).Fatal("Creating thumbnail image")
		}

		story.Thumbnail = data
		return story
	} else {
		logrus.Info("Skipping story generation, loading from file...")
		story, err := client.LoadStory(loadStoryFile)
		if err != nil {
			logrus.WithError(err).Fatal("Loading story from file")
		}
		return story
	}
}
