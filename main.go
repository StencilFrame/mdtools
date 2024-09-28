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

	// Parse the markdown into a syntax tree
	parser := blackfriday.New(blackfriday.WithExtensions(blackfriday.CommonExtensions | blackfriday.AutoHeadingIDs | blackfriday.Tables))
	node := parser.Parse(markdownData)

	// Walk the parsed syntax tree with our custom renderer
	node.Walk(func(n *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		return renderer.RenderNode(os.Stdout, n, entering)
	})

	// Print the resulting JSON
	renderer.RenderFooter(os.Stdout, node)
}
