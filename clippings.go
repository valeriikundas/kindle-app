package main

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type LocationData struct {
	type_, location string
	time            time.Time
}

// todo: rename `highlights` to `clippings`
func ImportHighlights(db *gorm.DB) {
	notedItems := parseHighlightsFile()
	loadHighlightsToDatabase(db, notedItems)
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

// todo: add logging to main functions

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

func loadHighlightsToDatabase(db *gorm.DB, notedItems []NotedItem) {
	db.Create(&notedItems)
}
