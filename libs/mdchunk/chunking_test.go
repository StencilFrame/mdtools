package mdchunk

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChunking(t *testing.T) {
	tests := []struct {
		name                   string
		inputFileName          string
		expectedChunksFileName string
		expectedImagesFileName string
		chunkSize              int
	}{
		{
			name:                   "Headers",
			inputFileName:          "testdata/headers.md",
			expectedChunksFileName: "testdata/headers.chunked.md",
			chunkSize:              1000,
		},
		{
			name:                   "Tables",
			inputFileName:          "testdata/tables.md",
			expectedChunksFileName: "testdata/tables.chunked.md",
			chunkSize:              1000,
		},
		{
			name:                   "Images",
			inputFileName:          "testdata/images.md",
			expectedChunksFileName: "testdata/images.chunked.md",
			expectedImagesFileName: "testdata/images.chunked.json",
			chunkSize:              100,
		},

		// TODO: Implement the following tests
		// {
		// 	name:             "Lists",
		// 	inputFileName:    "testdata/lists.md",
		// 	expectedFileName: "testdata/lists.chunked.md",
		// },
		// {
		// 	name:             "Links",
		// 	inputFileName:    "testdata/links.md",
		// 	expectedFileName: "testdata/links.chunked.md",
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read the markdown file
			markdownData, err := os.ReadFile(tt.inputFileName)
			assert.NoError(t, err)

			// Initialize a new JSONRenderer
			chunker := NewMarkdownChunk(tt.chunkSize)

			// Chunk the markdown
			chunks, images := chunker.ChunkMarkdown(markdownData)

			results := ""
			for i, chunk := range chunks {
				chunk = strings.TrimSpace(chunk)
				l := len(chunk)
				results += fmt.Sprintf("%s\n\n--- CHUNK BREAK [id: %d, len: %d] ---\n\n", chunk, i, l)
			}

			// Assert the resulting JSON
			expectedData, err := os.ReadFile(tt.expectedChunksFileName)
			assert.NoError(t, err)
			assert.Equal(t, string(expectedData), results)

			// Assert the images
			if tt.expectedImagesFileName != "" {
				expectedImagesData, err := os.ReadFile(tt.expectedImagesFileName)
				assert.NoError(t, err)
				expectedImages := []string{}
				err = json.Unmarshal(expectedImagesData, &expectedImages)
				assert.NoError(t, err)
				// NOTICE the order of the images is important
				assert.Equal(t, expectedImages, images)
			}
		})
	}
}
