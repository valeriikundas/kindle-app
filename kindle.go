package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

// todo: rename all `clone` to `copy`
func cloneFilesFromKindleIfNeeded() {
	// todo: check if kindle is connected, then copy

	shouldClone := false

	if shouldClone {
		cloneKindleFilesToLocalStorage()
	}

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
