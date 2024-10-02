package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/stencilframe/mdtools/libs/mdchunk"
)

func main() {
	// Check if a file was provided as an argument
	if len(os.Args) < 2 {
		log.Fatal("Please provide a markdown file as an argument")
	}

	// Read the markdown file
	markdownFile := os.Args[1]
	markdownData, err := os.ReadFile(markdownFile)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	chunker := mdchunk.NewMarkdownChunk()
	chunks := chunker.ChunkMarkdown(markdownData)

	for i, chunk := range chunks {
		chunk = strings.TrimSpace(chunk)
		os.Stdout.WriteString(chunk)
		l := len(chunk)
		fmt.Printf("\n\n--- CHUNK BREAK [id: %d, len: %d] ---\n\n", i, l)
	}
}
