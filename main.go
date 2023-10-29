package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dpetersen/language-learning/gpt"
	"github.com/dpetersen/language-learning/lingq"
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

func main() {
	config := GetVarsOrDieTrying()

	log.Println("Checking local database...")
	database := lingq.NewWordDatabase(config.LingQDatabasePath)
	words, err := database.FetchIfFresh()
	if err != nil {
		log.Panicf("Fetching LingQ words from database: %v", err)
	}

	if words == nil {
		log.Println("Database not fresh, fetching new words...")
		fetchedWords, err := lingq.NewVocabularyClient(config.LingQAPIKey).GetNonNewWords()
		if err != nil {
			log.Panicf("Getting non-new words from LingQ: %v", err)
		}
		words = fetchedWords

		log.Println("Storing fetched words in database...")
		if err := database.Store(words); err != nil {
			log.Panicf("Storing LingQ words into database: %v", err)
		}
	}

	log.Printf("Loaded %d words...", len(words))

	storyClient := gpt.NewStoryClient(config.OpenAIAPIKey, config.GPTModel)
	story, err := storyClient.Create(words, 3)
	if err != nil {
		log.Panicf("Creating story: %v", err)
	}

	fmt.Printf("Got story: %s\n", story)
}

func GetVarsOrDieTrying() Config {
	lingqAPIKey := os.Getenv("LINGQ_API_KEY")
	if lingqAPIKey == "" {
		log.Panic("LINGQ_API_KEY must be set!")
	}
	lingqDatabasePath := os.Getenv("LINGQ_DATABASE_PATH")
	if lingqDatabasePath == "" {
		log.Panic("LINGQ_DATABASE_PATH must be set!")
	}
	openAIAPIKey := os.Getenv("OPENAI_API_KEY")
	if openAIAPIKey == "" {
		log.Panic("OPENAI_API_KEY must be set!")
	}
	gptModel := os.Getenv("GPT_MODEL")
	if gptModel == "" {
		log.Panic("GPT_MODEL must be set!")
	}

	return Config{
		LingQAPIKey:       lingqAPIKey,
		LingQDatabasePath: lingqDatabasePath,
		OpenAIAPIKey:      openAIAPIKey,
		GPTModel:          gptModel,
	}
}
