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
	cardToUpload := AnkiCard{

		DeckName:  deckName,
		ModelName: "Basic",
		Fields: AnkiCardFields{
			Front: "test front",
			Back:  "test back",
		},
	}
	ankiClient.CreateCard(deckName, cardToUpload)
}
