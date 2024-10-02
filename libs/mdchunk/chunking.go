package mdchunk

import (
	"io"
	"mdtools/libs/mdtojson"

	"github.com/russross/blackfriday/v2"
)

// Token limit per chunk (e.g., 200 tokens)
const defaultTokenLimit = 200

// Charecter limit per chunk (e.g., 1000 charecters)
const defaultCharLimit = 1000

// MarkdownChunk represents a chunk of the markdown document.
type MarkdownChunk struct {
	TokenCount int // Number of tokens in the chunk
	CharCount  int // Number of charecters in the chunk
}

// NewMarkdownChunk creates a new MarkdownChunk.
func NewMarkdownChunk() *MarkdownChunk {
	return &MarkdownChunk{
		TokenCount: defaultTokenLimit,
		CharCount:  defaultCharLimit,
	}
}

// ChunkMarkdown splits the markdown data into chunks.
func (mc *MarkdownChunk) ChunkMarkdown(markdownData []byte) []string {
	// Parse the markdown into a syntax tree
	parser := blackfriday.New(blackfriday.WithExtensions(blackfriday.CommonExtensions | blackfriday.AutoHeadingIDs | blackfriday.Tables))
	node := parser.Parse(markdownData)

	// Create a new JSONRenderer
	renderer := mdtojson.NewJSONRenderer()

	// Walk the parsed syntax tree with our custom renderer
	node.Walk(func(n *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		return renderer.RenderNode(io.Discard, n, entering)
	})
	nodes := renderer.GetNodes()
	return mc.ChunkJSONMarkdown(1000, nodes)
}

// ChunkJSONMarkdown splits the JSON markdown data into chunks.
func (mc *MarkdownChunk) ChunkJSONMarkdown(charLimit int, markdownData []mdtojson.Node) []string {
	chunks := []string{}
	currentChunk := ""

	for i := 0; i < len(markdownData); i++ {
		section := markdownData[i].ToMarkdown()
		sectionLen := len(section)
		currentChunk += section

		// Process the children of the current node first
		childs := markdownData[i].GetChildren()
		if childs != nil {
			childrenChunks := mc.ChunkJSONMarkdown(charLimit-sectionLen, childs)

			for _, child := range childrenChunks {
				// Try to append the child to the current chunk
				if len(currentChunk)+len(child) > charLimit {
					// If the current chunk is too large, finalize it
					chunks = append(chunks, currentChunk)
					currentChunk = section // Reset to the parent section, continuing the structure
				}
				currentChunk += child
			}
		}

		if markdownData[i].GetType() == "paragraph" {
			currentChunk += "\n\n"
		}

		if currentChunk != section {
			// If the section alone is larger than charLimit, add it as a single chunk
			if len(currentChunk) > charLimit {
				chunks = append(chunks, currentChunk)
				currentChunk = section // Reset to the current section
			}
		}
	}

	// Add any remaining content in currentChunk as the last chunk
	if len(currentChunk) > 0 {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}
