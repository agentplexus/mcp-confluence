package mcpserver

// Tools returns the list of available MCP tools.
func (s *Server) Tools() []Tool {
	return []Tool{
		{
			Name:        "confluence_read_page",
			Description: "Read a Confluence page as structured content blocks. Returns the page content parsed into blocks (paragraphs, tables, headings, etc.) that can be safely modified.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"page_id": map[string]interface{}{
						"type":        "string",
						"description": "The Confluence page ID",
					},
				},
				"required": []string{"page_id"},
			},
		},
		{
			Name:        "confluence_read_page_xhtml",
			Description: "Read a Confluence page as raw Storage Format XHTML. Returns the unparsed XHTML body for debugging or when the block parser doesn't support certain content.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"page_id": map[string]interface{}{
						"type":        "string",
						"description": "The Confluence page ID",
					},
				},
				"required": []string{"page_id"},
			},
		},
		{
			Name:        "confluence_update_page",
			Description: "Update a Confluence page with structured content blocks. Accepts an array of blocks (paragraphs, tables, headings, etc.) and safely renders them to valid Confluence Storage XHTML.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"page_id": map[string]interface{}{
						"type":        "string",
						"description": "The Confluence page ID",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "The page title",
					},
					"blocks": map[string]interface{}{
						"type":        "array",
						"description": "Array of content blocks",
						"items": map[string]interface{}{
							"type": "object",
						},
					},
				},
				"required": []string{"page_id", "title", "blocks"},
			},
		},
		{
			Name:        "confluence_update_page_xhtml",
			Description: "Update a Confluence page with raw Storage Format XHTML. Use this when you need to preserve all formatting, attributes, and structure that the block-based update would lose (complex tables, inline styles, macros, etc.).",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"page_id": map[string]interface{}{
						"type":        "string",
						"description": "The Confluence page ID",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "The page title",
					},
					"xhtml": map[string]interface{}{
						"type":        "string",
						"description": "The raw Storage Format XHTML content",
					},
				},
				"required": []string{"page_id", "title", "xhtml"},
			},
		},
		{
			Name:        "confluence_create_page",
			Description: "Create a new Confluence page with structured content blocks. Accepts an array of blocks and safely renders them to valid Confluence Storage XHTML.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"space_key": map[string]interface{}{
						"type":        "string",
						"description": "The space key where the page will be created",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "The page title",
					},
					"blocks": map[string]interface{}{
						"type":        "array",
						"description": "Array of content blocks",
						"items": map[string]interface{}{
							"type": "object",
						},
					},
					"parent_id": map[string]interface{}{
						"type":        "string",
						"description": "Optional parent page ID",
					},
				},
				"required": []string{"space_key", "title", "blocks"},
			},
		},
		{
			Name:        "confluence_create_table",
			Description: "Create a table block from structured data. Returns a table block that can be included in page content.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"headers": map[string]interface{}{
						"type":        "array",
						"description": "Column headers",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"rows": map[string]interface{}{
						"type":        "array",
						"description": "Table rows, each row is an array of cells",
						"items": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{
								"oneOf": []map[string]interface{}{
									{"type": "string"},
									{
										"type": "object",
										"properties": map[string]interface{}{
											"text":  map[string]string{"type": "string"},
											"macro": map[string]string{"type": "object"},
										},
									},
								},
							},
						},
					},
				},
				"required": []string{"headers", "rows"},
			},
		},
		{
			Name:        "confluence_delete_page",
			Description: "Delete a Confluence page by ID.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"page_id": map[string]interface{}{
						"type":        "string",
						"description": "The Confluence page ID to delete",
					},
				},
				"required": []string{"page_id"},
			},
		},
		{
			Name:        "confluence_search_pages",
			Description: "Search for Confluence pages using CQL (Confluence Query Language).",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"cql": map[string]interface{}{
						"type":        "string",
						"description": "CQL query string (e.g., 'space=TEST and type=page')",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of results (default 25)",
					},
				},
				"required": []string{"cql"},
			},
		},
	}
}
