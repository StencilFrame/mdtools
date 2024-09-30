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

			// Initialize a new JSONRenderer
			renderer := NewJSONRenderer()

			// Convert the markdown to JSON
			out := blackfriday.Run(markdownData,
				blackfriday.WithExtensions(blackfriday.CommonExtensions|blackfriday.AutoHeadingIDs|blackfriday.Tables),
				blackfriday.WithRenderer(renderer),
			)

			// Assert the resulting JSON
			expectedData, err := os.ReadFile(tt.expectedFileName)
			assert.NoError(t, err)
			assert.JSONEq(t, string(expectedData), string(out))
		})
	}
}
