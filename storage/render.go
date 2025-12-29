package storage

import (
	"fmt"
	"html"
	"strings"
)

// Render converts a Page to Confluence Storage XHTML.
func Render(page *Page) (string, error) {
	if page == nil {
		return "", nil
	}
	var buf strings.Builder
	for _, block := range page.Blocks {
		s, err := RenderBlock(block)
		if err != nil {
			return "", err
		}
		buf.WriteString(s)
	}
	return buf.String(), nil
}

// RenderBlock converts a single Block to Storage XHTML.
func RenderBlock(block Block) (string, error) {
	switch b := block.(type) {
	case *Table:
		return renderTable(b)
	case Table:
		return renderTable(&b)
	case *Paragraph:
		return renderParagraph(b)
	case Paragraph:
		return renderParagraph(&b)
	case *Heading:
		return renderHeading(b)
	case Heading:
		return renderHeading(&b)
	case *Macro:
		return RenderMacro(b)
	case Macro:
		return RenderMacro(&b)
	case *BulletList:
		return renderBulletList(b)
	case BulletList:
		return renderBulletList(&b)
	case *NumberedList:
		return renderNumberedList(b)
	case NumberedList:
		return renderNumberedList(&b)
	case *CodeBlock:
		return renderCodeBlock(b)
	case CodeBlock:
		return renderCodeBlock(&b)
	case *HorizontalRule:
		return "<hr/>", nil
	case HorizontalRule:
		return "<hr/>", nil
	default:
		return "", fmt.Errorf("unsupported block type: %T", block)
	}
}

func renderTable(t *Table) (string, error) {
	if t == nil {
		return "", nil
	}
	var buf strings.Builder
	buf.WriteString("<table><tbody>")

	// Header row
	if len(t.Headers) > 0 {
		buf.WriteString("<tr>")
		for _, h := range t.Headers {
			buf.WriteString("<th>")
			buf.WriteString(html.EscapeString(h))
			buf.WriteString("</th>")
		}
		buf.WriteString("</tr>")
	}

	// Data rows
	for _, row := range t.Rows {
		buf.WriteString("<tr>")
		for _, cell := range row.Cells {
			buf.WriteString("<td>")
			buf.WriteString(renderCell(&cell))
			buf.WriteString("</td>")
		}
		buf.WriteString("</tr>")
	}

	buf.WriteString("</tbody></table>")
	return buf.String(), nil
}

func renderCell(c *Cell) string {
	if c == nil {
		return ""
	}
	if c.Macro != nil {
		s, _ := RenderMacro(c.Macro)
		return s
	}
	return html.EscapeString(c.Text)
}

// RenderMacro converts a Macro to Storage XHTML.
func RenderMacro(m *Macro) (string, error) {
	if m == nil {
		return "", nil
	}
	var buf strings.Builder
	buf.WriteString(`<ac:structured-macro ac:name="`)
	buf.WriteString(html.EscapeString(m.Name))
	buf.WriteString(`">`)

	for key, val := range m.Params {
		buf.WriteString(`<ac:parameter ac:name="`)
		buf.WriteString(html.EscapeString(key))
		buf.WriteString(`">`)
		buf.WriteString(html.EscapeString(val))
		buf.WriteString(`</ac:parameter>`)
	}

	if m.Body != "" {
		buf.WriteString(`<ac:rich-text-body>`)
		buf.WriteString(m.Body) // Body is assumed to be valid Storage XHTML
		buf.WriteString(`</ac:rich-text-body>`)
	}

	buf.WriteString(`</ac:structured-macro>`)
	return buf.String(), nil
}

func renderParagraph(p *Paragraph) (string, error) {
	if p == nil {
		return "", nil
	}
	return "<p>" + html.EscapeString(p.Text) + "</p>", nil
}

func renderHeading(h *Heading) (string, error) {
	if h == nil {
		return "", nil
	}
	if h.Level < 1 || h.Level > 6 {
		return "", fmt.Errorf("invalid heading level: %d", h.Level)
	}
	return fmt.Sprintf("<h%d>%s</h%d>", h.Level, html.EscapeString(h.Text), h.Level), nil
}

func renderBulletList(bl *BulletList) (string, error) {
	if bl == nil {
		return "", nil
	}
	var buf strings.Builder
	buf.WriteString("<ul>")
	for _, item := range bl.Items {
		buf.WriteString("<li>")
		buf.WriteString(html.EscapeString(item.Text))
		buf.WriteString("</li>")
	}
	buf.WriteString("</ul>")
	return buf.String(), nil
}

func renderNumberedList(nl *NumberedList) (string, error) {
	if nl == nil {
		return "", nil
	}
	var buf strings.Builder
	buf.WriteString("<ol>")
	for _, item := range nl.Items {
		buf.WriteString("<li>")
		buf.WriteString(html.EscapeString(item.Text))
		buf.WriteString("</li>")
	}
	buf.WriteString("</ol>")
	return buf.String(), nil
}

func renderCodeBlock(cb *CodeBlock) (string, error) {
	if cb == nil {
		return "", nil
	}
	var buf strings.Builder
	buf.WriteString(`<ac:structured-macro ac:name="code">`)
	if cb.Language != "" {
		buf.WriteString(`<ac:parameter ac:name="language">`)
		buf.WriteString(html.EscapeString(cb.Language))
		buf.WriteString(`</ac:parameter>`)
	}
	buf.WriteString(`<ac:plain-text-body><![CDATA[`)
	buf.WriteString(cb.Code)
	buf.WriteString(`]]></ac:plain-text-body>`)
	buf.WriteString(`</ac:structured-macro>`)
	return buf.String(), nil
}
