package main

import (
	"log"
	"os"

	"github.com/russross/blackfriday/v2"
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

	// Initialize a new JSONRenderer
	renderer := NewJSONRenderer()

	// Convert the markdown to JSON
	out := blackfriday.Run(markdownData,
		blackfriday.WithExtensions(blackfriday.CommonExtensions|blackfriday.AutoHeadingIDs|blackfriday.Tables),
		blackfriday.WithRenderer(renderer),
	)

	// Write the JSON to stdout
	os.Stdout.Write(out)
}
