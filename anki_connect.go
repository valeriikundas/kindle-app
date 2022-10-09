package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type AnkiConnectClient struct {
	serverUrl string
	deckName  string
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
	deckName  string
	modelName string
	fields    AnkiCardFields
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

func (ankiConnectClient AnkiConnectClient) CreateCards(deckName string, cards []AnkiCard) {
	// todo: use arg for initialization of `notes` var
	notes := []AnkiCard{
		{
			deckName:  deckName,
			modelName: "Basic",
			fields: AnkiCardFields{
				Front: "front_content",
				Back:  "back_content",
			},
		},
	}

	requestBody := map[string]interface{}{
		"action":  "addNotes",
		"version": 6,
		"params": struct{ notes []AnkiCard }{
			notes: notes,
		},
	}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(requestBody)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(ankiConnectClient.serverUrl, "application/json", &buf)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
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
			Query: "deck:Kindle Vocab",
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
		log.Fatalf("failed to unmarshall findNotes response\n json: %s\n details: %s", bytes, err)
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
		log.Fatalf("failed to unmarshmall notes info response\njson=%s\nerror=%s", bytes, err)
	}

	notes := notesInfoResponse.Result
	cards := make([]AnkiCard, len(notes))
	for i, val := range notesInfoResponse.Result {
		cards[i] = AnkiCard{
			deckName:  deckName,
			modelName: val.ModelName,
			fields: AnkiCardFields{
				Front: val.Fields.Front.Value,
				Back:  val.Fields.Back.Value,
			},
		}
	}

	return cards
}
