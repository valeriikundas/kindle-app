package main

import (
	"encoding/csv"
	"log"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type VocabWord struct {
	Word      string
	Stem      string
	Lang      string
	Category  int
	Timestamp int
	Profileid string
}

type Word struct {
	Word string
}

func openVocabDb(vocabDbPath string) *gorm.DB {
	vocabDb, err := gorm.Open(sqlite.Open(vocabDbPath), &gorm.Config{})
	if err != nil {
		log.Fatal("error: couldn't open vocab db")
	}
	return vocabDb
}

func ImportVocabDb() []VocabWord {
	vocabDbPath := filepath.Join(FolderToCloneTo, VocabFileName)
	vocabDb := openVocabDb(vocabDbPath)
	words := getWordsFromVocabDb(vocabDb)
	return words
}

func getWordsFromVocabDb(vocabDb *gorm.DB) []VocabWord {
	var vocabWords []VocabWord
	vocabDb.Table("WORDS").Unscoped().Find(&vocabWords)
	return vocabWords
}

// todo: delete db, don't store vocab in my db
func GetAllCards(db *gorm.DB, deckName string) []AnkiCard {
	var words []Word
	tx := db.Table("words").Find(&words)
	if tx.Error != nil {
		log.Fatal(tx.Error)
	}

	cards := make([]AnkiCard, len(words))
	// todo: extract modelName to consts
	for i, word := range words {
		cards[i] = AnkiCard{
			DeckName:  deckName,
			ModelName: "Basic",
			Fields: AnkiCardFields{
				Front: word.Word,
				Back:  "todo:",
			},
		}
	}

	return cards
}

func saveWordsToCsvFile(words []VocabWord) {
	csvFile, err := os.Create("words.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer csvFile.Close()

	csvWriter := csv.NewWriter(csvFile)
	defer csvWriter.Flush()

	for _, val := range words {

		record := []string{val.Word, val.Stem}
		csvWriter.Write(record)
	}
}
