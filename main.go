package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const FolderToCloneTo string = "files"
const ClippingsFileName string = "clippings.txt"
const VocabFileName string = "vocab.db"

// FIXME: move to config
const clippingsFilePath string = "/Volumes/Kindle/documents/My Clippings.txt"
const dictFilePath string = "/Volumes/Kindle/system/vocabulary/vocab.db"

func main() {
	cloneFilesFromKindleIfNeeded()

	db := setupDb()

	// importHighlights(db)

	vocabDbPath := filepath.Join(FolderToCloneTo, VocabFileName)
	vocabDb := openVocabDb(vocabDbPath)
	importVocabDb(db, vocabDb)
}

// todo: rename all `clone` to `copy`
func cloneFilesFromKindleIfNeeded() {
	// todo: check if kindle is connected, then copy
	cloneKindleFilesToLocalStorage()

}

func importHighlights(db *gorm.DB) {
	notedItems := parseHighlightsFile()
	loadHighlightsToDatabase(db, notedItems)
}

func setupDb() *gorm.DB {
	db := connectToDb()
	migrateDb(db)
	return db
}

func connectToDb() *gorm.DB {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: "host=0.0.0.0 port=5432 user=postgres password=pass dbname=kindle_db",
	}), &gorm.Config{})

	if err != nil {
		log.Fatal("failed to connect to database")
	}
	return db
}

func migrateDb(db *gorm.DB) {
	db.AutoMigrate(&NotedItem{}, &Word{})
}

func parseHighlightsFile() []NotedItem {
	bytes, err := os.ReadFile(filepath.Join(FolderToCloneTo, ClippingsFileName))
	if err != nil {
		log.Fatal(err)
	}

	// TODO: why does it return bytes in golang? understand it
	text := string(bytes)

	// TODO: looks strange, is it ok?
	highlights := strings.Split(text, "==========")
	highlights = highlights[:len(highlights)-1]
	log.Printf("total highlights amount is %d", len(highlights))

	data := parseHighlightsData(highlights)
	return data
}

func loadHighlightsToDatabase(db *gorm.DB, notedItems []NotedItem) {
	db.Create(&notedItems)
}

type NotedItem struct {
	gorm.Model

	// todo: make title + time as primary key
	// todo: make it polymorphic as in sqlalchemy

	// todo: move title and author to separate tables
	// todo: add index for not duplicating notes
	Title  string
	Author string

	// todo: convert to union type
	Type_    string
	Location string
	Time     time.Time

	Highlight string
}

func parseHighlightsData(highlights []string) []NotedItem {
	data := make([]NotedItem, len(highlights))

	for i := 0; i < len(highlights); i++ {
		highlightBlock := highlights[i]
		highlightBlock = strings.Trim(highlightBlock, "\r\n\ufeff")

		lines := strings.Split(highlightBlock, "\n")

		var title, author, highlight string
		var locationData LocationData

		if len(lines) > 0 {
			title, author = parseTitleAndAuthor(lines[0])

			if len(lines) > 1 {
				locationData = parseLocationString(lines[1])

				if len(lines) > 3 {
					highlight = lines[3]
				}
			}
		}

		data[i] = NotedItem{
			Title:     title,
			Author:    author,
			Type_:     locationData.type_,
			Location:  locationData.location,
			Time:      locationData.time,
			Highlight: highlight,
		}
	}

	return data
}

type LocationData struct {
	type_, location string
	time            time.Time
}

// FIXME: regexp compilation should be extracted, so they are evaluated once
// FIXME: should return struct
func parseLocationString(s string) LocationData {
	regExp := regexp.MustCompile(`- Your (Highlight|Bookmark) on Location ([\d-]+) \| Added on .+ (January|February|August) (\d{1,2}), (\d{4}) (\d{1,2}):(\d{2}):(\d{2}) (AM|PM)`)
	matches := regExp.FindStringSubmatch(s)

	var type_, location string
	var day, year, hour, minutes, seconds int
	var month, daytime string

	if len(matches) > 1 {
		type_ = matches[1]

		if len(matches) > 2 {
			location = matches[2]

			if len(matches) > 5 {
				var err error

				month = matches[3]

				day, err = strconv.Atoi(matches[4])
				if err != nil {
					log.Fatalf("error converting day to int in %s", matches[4])
				}

				year, err = strconv.Atoi(matches[5])
				if err != nil {
					log.Fatalf("error converting year to int in %s", matches[5])
				}
			}

			if len(matches) > 4 {
				var err error

				hour, err = strconv.Atoi(matches[6])
				if err != nil {
					log.Fatalf("error converting year to int in %s", matches[6])
				}

				minutes, err = strconv.Atoi(matches[7])
				if err != nil {
					log.Fatalf("error converting minutes to int in %s", matches[7])
				}

				seconds, err = strconv.Atoi(matches[8])
				if err != nil {
					log.Fatalf("error converting seconds to int in %s", matches[8])
				}

				daytime = matches[9]
				if daytime == "PM" {
					hour += 12
				}

			}
		}
	}

	return LocationData{
		type_:    type_,
		location: location,
		time:     time.Date(year, time.Month(monthToIndex(month)), day, hour, minutes, seconds, 0, time.UTC),
	}
}

func monthToIndex(month string) int {
	d := map[string]int{
		"January":   1,
		"February":  2,
		"March ":    3,
		"April":     4,
		"May":       5,
		"June":      6,
		"July":      7,
		"August":    8,
		"September": 9,
		"October":   10,
		"November":  11,
		"December":  12,
	}
	return d[month]
}

func parseTitleAndAuthor(s string) (string, string) {
	s = strings.Trim(s, "\r")
	regExp := regexp.MustCompile(`(.+) \((.+)\)`)
	matches := regExp.FindStringSubmatch(s)

	var title, author string

	if len(matches) > 1 {
		title = matches[1]
	}
	if len(matches) > 2 {
		author = matches[2]
	}

	return title, author
}

func cloneKindleFilesToLocalStorage() {
	cloneFile(clippingsFilePath, ClippingsFileName)
	cloneFile(dictFilePath, VocabFileName)
}

func cloneFile(srcPath string, dstName string) {
	src, err := os.Open(srcPath)
	if err != nil {
		log.Fatal(err)
	}
	defer src.Close()

	os.MkdirAll(FolderToCloneTo, 0755)

	dstPath := filepath.Join(FolderToCloneTo, dstName)
	dst, err := os.Create(dstPath)
	if err != nil {
		log.Fatal(err)
	}
	defer dst.Close()

	bytesWritten, err := io.Copy(dst, src)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Has copied %d bytes from %s to %s", bytesWritten, srcPath, dstPath)
}

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

func importVocabDb(appDb *gorm.DB, vocabDb *gorm.DB) {
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

func openVocabDb(vocabDbPath string) *gorm.DB {
	vocabDb, err := gorm.Open(sqlite.Open(vocabDbPath), &gorm.Config{})
	if err != nil {
		log.Fatal("error: couldn't open vocab db")
	}
	return vocabDb
}
