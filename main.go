package main

import (
	"os"
	"log"
	"fmt"
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

func main() {
	lingqAPIKey := os.Getenv("LINGQ_API_KEY")
	if lingqAPIKey == "" {
		log.Panic("LINGQ_API_KEY must be set!")
	}
	lingqDatabasePath := os.Getenv("LINGQ_DATABASE_PATH")
	if lingqDatabasePath == "" {
		log.Panic("LINGQ_DATABASE_PATH must be set!")
	}

	log.Println("Checking local database...")
	database := lingq.NewWordDatabase(lingqDatabasePath)
	words, err := database.FetchIfFresh()
	if err != nil {
		log.Panicf("Fetching LingQ words from database: %v", err)
	}

	if words == nil {
		log.Println("Database not fresh, fetching new words...")
		fetchedWords, err := lingq.NewVocabularyClient(lingqAPIKey).GetNonNewWords()
		if err != nil {
			log.Panicf("Getting non-new words from LingQ: %v", err)
		}
		words = fetchedWords

		log.Println("Storing fetched words in database...")
		if err := database.Store(words); err != nil {
			log.Panicf("Storing LingQ words into database: %v", err)
		}
	}

	fmt.Printf("Got: %+v\n\n", words)
	fmt.Printf("Len: %d", len(words))
}
