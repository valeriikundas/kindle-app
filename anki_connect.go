package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type AnkiConnectClient struct {
	serverUrl string
	deckName  string
}

func NewAnkiConnectClient(serverUrl string, deckName string) *AnkiConnectClient {
	isAnkiConnectServerAvailable(serverUrl)

	return &AnkiConnectClient{
		serverUrl: serverUrl,
		deckName:  deckName,
	}
}

func isAnkiConnectServerAvailable(serverUrl string) {
	resp, err := http.Get(serverUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	body := string(bytes)
	isOk := strings.HasPrefix(body, "AnkiConnect")
	if !isOk {
		panic("couldn't connect to AnkiConnect server")
	}
}

type AnkiConnectFindNotesResponseSchema struct {
	Result []int64     `json:"result"`
	Error  interface{} `json:"error"`
}

type AnkiConnectNotesInfoResponseSchema struct {
	Result []AnkiConnectNotesInfoResponseResultSchema `json:"result"`
	Error  interface{}                                `json:"error"`
}

type AnkiConnectNotesInfoResponseResultSchema struct {
	NoteID    int64                       `json:"noteId"`
	ModelName string                      `json:"modelName"`
	Tags      []string                    `json:"tags"`
	Fields    AnkiConnectCardFieldsSchema `json:"fields"`
}

type AnkiConnectCardFieldsSchema struct {
	Front AnkiConnectCardContentSchema `json:"Front"`
	Back  AnkiConnectCardContentSchema `json:"Back"`
}

type AnkiCard struct {
	DeckName  string         `json:"deckName"`
	ModelName string         `json:"modelName"`
	Fields    AnkiCardFields `json:"fields"`
}

type AnkiCardFields struct {
	Front string
	Back  string
}

type AnkiConnectCardContentSchema struct {
	Value string `json:"value"`
	Order int    `json:"order"`
}

func (ankiClient AnkiConnectClient) FetchCards(deckName string) []AnkiCard {
	insertedNoteIds := ankiClient.fetchNoteIds()
	cards := ankiClient.fetchNotes(deckName, insertedNoteIds)
	return cards
}

type AnkiConnectCreateCardParamsNoteSchema struct {
	Note AnkiCard `json:"note"`
}

type CreateCardsRequestBodySchema struct {
	Action  string                                `json:"action"`
	Version int                                   `json:"version"`
	Params  AnkiConnectCreateCardParamsNoteSchema `json:"params"`
}

func (ankiConnectClient AnkiConnectClient) CreateCard(deckName string, card AnkiCard) {
	requestBody := CreateCardsRequestBodySchema{
		Action:  "addNote",
		Version: 6,
		Params: AnkiConnectCreateCardParamsNoteSchema{
			Note: card,
		},
	}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(requestBody)
	if err != nil {
		log.Print(err)
	}

	resp, err := http.Post(ankiConnectClient.serverUrl, "application/json", &buf)
	if err != nil {
		log.Printf("couldn't upload word to AnkiConnect server. details: %s", err)
	}
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
	}

	var res struct {
		Result *string `json:"result"`
		Error  *string `json:"error"`
	}
	err = json.Unmarshal(bytes, &res)
	if err != nil {
		log.Printf("error unmarshaling upload response. details: %s. res=%+v\n", err, res)
	}

	if res.Error != nil {
		log.Printf("error uploading word to AnkiConnect server. details: %s", *res.Error)
	}
}

func (ankiClient AnkiConnectClient) fetchNoteIds() []int64 {
	var buf bytes.Buffer
	requestBody := struct {
		Action  string `json:"action"`
		Version int    `json:"version"`
		Params  struct {
			Query string `json:"query"`
		} `json:"params"`
	}{
		Action:  "findNotes",
		Version: 6,
		Params: struct {
			Query string `json:"query"`
		}{
			Query: fmt.Sprintf("deck:%s", ankiClient.deckName),
		},
	}
	json.NewEncoder(&buf).Encode(requestBody)
	resp, err := http.Post(ankiClient.serverUrl, "application/json", &buf)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var findNotesResponse AnkiConnectFindNotesResponseSchema
	err = json.Unmarshal(bytes, &findNotesResponse)
	if err != nil {
		log.Fatalf("failed to unmarshal findNotes response\n json: %s\n details: %s", bytes, err)
	}

	insertedNoteIds := findNotesResponse.Result
	return insertedNoteIds
}

func (ankiClient AnkiConnectClient) fetchNotes(deckName string, noteIds []int64) []AnkiCard {
	requestBody := struct {
		Action  string `json:"action"`
		Version int    `json:"version"`
		Params  struct {
			Notes []int64 `json:"notes"`
		} `json:"params"`
	}{
		Action:  "notesInfo",
		Version: 6,
		Params: struct {
			Notes []int64 `json:"notes"`
		}{
			Notes: noteIds,
		},
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(requestBody)
	resp, err := http.Post(ankiClient.serverUrl, "application/json", &buf)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var notesInfoResponse AnkiConnectNotesInfoResponseSchema
	err = json.Unmarshal(bytes, &notesInfoResponse)
	if err != nil {
		log.Fatalf("failed to unmarshmal notes info response\njson=%s\nerror=%s", bytes, err)
	}

	notes := notesInfoResponse.Result
	cards := make([]AnkiCard, len(notes))
	for i, val := range notesInfoResponse.Result {
		cards[i] = AnkiCard{
			DeckName:  deckName,
			ModelName: val.ModelName,
			Fields: AnkiCardFields{
				Front: val.Fields.Front.Value,
				Back:  val.Fields.Back.Value,
			},
		}
	}

	return cards
}

func (ankiClient AnkiConnectClient) CreateDeckIfNotExists(deckName string) {
	requestBody := map[string]interface{}{
		"action":  "createDeck",
		"version": 6,
		"params": map[string]interface{}{
			"deck": deckName,
		},
	}

	var buf bytes.Buffer
	bytes, err := json.Marshal(requestBody)
	if err != nil {
		log.Fatal(err)
	}

	buf.Write(bytes)

	resp, err := http.Post(ankiClient.serverUrl, "application/json", &buf)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	bytes, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var res struct {
		Result int64   `json:"result"`
		Error  *string `json:"error"`
	}
	err = json.Unmarshal(bytes, &res)
	if err != nil {
		log.Fatal(err)
	}
	if res.Error != nil {
		log.Fatal(res.Error)
	}
	log.Print(res.Result)
}

func uploadCardsToAnkiConnect(words []VocabWord, ankiConnectClient *AnkiConnectClient) {
	for _, word := range words {
		translation := getWordTranslation(word.Word)
		card := AnkiCard{
			DeckName:  ankiConnectClient.deckName,
			ModelName: "Basic",
			Fields: AnkiCardFields{
				Front: word.Word,
				Back:  translation,
			},
		}
		ankiConnectClient.CreateCard(AnkiDeckName, card)
	}
}

func getWordTranslation(s string) string {
	// todo: use online translate api
	panic("unimplemented")
}
