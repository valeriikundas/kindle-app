package main

import (
	"log"
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
	gorm.Model

	Word string
}

func openVocabDb(vocabDbPath string) *gorm.DB {
	vocabDb, err := gorm.Open(sqlite.Open(vocabDbPath), &gorm.Config{})
	if err != nil {
		log.Fatal("error: couldn't open vocab db")
	}
	return vocabDb
}

func ImportVocabDb(db *gorm.DB) {
	vocabDbPath := filepath.Join(FolderToCloneTo, VocabFileName)
	vocabDb := openVocabDb(vocabDbPath)
	importVocabDbToAppDb(db, vocabDb)
}

func importVocabDbToAppDb(appDb *gorm.DB, vocabDb *gorm.DB) {
	var vocabWords []VocabWord
	vocabDb.Table("WORDS").Unscoped().Find(&vocabWords)

	words := make([]Word, len(vocabWords))
	for i, val := range vocabWords {
		words[i] = Word{
			Word: val.Word,
		}
	}

	if result := appDb.Table("words").Create(&words); result.Error != nil {
		log.Fatal(result.Error)
	}
}
