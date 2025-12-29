package mcpserver

import (
	"fmt"

	"github.com/agentplexus/mcp-confluence/storage"
)

// blocksToJSON converts a slice of storage.Block to JSON-serializable format.
func blocksToJSON(blocks []storage.Block) []interface{} {
	result := make([]interface{}, 0, len(blocks))
	for _, block := range blocks {
		result = append(result, blockToJSON(block))
	}
	return result
}

// blockToJSON converts a single storage.Block to JSON-serializable format.
func blockToJSON(block storage.Block) map[string]interface{} {
	switch b := block.(type) {
	case *storage.Table:
		rows := make([]interface{}, len(b.Rows))
		for i, row := range b.Rows {
			cells := make([]interface{}, len(row.Cells))
			for j, cell := range row.Cells {
				if cell.Macro != nil {
					cells[j] = map[string]interface{}{
						"macro": map[string]interface{}{
							"name":   cell.Macro.Name,
							"params": cell.Macro.Params,
						},
					}
				} else {
					cells[j] = cell.Text
				}
			}
			rows[i] = cells
		}
		return map[string]interface{}{
			"type":    "table",
			"headers": b.Headers,
			"rows":    rows,
		}
	case *storage.Paragraph:
		return map[string]interface{}{
			"type": "paragraph",
			"text": b.Text,
		}
	case *storage.Heading:
		return map[string]interface{}{
			"type":  "heading",
			"level": b.Level,
			"text":  b.Text,
		}
	case *storage.BulletList:
		items := make([]string, len(b.Items))
		for i, item := range b.Items {
			items[i] = item.Text
		}
		return map[string]interface{}{
			"type":  "bullet_list",
			"items": items,
		}
	case *storage.NumberedList:
		items := make([]string, len(b.Items))
		for i, item := range b.Items {
			items[i] = item.Text
		}
		return map[string]interface{}{
			"type":  "numbered_list",
			"items": items,
		}
	case *storage.Macro:
		return map[string]interface{}{
			"type":   "macro",
			"name":   b.Name,
			"params": b.Params,
			"body":   b.Body,
		}
	case *storage.CodeBlock:
		return map[string]interface{}{
			"type":     "code_block",
			"language": b.Language,
			"code":     b.Code,
		}
	case *storage.HorizontalRule:
		return map[string]interface{}{
			"type": "horizontal_rule",
		}
	default:
		return map[string]interface{}{
			"type": "unknown",
		}
	}
}

// parseBlocks converts JSON input to a storage.Page.
func parseBlocks(blocksRaw []interface{}) (*storage.Page, error) {
	page := &storage.Page{Blocks: []storage.Block{}}
	for _, b := range blocksRaw {
		block, err := parseBlock(b)
		if err != nil {
			return nil, err
		}
		if block != nil {
			page.Blocks = append(page.Blocks, block)
		}
	}
	return page, nil
}

// parseBlock converts a single JSON block to a storage.Block.
func parseBlock(raw interface{}) (storage.Block, error) {
	m, ok := raw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("block must be an object")
	}

	blockType, _ := m["type"].(string)

	switch blockType {
	case "table":
		return parseTableBlock(m), nil
	case "paragraph":
		text, _ := m["text"].(string)
		return &storage.Paragraph{Text: text}, nil
	case "heading":
		level := 1
		if l, ok := m["level"].(float64); ok {
			level = int(l)
		}
		text, _ := m["text"].(string)
		return &storage.Heading{Level: level, Text: text}, nil
	case "bullet_list":
		return parseBulletListBlock(m), nil
	case "numbered_list":
		return parseNumberedListBlock(m), nil
	case "macro":
		return parseMacroBlock(m), nil
	case "code_block":
		language, _ := m["language"].(string)
		code, _ := m["code"].(string)
		return &storage.CodeBlock{Language: language, Code: code}, nil
	case "horizontal_rule":
		return &storage.HorizontalRule{}, nil
	default:
		return nil, fmt.Errorf("unknown block type: %s", blockType)
	}
}

func parseTableBlock(m map[string]interface{}) *storage.Table {
	table := &storage.Table{
		Headers: []string{},
		Rows:    []storage.Row{},
	}

	if headers, ok := m["headers"].([]interface{}); ok {
		for _, h := range headers {
			if s, ok := h.(string); ok {
				table.Headers = append(table.Headers, s)
			}
		}
	}

	if rows, ok := m["rows"].([]interface{}); ok {
		for _, rowRaw := range rows {
			row := storage.Row{Cells: []storage.Cell{}}
			if cells, ok := rowRaw.([]interface{}); ok {
				for _, cellRaw := range cells {
					cell := parseCell(cellRaw)
					row.Cells = append(row.Cells, cell)
				}
			}
			table.Rows = append(table.Rows, row)
		}
	}

	return table
}

func parseBulletListBlock(m map[string]interface{}) *storage.BulletList {
	list := &storage.BulletList{Items: []storage.ListItem{}}
	if items, ok := m["items"].([]interface{}); ok {
		for _, item := range items {
			if s, ok := item.(string); ok {
				list.Items = append(list.Items, storage.ListItem{Text: s})
			}
		}
	}
	return list
}

func parseNumberedListBlock(m map[string]interface{}) *storage.NumberedList {
	list := &storage.NumberedList{Items: []storage.ListItem{}}
	if items, ok := m["items"].([]interface{}); ok {
		for _, item := range items {
			if s, ok := item.(string); ok {
				list.Items = append(list.Items, storage.ListItem{Text: s})
			}
		}
	}
	return list
}

func parseMacroBlock(m map[string]interface{}) *storage.Macro {
	macro := &storage.Macro{
		Params: make(map[string]string),
	}
	macro.Name, _ = m["name"].(string)
	macro.Body, _ = m["body"].(string)

	if params, ok := m["params"].(map[string]interface{}); ok {
		for k, v := range params {
			if s, ok := v.(string); ok {
				macro.Params[k] = s
			}
		}
	}

	return macro
}

func parseCell(raw interface{}) storage.Cell {
	switch v := raw.(type) {
	case string:
		return storage.Cell{Text: v}
	case map[string]interface{}:
		if text, ok := v["text"].(string); ok {
			return storage.Cell{Text: text}
		}
		if macroRaw, ok := v["macro"].(map[string]interface{}); ok {
			macro := &storage.Macro{
				Params: make(map[string]string),
			}
			macro.Name, _ = macroRaw["name"].(string)
			if params, ok := macroRaw["params"].(map[string]interface{}); ok {
				for k, val := range params {
					if s, ok := val.(string); ok {
						macro.Params[k] = s
					}
				}
			}
			return storage.Cell{Macro: macro}
		}
	}
	return storage.Cell{}
}
