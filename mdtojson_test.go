package main

import (
	"io"
	"os"
	"testing"

	"github.com/russross/blackfriday/v2"
	"github.com/stretchr/testify/assert"
)

func TestJSONRenderer(t *testing.T) {
	tests := []struct {
		name             string
		inputFileName    string
		expectedFileName string
	}{
		{
			name:             "Headers",
			inputFileName:    "testdata/headers.md",
			expectedFileName: "testdata/headers.json",
		},
		{
			name:             "Tables",
			inputFileName:    "testdata/tables.md",
			expectedFileName: "testdata/tables.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read the markdown file
			markdownData, err := os.ReadFile(tt.inputFileName)
			assert.NoError(t, err)

			// Parse the markdown into a syntax tree
			parser := blackfriday.New(blackfriday.WithExtensions(blackfriday.CommonExtensions | blackfriday.AutoHeadingIDs | blackfriday.Tables))
			node := parser.Parse(markdownData)

			// Create a new JSONRenderer
			renderer := NewJSONRenderer()

			// Create a buffered writer
			buffer := new(bytes.Buffer)

			// Walk the parsed syntax tree with our custom renderer
			node.Walk(func(n *blackfriday.Node, entering bool) blackfriday.WalkStatus {
				return renderer.RenderNode(buffer, n, entering)
			})
			// finalize the JSON
			renderer.RenderFooter(buffer, node)

			// Assert the resulting JSON
			expectedData, err := os.ReadFile(tt.expectedFileName)
			assert.NoError(t, err)
			assert.JSONEq(t, string(expectedData), buffer.String())
		})
	}
}
