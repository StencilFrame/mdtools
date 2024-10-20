package mdrenderer_test

import (
	"os"
	"strings"
	"testing"

	"github.com/russross/blackfriday/v2"
	"github.com/stencilframe/mdtools/libs/mdrenderer"
	"github.com/stretchr/testify/assert"
)

func TestMdRenderer(t *testing.T) {
	tests := []struct {
		name             string
		inputFileName    string
		expectedFileName string
	}{
		{
			name:             "Full",
			inputFileName:    "testdata/lists.md",
			expectedFileName: "testdata/lists.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read the markdown file
			markdownData, err := os.ReadFile(tt.inputFileName)
			assert.NoError(t, err)

			// Initialize a new JSONRenderer
			renderer := mdrenderer.NewRenderer()

			// Convert the markdown to JSON
			out := blackfriday.Run(markdownData,
				blackfriday.WithExtensions(blackfriday.CommonExtensions|blackfriday.AutoHeadingIDs|blackfriday.Tables),
				blackfriday.WithRenderer(renderer),
			)

			// Assert the resulting JSON
			expectedData, err := os.ReadFile(tt.expectedFileName)
			assert.NoError(t, err)
			// Trim the whitespaces (TODO: the renderer is adding extra newlines)
			expected := strings.TrimSpace(string(expectedData))
			actual := strings.TrimSpace(string(out))
			assert.Equal(t, expected, actual)
		})
	}
}
