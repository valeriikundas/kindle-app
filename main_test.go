package main

import (
	"fmt"
	"testing"
	"time"
)

func TestParseTitleAndAuthor(t *testing.T) {
	s := "Sleep: The Myth of 8 Hours, the Power of Naps, and the New Plan to Recharge Your Body and Mind by Nick Littlehales (Littlehales, Nick)"
	title, author := parseTitleAndAuthor(s)
	if title != "Sleep: The Myth of 8 Hours, the Power of Naps, and the New Plan to Recharge Your Body and Mind by Nick Littlehales" {
		t.Errorf("Parsed title incorrectly")
	}
	if author != "Littlehales, Nick" {
		t.Errorf("Parsed author incorrectly")
	}
}

// FIXME: fix timezone usage

func TestParseLocationString(t *testing.T) {
	tests := []struct {
		locationString, type_, location string
		time                            time.Time
	}{{
		"- Your Highlight on Location 84-88 | Added on Wednesday, January 26, 2022 11:19:28 PM",
		"Highlight",
		"84-88",
		time.Date(2022, 1, 26, 23, 19, 28, 0, time.UTC),
	},
		{
			"- Your Bookmark on Location 6524 | Added on Thursday, August 25, 2022 11:25:27 PM",
			"Bookmark",
			"6524",
			time.Date(2022, 8, 25, 23, 25, 27, 0, time.UTC),
		}}

	for _, tt := range tests {
		testName := fmt.Sprintf("TestParseLocationString - %s", tt.locationString)
		t.Run(testName, func(t *testing.T) {
			locationData := parseLocationString(tt.locationString)
			if locationData.type_ != tt.type_ {
				t.Errorf("error type. received %s, expected %s", locationData.type_, tt.type_)
			}

			if locationData.location != tt.location {
				t.Errorf("error location, received %s, expected %s", locationData.location, tt.location)
			}

			if locationData.time != tt.time {
				t.Errorf("error date, received %s, expected %s", locationData.time, tt.time)
			}
		})
	}
}

// TODO: test with multiline string that parses full highlighted block
// func TestParseHighlightBlock(t *testing.T) {
// 	s := "Meditation for Fidgety Skeptics: A 10% Happier How-to Book (Dan Harris)"
// 	"- Your Highlight on Location 84-84 | Added on Wednesday, January 26, 2022 11:19:06 PM"
// 	""
// 	"The first was the science. In"

// 	log.Println(s)
// }
