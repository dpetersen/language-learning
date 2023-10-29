package lingq

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

type WordDatabase struct {
	path string
}

type storageFormat struct {
	Words []Word
	Time  time.Time
}

func NewWordDatabase(path string) *WordDatabase {
	return &WordDatabase{path: path}
}

func (wd *WordDatabase) Store(words []Word) error {
	data := storageFormat{
		Words: words,
		Time:  time.Now(),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshalling JSON: %v", err)
	}

	file, err := os.Create(wd.path)
	if err != nil {
		return fmt.Errorf("creating database file: %v", err)
	}
	defer file.Close()

	if _, err = io.WriteString(file, string(jsonData)); err != nil {
		return fmt.Errorf("writing data: %v", err)
	}

	return nil
}

func (wd *WordDatabase) FetchIfFresh() ([]Word, error) {
	file, err := os.Open(wd.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil // Return nil slice and nil error if file does not exist
		}

		return nil, fmt.Errorf("opening database file : %v", err)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("reading database file : %v", err)
	}

	var storedData storageFormat
	err = json.Unmarshal(data, &storedData)
	if err != nil {
		return nil, fmt.Errorf("Deserializing database file : %v", err)
	}

	if time.Since(storedData.Time).Minutes() >= 60 {
		return nil, nil // Return nil slice and nil error if data is not fresh
	}

	return storedData.Words, nil
}
