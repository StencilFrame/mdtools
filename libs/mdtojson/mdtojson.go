package mdtojson

import (
	"bytes"
	"encoding/json"
	"io"
	"log"

	"github.com/russross/blackfriday/v2"
)

type (
	// Node represents a parsed Markdown element
	Node struct {
		Type    string      `json:"type"`
		Title   string      `json:"title,omitempty"`   // Title of the header
		Level   int         `json:"level,omitempty"`   // Level of the header
		Content interface{} `json:"content,omitempty"` // Content of the node
	}

	// Custom JSON Renderer
	JSONRenderer struct {
		nodes         []*Node           // Root-level nodes
		headerStack   []*Node           // Stack to manage nested headers
		currentHeader *Node             // Current header node
		imageRefs     map[string]string // Stores image references (e.g., [image1]: <url>)
	}
)

// NewJSONRenderer creates a new JSONRenderer instance
func NewJSONRenderer() *JSONRenderer {
	return &JSONRenderer{
		imageRefs: make(map[string]string),
	}
}

// RenderNode processes each node and converts it to a JSON-friendly structure
func (r *JSONRenderer) RenderNode(w io.Writer, node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
	if entering {
		var contentNode *Node
		switch node.Type {
		case blackfriday.Document:
			// Document is the root node, no specific action needed other than ensuring the document is parsed
			return blackfriday.GoToNext

		case blackfriday.Heading:
			r.handleHeader(node)

		case blackfriday.Table:
			contentNode = handleTable(node)
			if r.currentHeader != nil {
				r.currentHeader.Content = appendContent(r.currentHeader.Content, contentNode)
			} else {
				r.nodes = append(r.nodes, contentNode)
			}
			return blackfriday.SkipChildren

		case blackfriday.List:
			contentNode = handleList(node)
			if r.currentHeader != nil {
				r.currentHeader.Content = appendContent(r.currentHeader.Content, contentNode)
			} else {
				r.nodes = append(r.nodes, contentNode)
			}
			return blackfriday.SkipChildren

		case blackfriday.Paragraph:
			contentNode = extractContent(node)

		case blackfriday.Hardbreak:
			contentNode = &Node{
				Type: "line-break",
			}

		case blackfriday.Softbreak:
			contentNode = &Node{
				Type: "softbreak",
			}

		case blackfriday.HorizontalRule:
			contentNode = &Node{
				Type: "line-separator",
			}

		case blackfriday.BlockQuote:
			quoteContent := extractContent(node)
			contentNode = &Node{
				Type:    "blockquote",
				Content: quoteContent,
			}

		case blackfriday.CodeBlock:
			codeContent := string(node.Literal)
			language := string(node.Info)
			contentNode = &Node{
				Type: "code-block",
				Content: map[string]string{
					"code":     codeContent,
					"language": language,
				},
			}

		case blackfriday.HTMLBlock:
			htmlContent := string(node.Literal)
			contentNode = &Node{
				Type:    "html-block",
				Content: htmlContent,
			}

		case blackfriday.HTMLSpan:
			htmlContent := string(node.Literal)
			contentNode = &Node{
				Type:    "html-span",
				Content: htmlContent,
			}
		}

		if contentNode != nil {
			if r.currentHeader != nil {
				r.currentHeader.Content = appendContent(r.currentHeader.Content, contentNode)
			} else {
				r.nodes = append(r.nodes, contentNode)
			}
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
func (r *JSONRenderer) GetNodes() []*Node {
	// Finalize and append any remaining headers to the root node
	r.finalizeHeaders(0)

	// Return the root nodes
	return r.nodes
}

// handleHeader manages the heading elements and finalizes them.
func (r *JSONRenderer) handleHeader(node *blackfriday.Node) {
	level := node.HeadingData.Level
	headerText := extractText(node) // Extract heading text
	headerNode := &Node{
		Type:    "heading",
		Title:   headerText,      // The title of the header
		Level:   level,           // The level of the header
		Content: []interface{}{}, // Content under this header
	}

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
		finishedHeaders := []*Node{}
		// Pop headers from the stack if the new header has a lower or equal level
		for len(r.headerStack) > 0 && level <= r.headerStack[len(r.headerStack)-1].Level {
			finishedHeader := r.headerStack[len(r.headerStack)-1]
			r.headerStack = r.headerStack[:len(r.headerStack)-1]

			// Append the finished header as content to the parent header or root
			if len(r.headerStack) > 0 {
				parentHeader := r.headerStack[len(r.headerStack)-1]
				parentHeader.Content = appendContent(parentHeader.Content, finishedHeader)
			} else {
				finishedHeaders = append(finishedHeaders, finishedHeader)
			}
		}

		// Append the finished headers to the last header in the stack
		if len(r.headerStack) > 0 {
			parent := r.headerStack[len(r.headerStack)-1]
			for i := 0; i < len(finishedHeaders); i++ {
				parent.Content = appendContent(parent.Content, finishedHeaders[i])
			}
		} else {
			// Append the finished headers to the root nodes
			r.nodes = append(r.nodes, finishedHeaders...)
		}
	}
}

// appendContent appends new content to the current header's content array
func appendContent(existingContent interface{}, newContent interface{}) interface{} {
	switch content := existingContent.(type) {
	case []interface{}:
		return append(content, newContent)
	case *Node:
		return []interface{}{content, newContent}
	default:
		return newContent
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
func extractContent(node *blackfriday.Node) *Node {
	children := []interface{}{}

	node.Walk(func(n *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if entering {
			switch n.Type {
			case blackfriday.Text:
				content := extractText(n)
				if content == "" {
					return blackfriday.GoToNext
				}
				item := &Node{
					Type:    "text",
					Content: content,
				}
				children = append(children, item)
			case blackfriday.List:
				return blackfriday.SkipChildren
			case blackfriday.Link:
				linkUrl := string(n.LinkData.Destination)
				link := &Node{
					Type: "link",
					Content: map[string]string{
						"url":  linkUrl,
						"text": extractText(n),
					},
				}
				children = append(children, link)
				return blackfriday.SkipChildren
			case blackfriday.Image:
				image := &Node{
					Type: "image",
					Content: map[string]string{
						"url": string(n.LinkData.Destination),
						"alt": extractText(n),
					},
				}
				children = append(children, image)
				return blackfriday.SkipChildren
			case blackfriday.Code:
				codeContent := string(n.Literal)
				code := &Node{
					Type:    "code",
					Content: codeContent,
				}
				children = append(children, code)
			}
		}
		return blackfriday.GoToNext
	})

	return &Node{
		Type:    "paragraph",
		Content: children,
	}
}

// handleList processes list nodes and extracts list items
func handleList(node *blackfriday.Node) *Node {
	var listItems []interface{}
	node.Walk(func(n *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if entering && n.Type == blackfriday.Item {
			listItem := extractListItems(n)
			listItems = append(listItems, listItem)
			return blackfriday.SkipChildren
		}
		return blackfriday.GoToNext
	})
	return &Node{
		Type:    "list",
		Content: listItems,
	}
}

// extractListItems extracts list items from a list node
func extractListItems(node *blackfriday.Node) *Node {
	children := []interface{}{}

	node.Walk(func(n *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if entering {
			switch n.Type {
			case blackfriday.List:
				list := handleList(n)
				children = append(children, list)
				return blackfriday.SkipChildren
			case blackfriday.Item:
				listItem := extractContent(n)
				children = append(children, listItem)
			}
		}
		return blackfriday.GoToNext
	})

	return &Node{
		Type:    "list-item",
		Content: children,
	}
}

// handleTable processes table nodes and extracts rows and cells
func handleTable(node *blackfriday.Node) *Node {
	var tableData interface{}
	var headers []string

	node.Walk(func(n *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		switch n.Type {
		case blackfriday.TableHead:
			headers = collectTableHeaders(n)
		case blackfriday.TableBody:
			if headers[0] == "" {
				tableData = collectTableRowsWithKeys(headers, n)
			} else {
				tableData = collectTableRowsRegular(headers, n)
			}
		}
		return blackfriday.GoToNext
	})
	return &Node{
		Type:    "table",
		Content: tableData,
	}
}

// collectTableHeaders collects the headers from the table's TableHead node
func collectTableHeaders(node *blackfriday.Node) []string {
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
func collectTableRowsRegular(headers []string, node *blackfriday.Node) []map[string]string {
	var tableData []map[string]string
	node.Walk(func(n *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if entering && n.Type == blackfriday.TableRow {
			currentRow := collectRowCells(headers, n)
			tableData = append(tableData, currentRow)
		}
		return blackfriday.GoToNext
	})

	for i := range tableData {
		for k, v := range tableData[i] {
			if k == "" && v == "" {
				delete(tableData[i], k)
			}
		}
	}

	return tableData
}

// collectRowCells collects the cells from a table row node
func collectRowCells(headers []string, node *blackfriday.Node) map[string]string {
	rowData := make(map[string]string)
	node.Walk(func(n *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if entering && n.Type == blackfriday.TableCell {
			headerIndex := len(rowData)
			if headerIndex < len(headers) {
				rowData[headers[headerIndex]] = extractText(n)
			}
		}
		return blackfriday.GoToNext
	})
	return rowData
}

// collectTableRowsWithKeys collects the rows from a table's TableBody node
func collectTableRowsWithKeys(headers []string, node *blackfriday.Node) map[string]map[string]string {
	tableData := make(map[string]map[string]string)
	node.Walk(func(n *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if entering && n.Type == blackfriday.TableRow {
			key, currentRow := collectRowCellsWithKeys(headers, n)
			tableData[key] = currentRow
		}
		return blackfriday.GoToNext
	})
	return tableData
}

// collectRowCellsWithKeys collects the cells from a table row node
func collectRowCellsWithKeys(headers []string, node *blackfriday.Node) (key string, rowData map[string]string) {
	firstCell := false
	rowData = collectRowCells(headers, node)
	node.Walk(func(n *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if entering && n.Type == blackfriday.TableCell {
			if !firstCell {
				key = extractText(n)
				firstCell = true
				delete(rowData, headers[0])
				return blackfriday.Terminate
			}
		}
		return blackfriday.GoToNext
	})
	return key, rowData
}
