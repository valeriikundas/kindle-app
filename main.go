package main

func main() {
	cloneFilesFromKindleIfNeeded()
	db := setupDb()
	ImportHighlights(db)
	words := ImportVocabDb()

	saveWordsToCsvFile(words)

	ankiConnectClient := NewAnkiConnectClient(AnkiServerUrl, AnkiDeckName)

	ankiConnectClient.CreateDeckIfNotExists(AnkiDeckName)
	uploadCardsToAnkiConnect(words, ankiConnectClient)
}
