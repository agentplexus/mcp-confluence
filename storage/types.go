// Package storage provides types and utilities for working with Confluence Storage Format (XHTML).
// It includes an intermediate representation (IR) for content blocks, rendering to Storage XHTML,
// parsing from Storage XHTML, and validation.
package storage

// Block represents any content block in Confluence Storage Format.
type Block interface {
	// BlockType returns the type identifier (e.g., "table", "paragraph").
	BlockType() string
}

// Page represents a full page's content as a sequence of blocks.
type Page struct {
	Blocks []Block `json:"blocks"`
}

// Table represents a Confluence table.
type Table struct {
	Headers []string `json:"headers"`
	Rows    []Row    `json:"rows"`
}

// BlockType implements Block.
func (Table) BlockType() string { return "table" }

// Row represents a table row.
type Row struct {
	Cells []Cell `json:"cells"`
}

// Cell represents a table cell which can contain text or a macro.
type Cell struct {
	Text  string `json:"text,omitempty"`
	Macro *Macro `json:"macro,omitempty"`
}

// Macro represents a Confluence macro (ac:structured-macro).
type Macro struct {
	Name   string            `json:"name"`
	Params map[string]string `json:"params,omitempty"`
	Body   string            `json:"body,omitempty"`
}

// BlockType implements Block.
func (Macro) BlockType() string { return "macro" }

// Paragraph represents a simple text paragraph.
type Paragraph struct {
	Text string `json:"text"`
}

// BlockType implements Block.
func (Paragraph) BlockType() string { return "paragraph" }

// Heading represents a heading (h1-h6).
type Heading struct {
	Level int    `json:"level"` // 1-6
	Text  string `json:"text"`
}

// BlockType implements Block.
func (Heading) BlockType() string { return "heading" }

// BulletList represents an unordered list.
type BulletList struct {
	Items []ListItem `json:"items"`
}

// BlockType implements Block.
func (BulletList) BlockType() string { return "bullet_list" }

// NumberedList represents an ordered list.
type NumberedList struct {
	Items []ListItem `json:"items"`
}

// BlockType implements Block.
func (NumberedList) BlockType() string { return "numbered_list" }

// ListItem represents a list item.
type ListItem struct {
	Text string `json:"text"`
}

// CodeBlock represents a code block with optional language.
type CodeBlock struct {
	Language string `json:"language,omitempty"`
	Code     string `json:"code"`
}

// BlockType implements Block.
func (CodeBlock) BlockType() string { return "code_block" }

// HorizontalRule represents a horizontal rule (<hr/>).
type HorizontalRule struct{}

// BlockType implements Block.
func (HorizontalRule) BlockType() string { return "horizontal_rule" }
