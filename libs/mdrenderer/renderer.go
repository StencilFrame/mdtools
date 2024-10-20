package mdrenderer

import (
	"io"
	"log"
	"strconv"
	"strings"

	bf "github.com/russross/blackfriday/v2"
)

// Option defines the functional option type
type Option func(r *Renderer)

// NewRenderer will return a new renderer with sane defaults
func NewRenderer(options ...Option) *Renderer {
	r := &Renderer{}
	for _, option := range options {
		option(r)
	}
	return r
}

// Renderer is a custom Blackfriday renderer
type Renderer struct {
	paragraphDecoration []byte
	nestedListLevel     int
	orderedListCounters []int
	tableAlignment      []bf.CellAlignFlags
	inTableHeader       bool
	tableCellCounter    int
	indentLevel         int // New field for indentation level
}

// skipParagraphNewline returns true if the paragraph should not have an empty line after it
func skipParagraphNewline(node *bf.Node) bool {
	parent := node.Parent
	if parent != nil && parent.Type == bf.BlockQuote {
		return true
	}

	grandparent := node.Parent.Parent
	if grandparent == nil || grandparent.Type != bf.List {
		return false
	}

	return grandparent.Type == bf.List && grandparent.Tight
}

// returns the current indentation based on the nesting level
func (r *Renderer) currentIndentation() []byte {
	// Indentation is 4 spaces
	return []byte(strings.Repeat("    ", r.indentLevel))
}

// RenderNode satisfies the Renderer interface
func (r *Renderer) RenderNode(w io.Writer, node *bf.Node, entering bool) bf.WalkStatus {
	switch node.Type {
	case bf.Document:
		// No action needed
	case bf.BlockQuote:
		if entering {
			w.Write(r.currentIndentation())
			r.paragraphDecoration = append(r.paragraphDecoration, '>', ' ')
		} else { // leaving
			r.paragraphDecoration = r.paragraphDecoration[:len(r.paragraphDecoration)-2]
			w.Write([]byte("\n"))
		}
	case bf.List:
		if entering {
			r.nestedListLevel++
			r.orderedListCounters = append(r.orderedListCounters, 0)
		} else { // leaving
			r.nestedListLevel--
			r.orderedListCounters = r.orderedListCounters[:len(r.orderedListCounters)-1]
			if r.nestedListLevel == 0 {
				// Insert a newline after a top-level list
				w.Write([]byte("\n"))
			}
		}
	case bf.Item:
		if entering {
			w.Write(r.currentIndentation())
			r.indentLevel++
			if node.ListFlags&bf.ListTypeOrdered != 0 {
				r.orderedListCounters[len(r.orderedListCounters)-1]++
				counter := strconv.Itoa(r.orderedListCounters[len(r.orderedListCounters)-1])
				w.Write([]byte(counter))
				w.Write([]byte{node.ListData.Delimiter, ' '})
			} else {
				w.Write([]byte{node.ListData.BulletChar, ' '})
			}
		} else { // leaving
			r.indentLevel--
		}
	case bf.Paragraph:
		if entering {
			if node.Parent != nil && node.Parent.Type == bf.Item {
				if node.Prev != nil {
					w.Write(r.currentIndentation())
				}
			} else {
				w.Write(r.paragraphDecoration)
			}
		} else { // leaving
			w.Write([]byte("\n"))
			if !skipParagraphNewline(node) {
				w.Write([]byte("\n"))
			}
		}
	case bf.Heading:
		if entering {
			w.Write([]byte(strings.Repeat("#", node.Level) + " "))
		} else { // leaving
			w.Write([]byte("\n\n"))
		}
	case bf.HorizontalRule:
		w.Write([]byte("---\n\n"))
	case bf.Emph:
		w.Write([]byte("*"))
	case bf.Strong:
		w.Write([]byte("**"))
	case bf.Del:
		w.Write([]byte("~~"))
	case bf.Link:
		if entering {
			w.Write([]byte("["))
		} else { // leaving
			w.Write([]byte("]("))
			w.Write(node.LinkData.Destination)
			if len(node.LinkData.Title) > 0 {
				w.Write([]byte(""))
				w.Write(node.LinkData.Title)
				w.Write([]byte(`"`))
			}
			w.Write([]byte(")"))
		}
	case bf.Image:
		if entering {
			w.Write([]byte("!["))
		} else { // leaving
			w.Write([]byte("]("))
			w.Write(node.LinkData.Destination)
			if len(node.LinkData.Title) > 0 {
				w.Write([]byte(""))
				w.Write(node.LinkData.Title)
				w.Write([]byte(`"`))
			}
			w.Write([]byte(")"))
		}
	case bf.Code:
		w.Write([]byte("`"))
		w.Write(node.Literal)
		w.Write([]byte("`"))
	case bf.Text:
		w.Write(node.Literal)
	case bf.CodeBlock:
		w.Write(r.currentIndentation())
		w.Write([]byte("```"))
		w.Write(node.CodeBlockData.Info)
		w.Write([]byte("\n"))
		w.Write(node.Literal)
		w.Write([]byte("```\n\n"))
	case bf.Softbreak:
		w.Write([]byte("\n"))
	case bf.Hardbreak:
		w.Write([]byte("  \n"))
	case bf.HTMLBlock, bf.HTMLSpan:
		w.Write(node.Literal)
	case bf.Table:
		if entering {
			// No action needed
		} else { // leaving
			w.Write([]byte("\n"))
		}
	case bf.TableHead:
		if entering {
			r.inTableHeader = true
			r.tableAlignment = []bf.CellAlignFlags{}
		} else { // leaving
			w.Write([]byte("|"))
			r.inTableHeader = false
			// Write alignment row
			for i, align := range r.tableAlignment {
				if i > 0 {
					w.Write([]byte(" |"))
				}
				switch align {
				case bf.TableAlignmentLeft:
					w.Write([]byte(" :---"))
				case bf.TableAlignmentRight:
					w.Write([]byte(" ---:"))
				case bf.TableAlignmentCenter:
					w.Write([]byte(" :---:"))
				default:
					w.Write([]byte(" ---"))
				}
			}
			w.Write([]byte(" |"))
			w.Write([]byte("\n"))
		}
	case bf.TableBody:
		// No action needed
	case bf.TableRow:
		if entering {
			r.tableCellCounter = 0
			w.Write([]byte("| "))
		} else { // leaving
			w.Write([]byte(" |"))
			w.Write([]byte("\n"))
		}
	case bf.TableCell:
		if entering {
			if r.tableCellCounter > 0 {
				w.Write([]byte(" | "))
			}
			r.tableCellCounter++
			if r.inTableHeader {
				// Collect alignment info
				r.tableAlignment = append(r.tableAlignment, node.TableCellData.Align)
			}
		}
	default:
		log.Printf("Unknown node type: %s\n", node.Type)
	}
	return bf.GoToNext
}

// RenderHeader satisfies the Renderer interface
func (r *Renderer) RenderHeader(w io.Writer, ast *bf.Node) {
	// No action needed
}

// RenderFooter satisfies the Renderer interface
func (r *Renderer) RenderFooter(w io.Writer, ast *bf.Node) {
	// No action needed
}
