package main

import (
	"log"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestAnkiSync(t *testing.T) {
	// arrange
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	ankiServerUrl := "http://ankiServerMock.com"

	response := struct {
		result []int64
		error  *string
	}{result: []int64{1496198395707, 1496192395707}, error: nil}
	responder, err := httpmock.NewJsonResponder(200, response)
	if err != nil {
		log.Fatal(err)
	}
	httpmock.RegisterResponder("POST", ankiServerUrl, responder)

	deckName := "testDeck"
	ankiClient := AnkiConnectClient{serverUrl: ankiServerUrl, deckName: deckName}

	// act
	cardsToUpload := []AnkiCard{
		{
			DeckName:  deckName,
			ModelName: "Basic",
			Fields: AnkiCardFields{
				Front: "test front",
				Back:  "test back",
			},
		},
	}
	ankiClient.CreateCards(deckName, cardsToUpload)

	findNotesResponse := AnkiConnectFindNotesResponseSchema{
		Result: []int64{1483959289817, 1483959291695},
		Error:  nil,
	}

	// todo: limit calls to 1
	httpmock.RegisterResponder("POST", ankiServerUrl, httpmock.NewJsonResponderOrPanic(200, findNotesResponse))

	notesInfoResponse := AnkiConnectNotesInfoResponseSchema{
		Result: []AnkiConnectNotesInfoResponseResultSchema{
			{
				NoteID:    1502298033753,
				ModelName: "Basic",
				Tags:      []string{"tag", "another_tag"},
				Fields: AnkiConnectCardFieldsSchema{
					Front: AnkiConnectCardContentSchema{Value: "front content", Order: 0},
					Back:  AnkiConnectCardContentSchema{Value: "back content", Order: 1},
				},
			},
		},
		Error: nil,
	}
	httpmock.RegisterResponder("POST", ankiServerUrl, httpmock.NewJsonResponderOrPanic(200, notesInfoResponse))

	// assert
	cardsOnAnkiServer := ankiClient.FetchCards(deckName)

	if len(cardsOnAnkiServer) != len(cardsToUpload) {
		t.Fatalf("got %d cards on server, want %d cards", len(cardsOnAnkiServer), len(cardsToUpload))
	}

	for i, card := range cardsToUpload {
		if cardsOnAnkiServer[i] != card {
			t.Fatal()
		}
	}
}
