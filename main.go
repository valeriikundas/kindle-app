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
)

const FolderToCloneTo string = "files"
const ClippingsFileName string = "clippings.txt"
const VocabFileName string = "vocab.db"

// FIXME: move to config
const clippingsFilePath string = "/Volumes/Kindle/documents/My Clippings.txt"
const dictFilePath string = "/Volumes/Kindle/system/vocabulary/vocab.db"

func main() {
	shouldCloneFile := false
	if shouldCloneFile {
		cloneKindleFilesToLocalStorage()
	}

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

	for i := 0; i < 5; i++ {
		log.Printf("%+v\n", data[i])
	}

	// TODO: loadHighlightsToDatabase(db, data)
}

// func loadHighlightsToDatabase(db, data){
// }

type Noted struct {
	title  string
	author string

	type_    string
	location string
	time     time.Time

	highlight string
}

func parseHighlightsData(highlights []string) []Noted {
	data := make([]Noted, len(highlights))

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

				if len(lines) > 4 {
					highlight = lines[4]
				}
			}
		}

		data[i] = Noted{
			title:    title,
			author:   author,
			type_:    locationData.type_,
			location: locationData.location,

			time:      locationData.time,
			highlight: highlight,
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
