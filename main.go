package main

func main() {
	cloneFilesFromKindleIfNeeded()
	db := setupDb()
	ImportHighlights(db)
	ImportVocabDb(db)
}
