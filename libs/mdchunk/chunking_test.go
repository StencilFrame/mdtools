package mdchunk

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChunking(t *testing.T) {
	tests := []struct {
		name             string
		inputFileName    string
		expectedFileName string
	}{
		{
			name:             "Headers",
			inputFileName:    "testdata/headers.md",
			expectedFileName: "testdata/headers.chunked.md",
		},
		{
			name:             "Tables",
			inputFileName:    "testdata/tables.md",
			expectedFileName: "testdata/tables.chunked.md",
		},

		// TODO: Implement the following tests
		// {
		// 	name:             "Lists",
		// 	inputFileName:    "testdata/lists.md",
		// 	expectedFileName: "testdata/lists.json",
		// },
		// {
		// 	name:             "Links",
		// 	inputFileName:    "testdata/links.md",
		// 	expectedFileName: "testdata/links.json",
		// },
		// {
		// 	name:             "Images",
		// 	inputFileName:    "testdata/images.md",
		// 	expectedFileName: "testdata/images.json",
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read the markdown file
			markdownData, err := os.ReadFile(tt.inputFileName)
			assert.NoError(t, err)

			// Initialize a new JSONRenderer
			chunker := NewMarkdownChunk(1000)

			// Chunk the markdown
			chunks := chunker.ChunkMarkdown(markdownData)

			results := ""
			for i, chunk := range chunks {
				chunk = strings.TrimSpace(chunk)
				l := len(chunk)
				results += fmt.Sprintf("%s\n\n--- CHUNK BREAK [id: %d, len: %d] ---\n\n", chunk, i, l)
			}

			// Assert the resulting JSON
			expectedData, err := os.ReadFile(tt.expectedFileName)
			assert.NoError(t, err)
			assert.Equal(t, string(expectedData), results)
		})
	}
}
