package mdtojson

import (
	"bytes"
	"encoding/json"
	"io"
	"log"

	"github.com/russross/blackfriday/v2"
	ordered "github.com/stencilframe/mdtools/libs/ordered_map"
)

type (
	// Custom JSON Renderer
	JSONRenderer struct {
		nodes         []Node         // Root-level nodes
		headerStack   []*HeadingNode // Stack to manage nested headers
		currentHeader *HeadingNode   // Current header node
		imageRefs     []Node         // Stores image references (e.g., [1]: <image>, [2]: <image>)
	}
)

// NewJSONRenderer creates a new JSONRenderer instance
func NewJSONRenderer() *JSONRenderer {
	return &JSONRenderer{
		imageRefs: []Node{},
	}
}

// RenderNode processes each node and converts it to a JSON-friendly structure
func (r *JSONRenderer) RenderNode(w io.Writer, node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
	if entering {
		var contentNode Node
		switch node.Type {
		case blackfriday.Document:
			// Document is the root node, no specific action needed other than ensuring the document is parsed
			return blackfriday.GoToNext

		case blackfriday.Heading:
			r.handleHeader(node)

		case blackfriday.Table:
			contentNode = r.handleTable(node)

		case blackfriday.List:
			contentNode = r.handleList(node)

		case blackfriday.Paragraph:
			contentNode = r.handleParagraph(node)

		case blackfriday.Hardbreak:
			contentNode = &BaseNode{
				Type: NodeTypeLineBreak,
			}

		case blackfriday.Softbreak:
			contentNode = &BaseNode{
				Type: NodeTypeSoftBreak,
			}

		case blackfriday.HorizontalRule:
			contentNode = &BaseNode{
				Type: NodeTypeLineSeparator,
			}

		case blackfriday.BlockQuote:
			quoteContent := r.extractContent(node)
			contentNode = &BaseNode{
				Type:     NodeTypeBlockquote,
				Children: quoteContent,
			}

		case blackfriday.CodeBlock:
			codeContent := string(node.Literal)
			language := string(node.Info)
			contentNode = NewCodeBlockNode(codeContent, language)

			// TODO: Implement HTML block and span handling
			// case blackfriday.HTMLBlock:
			// 	htmlContent := string(node.Literal)
			// 	contentNode = &BaseNode{
			// 		Type:     "html-block",
			// 		Children: htmlContent,
			// 	}

			// case blackfriday.HTMLSpan:
			// 	htmlContent := string(node.Literal)
			// 	contentNode = &BaseNode{
			// 		Type:     "html-span",
			// 		Children: htmlContent,
			// 	}
		}

		if contentNode != nil {
			if r.currentHeader != nil {
				r.currentHeader.SetChildren(append(r.currentHeader.GetChildren(), contentNode))
			} else {
				r.nodes = append(r.nodes, contentNode)
			}
		}

		if node.Type == blackfriday.BlockQuote ||
			node.Type == blackfriday.Table ||
			node.Type == blackfriday.List {
			return blackfriday.SkipChildren
		}
	}
	return blackfriday.GoToNext
}

// RenderHeader is a no-op for this extension, but required by the Renderer interface
func (r *JSONRenderer) RenderHeader(w io.Writer, ast *blackfriday.Node) {}

// RenderFooter is called at the end of processing to finalize the output
func (r *JSONRenderer) RenderFooter(w io.Writer, ast *blackfriday.Node) {
	// Finalize and append any remaining headers to the root node
	r.finalizeHeaders(0)

	// Output the final JSON result
	output, err := json.MarshalIndent(r.nodes, "", "  ")
	if err != nil {
		log.Fatalf("Error generating JSON: %v", err)
	}
	w.Write(output)
}

// Return nodes
func (r *JSONRenderer) GetNodes() []Node {
	// Finalize and append any remaining headers to the root node
	r.finalizeHeaders(0)

	// Return the root nodes
	return r.nodes
}

// handleHeader manages the heading elements and finalizes them.
func (r *JSONRenderer) handleHeader(node *blackfriday.Node) {
	level := node.HeadingData.Level
	headerText := extractText(node) // Extract heading text
	headerNode := NewHeadingNode(level, headerText).(*HeadingNode)

	// Finalize and append any remaining headers
	r.finalizeHeaders(level)

	// Push the new header to the header stack
	r.headerStack = append(r.headerStack, headerNode)

	// Set the new header as the currentNode
	r.currentHeader = headerNode
}

// finalizeHeaders handles appending all headers to `r.nodes`
func (r *JSONRenderer) finalizeHeaders(level int) {
	// If there's an existing header being processed, finalize and store it in the stack
	if r.currentHeader != nil {
		finishedHeaders := []Node{}
		// Pop headers from the stack if the new header has a lower or equal level
		for len(r.headerStack) > 0 && level <= r.headerStack[len(r.headerStack)-1].Level {
			finishedHeader := r.headerStack[len(r.headerStack)-1]
			r.headerStack = r.headerStack[:len(r.headerStack)-1]

			// Append the finished header as content to the parent header or root
			if len(r.headerStack) > 0 {
				parentHeader := r.headerStack[len(r.headerStack)-1]
				parentHeader.SetChildren(append(parentHeader.GetChildren(), finishedHeader))
			} else {
				finishedHeaders = append(finishedHeaders, finishedHeader)
			}
		}

		// Append the finished headers to the last header in the stack
		if len(r.headerStack) > 0 {
			parent := r.headerStack[len(r.headerStack)-1]
			for i := 0; i < len(finishedHeaders); i++ {
				parent.SetChildren(append(parent.GetChildren(), finishedHeaders[i]))
			}
		} else {
			// Append the finished headers to the root nodes
			r.nodes = append(r.nodes, finishedHeaders...)
		}
	}
}

// extractText extracts plain text from a node
func extractText(node *blackfriday.Node) string {
	var buffer bytes.Buffer
	node.Walk(func(n *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if entering && n.Type == blackfriday.Text {
			buffer.Write(n.Literal)
		}
		return blackfriday.GoToNext
	})
	return buffer.String()
}

// extractContent handles text nodes, links, images, and inline elements.
func (r *JSONRenderer) extractContent(node *blackfriday.Node) []Node {
	children := []Node{}

	node.Walk(func(n *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if entering {
			switch n.Type {
			case blackfriday.Text:
				content := extractText(n)
				if content == "" {
					return blackfriday.GoToNext
				}
				item := NewTextNode(content)
				children = append(children, item)
			case blackfriday.List:
				return blackfriday.SkipChildren
			case blackfriday.Link:
				linkUrl := string(n.LinkData.Destination)
				link := NewLinkNode(linkUrl, extractText(n))
				children = append(children, link)
				return blackfriday.SkipChildren
			case blackfriday.Image:
				image := NewImageNode(string(n.LinkData.Destination), extractText(n))
				ref := r.addImage(image)
				// Update the image reference
				image.(*ImageNode).Reference = ref
				children = append(children, image)
				return blackfriday.SkipChildren
			case blackfriday.Code:
				codeContent := string(n.Literal)
				code := NewCodeNode(codeContent)
				children = append(children, code)
			case blackfriday.CodeBlock:
				codeContent := string(n.Literal)
				language := string(n.CodeBlockData.Info)
				codeBlock := NewCodeBlockNode(language, codeContent)
				children = append(children, codeBlock)
			case blackfriday.BlockQuote:
				content := extractText(n)
				item := NewTextNode(content)
				children = append(children, item)
				return blackfriday.SkipChildren
			}
		}
		return blackfriday.GoToNext
	})

	return children
}

// handleParagraph processes paragraph nodes and extracts text content
func (r *JSONRenderer) handleParagraph(node *blackfriday.Node) Node {
	children := r.extractContent(node)

	return NewParagraphNode(children)
}

// handleList processes list nodes and extracts list items
func (r *JSONRenderer) handleList(node *blackfriday.Node) Node {
	var listItems []Node
	node.Walk(func(n *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if entering && n.Type == blackfriday.Item {
			listItem := r.extractListItems(n)
			listItems = append(listItems, listItem)
			return blackfriday.SkipChildren
		}
		return blackfriday.GoToNext
	})
	return &BaseNode{
		Type:     NodeTypeList,
		Children: listItems,
	}
}

// extractListItems extracts list items from a list node
func (r *JSONRenderer) extractListItems(node *blackfriday.Node) Node {
	children := []Node{}

	node.Walk(func(n *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if entering {
			switch n.Type {
			case blackfriday.List:
				list := r.handleList(n)
				children = append(children, list)
				return blackfriday.SkipChildren
			case blackfriday.Item:
				listItem := r.handleParagraph(n)
				children = append(children, listItem)
			}
		}
		return blackfriday.GoToNext
	})

	return &BaseNode{
		Type:     NodeTypeListItem,
		Children: children,
	}
}

// handleTable processes table nodes and extracts rows and cells
func (r *JSONRenderer) handleTable(node *blackfriday.Node) Node {
	var tableData interface{}
	var headers []string

	node.Walk(func(n *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		switch n.Type {
		case blackfriday.TableHead:
			headers = r.collectTableHeaders(n)
		case blackfriday.TableBody:
			if headers[0] == "" {
				tableData = r.collectTableRowsWithKeys(headers, n)
			} else {
				tableData = r.collectTableRowsRegular(headers, n)
			}
		}
		return blackfriday.GoToNext
	})
	return NewTableNode(tableData)
}

// collectTableHeaders collects the headers from the table's TableHead node
func (r *JSONRenderer) collectTableHeaders(node *blackfriday.Node) []string {
	var headers []string
	node.Walk(func(n *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if entering && n.Type == blackfriday.TableCell {
			headers = append(headers, extractText(n))
		}
		return blackfriday.GoToNext
	})
	return headers
}

// collectTableRowsRegular collects the rows from a table's TableBody node
func (r *JSONRenderer) collectTableRowsRegular(headers []string, node *blackfriday.Node) []*ordered.OrderedMap {
	var tableData []*ordered.OrderedMap
	node.Walk(func(n *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if entering && n.Type == blackfriday.TableRow {
			currentRow := r.collectRowCells(headers, n)
			tableData = append(tableData, currentRow)
		}
		return blackfriday.GoToNext
	})

	for i := range tableData {
		for k, v := range tableData[i].KVIter() {
			if k == "" && v == "" {
				tableData[i].Delete(k)
			}
		}
	}

	return tableData
}

// collectRowCells collects the cells from a table row node
func (r *JSONRenderer) collectRowCells(headers []string, node *blackfriday.Node) *ordered.OrderedMap {
	rowData := ordered.NewOrderedMap()
	headerIndex := 0
	node.Walk(func(n *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if entering && n.Type == blackfriday.TableCell {
			if headerIndex < len(headers) {
				rowData.Set(headers[headerIndex], extractText(n))
				headerIndex++
			}
		}
		return blackfriday.GoToNext
	})
	return rowData
}

// collectTableRowsWithKeys collects the rows from a table's TableBody node
func (r *JSONRenderer) collectTableRowsWithKeys(headers []string, node *blackfriday.Node) *ordered.OrderedMap {
	tableData := ordered.NewOrderedMap()
	node.Walk(func(n *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if entering && n.Type == blackfriday.TableRow {
			key, currentRow := r.collectRowCellsWithKeys(headers, n)
			tableData.Set(key, currentRow)
		}
		return blackfriday.GoToNext
	})
	return tableData
}

// collectRowCellsWithKeys collects the cells from a table row node
func (r *JSONRenderer) collectRowCellsWithKeys(headers []string, node *blackfriday.Node) (key string, rowData *ordered.OrderedMap) {
	firstCell := false
	rowData = r.collectRowCells(headers, node)
	node.Walk(func(n *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if entering && n.Type == blackfriday.TableCell {
			if !firstCell {
				key = extractText(n)
				firstCell = true
				rowData.Delete(headers[0])
				return blackfriday.Terminate
			}
		}
		return blackfriday.GoToNext
	})
	return key, rowData
}

// addImage adds an image reference to the imageRefs map
func (r *JSONRenderer) addImage(image Node) int {
	// Check if the image reference already exists
	for i, ref := range r.imageRefs {
		if ref.(*ImageNode).URL == image.(*ImageNode).URL {
			return i + 1
		}
	}

	// Add new image reference
	r.imageRefs = append(r.imageRefs, image)
	return len(r.imageRefs)
}

// GetImageURLs returns the image references
func (r *JSONRenderer) GetImageURLs() map[int]string {
	images := make(map[int]string, len(r.imageRefs))
	for i, image := range r.imageRefs {
		images[i+1] = image.(*ImageNode).URL
	}
	return images
}
