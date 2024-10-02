package mdtojson

import (
	"encoding/json"
	"fmt"

	ordered "github.com/stencilframe/mdtools/libs/ordered_map"
)

const (
	// Node types
	NodeTypeHeading       = "heading"
	NodeTypeText          = "text"
	NodeTypeTable         = "table"
	NodeTypeLink          = "link"
	NodeTypeImage         = "image"
	NodeTypeCode          = "code"
	NodeTypeCodeBlock     = "codeblock"
	NodeTypeParagraph     = "paragraph"
	NodeTypeList          = "list"
	NodeTypeListItem      = "listitem"
	NodeTypeBlockquote    = "blockquote"
	NodeTypeLineBreak     = "linebreak"
	NodeTypeSoftBreak     = "softbreak"
	NodeTypeHTMLBlock     = "htmlblock"
	NodeTypeHTMLSpan      = "htmlspan"
	NodeTypeLineSeparator = "lineseparator"
)

type (
	Node interface {
		ToMarkdown() string
		GetType() string
		GetChildren() []Node
		SetChildren([]Node)
	}

	// BaseNode represents a parsed Markdown element
	BaseNode struct {
		Type     string `json:"type"`
		Children []Node `json:"content,omitempty"` // Content of the node
	}

	// TextNode represents a parsed text element
	TextNode struct {
		BaseNode

		Text string `json:"text"`
	}

	// HeadingNode represents a parsed heading element
	HeadingNode struct {
		BaseNode

		Title string `json:"title"`
		Level int    `json:"level"`
	}

	// TableNode represents a parsed table element
	TableNode struct {
		BaseNode

		Data interface{} `json:"data"`
	}

	// LinkNode represents a parsed link element
	LinkNode struct {
		BaseNode

		URL   string `json:"url"`
		Title string `json:"title"`
	}

	// ImageNode represents a parsed image element
	ImageNode struct {
		BaseNode

		URL string `json:"url"`
		Alt string `json:"alt"`
	}

	// CodeNode represents a parsed code element
	CodeNode struct {
		BaseNode

		Code string `json:"code"`
	}

	// CodeBlockNode represents a parsed code block element
	CodeBlockNode struct {
		BaseNode

		Language string `json:"language"`
		Code     string `json:"code"`
	}

	// ParagraphNode represents a parsed paragraph element
	// It has no additional fields, but is used to represent a paragraph
	ParagraphNode BaseNode
)

// --- BaseNode methods ---

func NewBaseNode(t string, content []Node) Node {
	return &BaseNode{
		Type:     t,
		Children: content,
	}
}

func (n *BaseNode) GetType() string {
	return n.Type
}

func (n *BaseNode) GetChildren() []Node {
	return n.Children
}

func (n *BaseNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *BaseNode) ToMarkdown() string {
	// TODO: Implement this
	return ""
}

// --- HeadingNode methods ---

func NewHeadingNode(level int, title string) Node {
	return &HeadingNode{
		BaseNode: BaseNode{
			Type: NodeTypeHeading,
		},
		Title: title,
		Level: level,
	}
}

func (n *HeadingNode) GetType() string {
	return n.BaseNode.Type
}

func (n *HeadingNode) GetChildren() []Node {
	return n.BaseNode.Children
}

func (n *HeadingNode) SetChildren(children []Node) {
	n.BaseNode.Children = children
}

func (n *HeadingNode) ToMarkdown() string {
	level := ""
	for i := 0; i < n.Level; i++ {
		level += "#"
	}
	return level + " " + n.Title + "\n\n"
}

// --- TextNode methods ---

func NewTextNode(text string) Node {
	return &TextNode{
		BaseNode: BaseNode{
			Type: NodeTypeText,
		},
		Text: text,
	}
}

func (n *TextNode) GetType() string {
	return n.BaseNode.Type
}

func (n *TextNode) GetChildren() []Node {
	return n.BaseNode.Children
}

func (n *TextNode) SetChildren(children []Node) {
	n.BaseNode.Children = children
}

func (n *TextNode) ToMarkdown() string {
	return n.Text
}

// --- TableNode methods ---

func NewTableNode(data interface{}) Node {
	return &TableNode{
		BaseNode: BaseNode{
			Type: NodeTypeTable,
		},
		Data: data,
	}
}

func (n *TableNode) GetType() string {
	return n.BaseNode.Type
}

func (n *TableNode) GetChildren() []Node {
	return n.BaseNode.Children
}

func (n *TableNode) SetChildren(children []Node) {
	n.BaseNode.Children = children
}

func (n *TableNode) ToMarkdown() string {
	j, _ := json.Marshal(n.Data)
	return n.toJSONTable(string(j))
}

func (n *TableNode) toJSONTable(str string) string {
	return ":::json_table\n" + str + "\n:::\n\n"
}

// Split the table into chunks
func (n *TableNode) ChunkTable(firstChunkLimit, nextChunksLimit int) []string {
	chunks := []string{}
	chunk := ""
	limit := firstChunkLimit

	switch data := n.Data.(type) {
	case []*ordered.OrderedMap:
		for _, row := range data {
			j, _ := json.Marshal(row)
			part := string(j) + ",\n"
			if len(chunk)+len(part) > limit {
				chunks = append(chunks, n.toJSONTable("[\n"+chunk+"]"))
				chunk = ""
				limit = nextChunksLimit
			}
			chunk += part
		}
	case *ordered.OrderedMap:
		for key, r := range data.KVIter() {
			switch row := r.(type) {
			case *ordered.OrderedMap:
				j, _ := json.Marshal(row)
				part := fmt.Sprintf("%q: %s,\n", key, string(j))
				if len(chunk)+len(part) > limit {
					chunks = append(chunks, n.toJSONTable("{\n"+chunk+"}"))
					chunk = ""
					limit = nextChunksLimit
				}
				chunk += part
			}
		}
	}

	if len(chunk) > 0 {
		switch n.Data.(type) {
		case []*ordered.OrderedMap:
			chunks = append(chunks, n.toJSONTable("[\n"+chunk+"]"))
		case *ordered.OrderedMap:
			chunks = append(chunks, n.toJSONTable("{\n"+chunk+"}"))
		}
	}

	return chunks
}

// --- LinkNode methods ---

func NewLinkNode(url, title string) Node {
	return &LinkNode{
		BaseNode: BaseNode{
			Type: NodeTypeLink,
		},
		URL:   url,
		Title: title,
	}
}

func (n *LinkNode) GetType() string {
	return n.BaseNode.Type
}

func (n *LinkNode) GetChildren() []Node {
	return n.BaseNode.Children
}

func (n *LinkNode) SetChildren(children []Node) {
	n.BaseNode.Children = children
}

func (n *LinkNode) ToMarkdown() string {
	return "[" + n.Title + "](" + n.URL + ")\n\n"
}

// --- ImageNode methods ---

func NewImageNode(url, alt string) Node {
	return &ImageNode{
		BaseNode: BaseNode{
			Type: NodeTypeImage,
		},
		URL: url,
		Alt: alt,
	}
}

func (n *ImageNode) GetType() string {
	return n.BaseNode.Type
}

func (n *ImageNode) GetChildren() []Node {
	return n.BaseNode.Children
}

func (n *ImageNode) SetChildren(children []Node) {
	n.BaseNode.Children = children
}

func (n *ImageNode) ToMarkdown() string {
	return "![Image](" + n.URL + ")\n"
}

// --- CodeNode methods ---

func NewCodeNode(code string) Node {
	return &CodeNode{
		BaseNode: BaseNode{
			Type: NodeTypeCode,
		},
		Code: code,
	}
}

func (n *CodeNode) GetType() string {
	return n.BaseNode.Type
}

func (n *CodeNode) GetChildren() []Node {
	return n.BaseNode.Children
}

func (n *CodeNode) SetChildren(children []Node) {
	n.BaseNode.Children = children
}

func (n *CodeNode) ToMarkdown() string {
	return "`" + n.Code + "`"
}

// --- CodeBlockNode methods ---

func NewCodeBlockNode(language, code string) Node {
	return &CodeBlockNode{
		BaseNode: BaseNode{
			Type: NodeTypeCodeBlock,
		},
		Language: language,
		Code:     code,
	}
}

func (n *CodeBlockNode) GetType() string {
	return n.BaseNode.Type
}

func (n *CodeBlockNode) GetChildren() []Node {
	return n.BaseNode.Children
}

func (n *CodeBlockNode) SetChildren(children []Node) {
	n.BaseNode.Children = children
}

func (n *CodeBlockNode) ToMarkdown() string {
	return "```" + n.Language + "\n" + n.Code + "\n```\n\n"
}

// --- ParagraphNode methods ---

func NewParagraphNode(children []Node) Node {
	return &ParagraphNode{
		Type:     NodeTypeParagraph,
		Children: children,
	}
}

func (n *ParagraphNode) GetType() string {
	return n.Type
}

func (n *ParagraphNode) GetChildren() []Node {
	return n.Children
}

func (n *ParagraphNode) SetChildren(children []Node) {
	n.Children = children
}

func (n *ParagraphNode) ToMarkdown() string {
	return ""
}
