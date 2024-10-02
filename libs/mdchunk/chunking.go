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

	for i := 0; i < len(markdownData); i++ {
		// Add the current node to the section
		section := markdownData[i].ToMarkdown()
		length := len(section)

		// Process the children of the current node
		childs := markdownData[i].GetChildren()
		if childs != nil {
			children := mc.ChunkJSONMarkdown(charLimit-length, childs)
			nn := section
			for _, child := range children {
				nn += child
				if len(nn) > charLimit {
					chunks = append(chunks, nn)
					nn = section
				}
			}

			if nn != section {
				chunks = append(chunks, nn)
			}
		} else if section != "" {
			chunks = append(chunks, section)
		}
	}

	return chunks
}
