package mdchunk

import (
	"fmt"
	"io"

	"github.com/russross/blackfriday/v2"
	"github.com/stencilframe/mdtools/libs/mdtojson"
)

// Charecter limit per chunk (e.g., 4000 charecters)
const defaultCharLimit = 4000

// MarkdownChunk represents a chunk of the markdown document.
type MarkdownChunk struct {
	CharCount int // Number of charecters in the chunk
}

// NewDefaultMarkdownChunk creates a new MarkdownChunk.
func NewDefaultMarkdownChunk() *MarkdownChunk {
	return &MarkdownChunk{
		CharCount: defaultCharLimit,
	}
}

// NewMarkdownChunk creates a new MarkdownChunk with custom charecter limit.
func NewMarkdownChunk(charLimit int) *MarkdownChunk {
	return &MarkdownChunk{
		CharCount: charLimit,
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
	return mc.ChunkJSONMarkdown(mc.CharCount, nodes)
}

// ChunkJSONMarkdown splits the JSON markdown data into chunks.
func (mc *MarkdownChunk) ChunkJSONMarkdown(charLimit int, markdownData []mdtojson.Node) []string {
	chunks := []string{}
	currentChunk := ""

	for i := 0; i < len(markdownData); i++ {
		switch markdownData[i].GetType() {
		case mdtojson.NodeTypeTable:
			// Chunk tables separately
			table, ok := markdownData[i].(*mdtojson.TableNode)
			if !ok {
				fmt.Println("Error: Unable to cast to TableNode")
				continue
			}
			tableChunks := table.ChunkTable(charLimit-len(currentChunk), charLimit)
			if len(tableChunks) == 0 {
				continue
			}

			// Append the first table chunk to the current chunk
			tableChunks[0] = currentChunk + tableChunks[0]
			// Append all but the last table chunk to the chunks list
			chunks = append(chunks, tableChunks[:len(tableChunks)-1]...)

			currentChunk = tableChunks[len(tableChunks)-1]
			// If the current chunk is too large, finalize it
			if len(currentChunk) > charLimit {
				chunks = append(chunks, currentChunk)
				currentChunk = ""
			}

			continue
		case mdtojson.NodeTypeImage:
			// Extract images
			image, ok := markdownData[i].(*mdtojson.ImageNode)
			if !ok {
				fmt.Println("Error: Unable to cast to ImageNode")
				continue
			}

			// Add the image reference to the current chunk
			currentChunk += image.ToReference()

			// If the current chunk is too large, finalize it
			if len(currentChunk) > charLimit {
				chunks = append(chunks, currentChunk)
				currentChunk = ""
			}

			continue
		}

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

		if markdownData[i].GetType() == mdtojson.NodeTypeParagraph {
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
