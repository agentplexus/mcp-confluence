package storage

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// Parse converts Confluence Storage XHTML to a Page with Blocks.
func Parse(xhtml string) (*Page, error) {
	if xhtml == "" {
		return &Page{}, nil
	}

	wrapped := "<root xmlns:ac=\"http://atlassian.com/confluence\" xmlns:ri=\"http://atlassian.com/confluence\">" + xhtml + "</root>"
	decoder := xml.NewDecoder(strings.NewReader(wrapped))
	decoder.Entity = htmlEntities

	page := &Page{}

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("parse error: %w", err)
		}

		if se, ok := tok.(xml.StartElement); ok {
			block, err := parseElement(decoder, se)
			if err != nil {
				return nil, err
			}
			if block != nil {
				page.Blocks = append(page.Blocks, block)
			}
		}
	}

	return page, nil
}

func parseElement(decoder *xml.Decoder, start xml.StartElement) (Block, error) {
	switch start.Name.Local {
	case "root":
		// Skip the wrapper element, continue parsing children
		return nil, nil
	case "table":
		return parseTable(decoder, start)
	case "p":
		return parseParagraph(decoder, start)
	case "h1", "h2", "h3", "h4", "h5", "h6":
		return parseHeading(decoder, start)
	case "ul":
		return parseBulletList(decoder, start)
	case "ol":
		return parseNumberedList(decoder, start)
	case "hr":
		return &HorizontalRule{}, nil
	case "structured-macro":
		return parseMacro(decoder, start)
	default:
		// Skip unknown elements
		if err := skipElement(decoder); err != nil {
			return nil, err
		}
		return nil, nil
	}
}

func parseTable(decoder *xml.Decoder, _ xml.StartElement) (*Table, error) {
	table := &Table{
		Headers: []string{},
		Rows:    []Row{},
	}

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "tbody":
				if err := parseTbody(decoder, table); err != nil {
					return nil, err
				}
			case "tr":
				// Table row outside tbody (shouldn't happen with valid storage format)
				row, isHeader, err := parseTableRow(decoder)
				if err != nil {
					return nil, err
				}
				if isHeader {
					for _, cell := range row.Cells {
						table.Headers = append(table.Headers, cell.Text)
					}
				} else {
					table.Rows = append(table.Rows, *row)
				}
			default:
				if err := skipElement(decoder); err != nil {
					return nil, err
				}
			}
		case xml.EndElement:
			if t.Name.Local == "table" {
				return table, nil
			}
		}
	}

	return table, nil
}

func parseTbody(decoder *xml.Decoder, table *Table) error {
	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "tr" {
				row, isHeader, err := parseTableRow(decoder)
				if err != nil {
					return err
				}
				if isHeader {
					for _, cell := range row.Cells {
						table.Headers = append(table.Headers, cell.Text)
					}
				} else {
					table.Rows = append(table.Rows, *row)
				}
			} else {
				if err := skipElement(decoder); err != nil {
					return err
				}
			}
		case xml.EndElement:
			if t.Name.Local == "tbody" {
				return nil
			}
		}
	}
	return nil
}

func parseTableRow(decoder *xml.Decoder) (*Row, bool, error) {
	row := &Row{Cells: []Cell{}}
	isHeader := false

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, false, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "th":
				isHeader = true
				cell, err := parseTableCell(decoder, "th")
				if err != nil {
					return nil, false, err
				}
				row.Cells = append(row.Cells, *cell)
			case "td":
				cell, err := parseTableCell(decoder, "td")
				if err != nil {
					return nil, false, err
				}
				row.Cells = append(row.Cells, *cell)
			default:
				if err := skipElement(decoder); err != nil {
					return nil, false, err
				}
			}
		case xml.EndElement:
			if t.Name.Local == "tr" {
				return row, isHeader, nil
			}
		}
	}

	return row, isHeader, nil
}

func parseTableCell(decoder *xml.Decoder, tagName string) (*Cell, error) {
	cell := &Cell{}
	var content strings.Builder

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "structured-macro" {
				macro, err := parseMacro(decoder, t)
				if err != nil {
					return nil, err
				}
				cell.Macro = macro
			} else {
				// Recursively extract text from nested elements (p, strong, ul, etc.)
				nested, err := extractNestedText(decoder, t.Name.Local)
				if err != nil {
					return nil, err
				}
				if content.Len() > 0 && len(nested) > 0 {
					content.WriteString(" ") // Add space between text from different elements
				}
				content.WriteString(nested)
			}
		case xml.CharData:
			content.Write(t)
		case xml.EndElement:
			if t.Name.Local == tagName {
				cell.Text = strings.TrimSpace(content.String())
				return cell, nil
			}
		}
	}

	return cell, nil
}

func parseParagraph(decoder *xml.Decoder, _ xml.StartElement) (*Paragraph, error) {
	var content strings.Builder
	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch t := tok.(type) {
		case xml.CharData:
			content.Write(t)
		case xml.EndElement:
			if t.Name.Local == "p" {
				return &Paragraph{Text: strings.TrimSpace(content.String())}, nil
			}
		case xml.StartElement:
			// Recursively extract text from nested elements (strong, code, a, etc.)
			nested, err := extractNestedText(decoder, t.Name.Local)
			if err != nil {
				return nil, err
			}
			content.WriteString(nested)
		}
	}
	return &Paragraph{Text: content.String()}, nil
}

func parseHeading(decoder *xml.Decoder, start xml.StartElement) (*Heading, error) {
	level := int(start.Name.Local[1] - '0')
	var content strings.Builder

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch t := tok.(type) {
		case xml.CharData:
			content.Write(t)
		case xml.EndElement:
			if strings.HasPrefix(t.Name.Local, "h") {
				return &Heading{Level: level, Text: strings.TrimSpace(content.String())}, nil
			}
		case xml.StartElement:
			// Recursively extract text from nested elements
			nested, err := extractNestedText(decoder, t.Name.Local)
			if err != nil {
				return nil, err
			}
			content.WriteString(nested)
		}
	}
	return &Heading{Level: level, Text: content.String()}, nil
}

func parseBulletList(decoder *xml.Decoder, _ xml.StartElement) (*BulletList, error) {
	list := &BulletList{Items: []ListItem{}}

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "li" {
				item, err := parseListItem(decoder)
				if err != nil {
					return nil, err
				}
				list.Items = append(list.Items, *item)
			} else {
				if err := skipElement(decoder); err != nil {
					return nil, err
				}
			}
		case xml.EndElement:
			if t.Name.Local == "ul" {
				return list, nil
			}
		}
	}

	return list, nil
}

func parseNumberedList(decoder *xml.Decoder, _ xml.StartElement) (*NumberedList, error) {
	list := &NumberedList{Items: []ListItem{}}

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "li" {
				item, err := parseListItem(decoder)
				if err != nil {
					return nil, err
				}
				list.Items = append(list.Items, *item)
			} else {
				if err := skipElement(decoder); err != nil {
					return nil, err
				}
			}
		case xml.EndElement:
			if t.Name.Local == "ol" {
				return list, nil
			}
		}
	}

	return list, nil
}

func parseListItem(decoder *xml.Decoder) (*ListItem, error) {
	var content strings.Builder

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch t := tok.(type) {
		case xml.CharData:
			content.Write(t)
		case xml.EndElement:
			if t.Name.Local == "li" {
				return &ListItem{Text: strings.TrimSpace(content.String())}, nil
			}
		case xml.StartElement:
			// Recursively extract text from nested elements (p, strong, code, etc.)
			nested, err := extractNestedText(decoder, t.Name.Local)
			if err != nil {
				return nil, err
			}
			content.WriteString(nested)
		}
	}

	return &ListItem{Text: content.String()}, nil
}

func parseMacro(decoder *xml.Decoder, start xml.StartElement) (*Macro, error) {
	macro := &Macro{Params: make(map[string]string)}

	// Get macro name from attributes
	for _, attr := range start.Attr {
		if attr.Name.Local == "name" {
			macro.Name = attr.Value
		}
	}

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "parameter":
				name := ""
				for _, attr := range t.Attr {
					if attr.Name.Local == "name" {
						name = attr.Value
					}
				}
				value, err := parseTextContent(decoder, "parameter")
				if err != nil {
					return nil, err
				}
				if name != "" {
					macro.Params[name] = value
				}
			case "rich-text-body", "plain-text-body":
				body, err := parseTextContent(decoder, t.Name.Local)
				if err != nil {
					return nil, err
				}
				macro.Body = body
			default:
				if err := skipElement(decoder); err != nil {
					return nil, err
				}
			}
		case xml.EndElement:
			if t.Name.Local == "structured-macro" {
				return macro, nil
			}
		}
	}

	return macro, nil
}

func parseTextContent(decoder *xml.Decoder, endTag string) (string, error) {
	var content strings.Builder

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		switch t := tok.(type) {
		case xml.CharData:
			content.Write(t)
		case xml.EndElement:
			if t.Name.Local == endTag {
				return strings.TrimSpace(content.String()), nil
			}
		case xml.StartElement:
			if err := skipElement(decoder); err != nil {
				return "", err
			}
		}
	}

	return content.String(), nil
}

func skipElement(decoder *xml.Decoder) error {
	depth := 1
	for depth > 0 {
		tok, err := decoder.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		switch tok.(type) {
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
		}
	}
	return nil
}

// extractNestedText recursively extracts all text content from an element,
// including text inside nested elements like <strong>, <code>, <a>, <p>, etc.
// It returns the concatenated text and consumes tokens up to and including the end tag.
func extractNestedText(decoder *xml.Decoder, _ string) (string, error) {
	var content strings.Builder
	depth := 1

	for depth > 0 {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		switch t := tok.(type) {
		case xml.CharData:
			content.Write(t)
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
		}
	}

	return strings.TrimSpace(content.String()), nil
}
