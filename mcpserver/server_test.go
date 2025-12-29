package mcpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agentplexus/mcp-confluence/confluence"
	"github.com/agentplexus/mcp-confluence/storage"
)

func TestNew(t *testing.T) {
	client := confluence.NewClient("http://example.com", confluence.BasicAuth{})
	server := New(client)

	if server == nil {
		t.Fatal("New() returned nil")
	}

	if server.client != client {
		t.Error("New() did not set client correctly")
	}
}

func TestTools(t *testing.T) {
	client := confluence.NewClient("http://example.com", confluence.BasicAuth{})
	server := New(client)

	tools := server.Tools()

	expectedTools := []string{
		"confluence_read_page",
		"confluence_read_page_xhtml",
		"confluence_update_page",
		"confluence_update_page_xhtml",
		"confluence_create_page",
		"confluence_create_table",
		"confluence_delete_page",
		"confluence_search_pages",
	}

	if len(tools) != len(expectedTools) {
		t.Errorf("Tools() returned %d tools, want %d", len(tools), len(expectedTools))
	}

	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}

	for _, expected := range expectedTools {
		if !toolNames[expected] {
			t.Errorf("Tools() missing expected tool: %s", expected)
		}
	}
}

func TestToolSchema(t *testing.T) {
	client := confluence.NewClient("http://example.com", confluence.BasicAuth{})
	server := New(client)

	tools := server.Tools()

	for _, tool := range tools {
		if tool.Name == "" {
			t.Error("Tool has empty name")
		}
		if tool.Description == "" {
			t.Errorf("Tool %s has empty description", tool.Name)
		}
		if tool.InputSchema == nil {
			t.Errorf("Tool %s has nil InputSchema", tool.Name)
		}

		schemaType, ok := tool.InputSchema["type"].(string)
		if !ok || schemaType != "object" {
			t.Errorf("Tool %s InputSchema type = %v, want object", tool.Name, schemaType)
		}
	}
}

func TestHandleToolUnknown(t *testing.T) {
	client := confluence.NewClient("http://example.com", confluence.BasicAuth{})
	server := New(client)

	_, err := server.HandleTool(context.Background(), "unknown_tool", nil)
	if err == nil {
		t.Error("HandleTool() should return error for unknown tool")
	}
}

func TestHandleGetPage(t *testing.T) {
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"id":     "12345",
			"type":   "page",
			"status": "current",
			"title":  "Test Page",
			"body": map[string]interface{}{
				"storage": map[string]string{
					"value": "<h1>Title</h1><p>Content</p>",
				},
			},
			"version": map[string]int{
				"number": 5,
			},
			"space": map[string]string{
				"key": "TEST",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			panic(err)
		}
	}))
	defer httpServer.Close()

	client := confluence.NewClient(httpServer.URL, confluence.BasicAuth{})
	server := New(client)

	result, err := server.HandleTool(context.Background(), "confluence_read_page", map[string]interface{}{
		"page_id": "12345",
	})

	if err != nil {
		t.Fatalf("HandleTool() error = %v", err)
	}

	if result.IsError {
		t.Errorf("HandleTool() returned error result: %v", result.Content)
	}

	if len(result.Content) == 0 {
		t.Fatal("HandleTool() returned empty content")
	}

	// Parse the JSON response
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(result.Content[0].Text), &response); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	if response["page_id"] != "12345" {
		t.Errorf("Response page_id = %v, want 12345", response["page_id"])
	}

	if response["title"] != "Test Page" {
		t.Errorf("Response title = %v, want Test Page", response["title"])
	}

	blocks, ok := response["blocks"].([]interface{})
	if !ok || len(blocks) != 2 {
		t.Errorf("Response blocks = %v, want 2 blocks", response["blocks"])
	}
}

func TestHandleGetPage_MissingPageID(t *testing.T) {
	client := confluence.NewClient("http://example.com", confluence.BasicAuth{})
	server := New(client)

	result, err := server.HandleTool(context.Background(), "confluence_read_page", map[string]interface{}{})

	if err != nil {
		t.Fatalf("HandleTool() error = %v", err)
	}

	if !result.IsError {
		t.Error("HandleTool() should return error for missing page_id")
	}
}

func TestHandleCreateTable(t *testing.T) {
	client := confluence.NewClient("http://example.com", confluence.BasicAuth{})
	server := New(client)

	result, err := server.HandleTool(context.Background(), "confluence_create_table", map[string]interface{}{
		"headers": []interface{}{"Name", "Age", "City"},
		"rows": []interface{}{
			[]interface{}{"Alice", "30", "NYC"},
			[]interface{}{"Bob", "25", "LA"},
		},
	})

	if err != nil {
		t.Fatalf("HandleTool() error = %v", err)
	}

	if result.IsError {
		t.Errorf("HandleTool() returned error: %v", result.Content)
	}

	var response map[string]interface{}
	if err := json.Unmarshal([]byte(result.Content[0].Text), &response); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	xhtml, ok := response["xhtml"].(string)
	if !ok {
		t.Fatal("Response missing xhtml field")
	}

	// Verify the XHTML is valid
	if err := storage.Validate(xhtml); err != nil {
		t.Errorf("Generated XHTML is invalid: %v", err)
	}
}

func TestHandleCreateTable_WithMacro(t *testing.T) {
	client := confluence.NewClient("http://example.com", confluence.BasicAuth{})
	server := New(client)

	result, err := server.HandleTool(context.Background(), "confluence_create_table", map[string]interface{}{
		"headers": []interface{}{"Status"},
		"rows": []interface{}{
			[]interface{}{
				map[string]interface{}{
					"macro": map[string]interface{}{
						"name": "status",
						"params": map[string]interface{}{
							"colour": "Green",
							"title":  "OK",
						},
					},
				},
			},
		},
	})

	if err != nil {
		t.Fatalf("HandleTool() error = %v", err)
	}

	if result.IsError {
		t.Errorf("HandleTool() returned error: %v", result.Content)
	}
}

func TestHandleUpdatePage(t *testing.T) {
	callCount := 0
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		switch r.Method {
		case "GET":
			// Return current page info
			response := map[string]interface{}{
				"id":     "12345",
				"type":   "page",
				"status": "current",
				"title":  "Test Page",
				"body": map[string]interface{}{
					"storage": map[string]string{
						"value": "<p>Old content</p>",
					},
				},
				"version": map[string]int{
					"number": 5,
				},
				"space": map[string]string{
					"key": "TEST",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(response); err != nil {
				panic(err)
			}
		case "PUT":
			// Verify the update request
			var payload map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				panic(err)
			}

			version := payload["version"].(map[string]interface{})
			if version["number"] != float64(6) {
				t.Errorf("Update version = %v, want 6", version["number"])
			}

			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write([]byte(`{"id": "12345"}`)); err != nil {
				panic(err)
			}
		}
	}))
	defer httpServer.Close()

	client := confluence.NewClient(httpServer.URL, confluence.BasicAuth{})
	server := New(client)

	result, err := server.HandleTool(context.Background(), "confluence_update_page", map[string]interface{}{
		"page_id": "12345",
		"title":   "Updated Title",
		"blocks": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"text": "Updated content",
			},
		},
	})

	if err != nil {
		t.Fatalf("HandleTool() error = %v", err)
	}

	if result.IsError {
		t.Errorf("HandleTool() returned error: %v", result.Content)
	}

	var response map[string]interface{}
	if err := json.Unmarshal([]byte(result.Content[0].Text), &response); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	if response["status"] != "updated" {
		t.Errorf("Response status = %v, want updated", response["status"])
	}
}

func TestHandleCreatePage(t *testing.T) {
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte(`{"id": "67890"}`)); err != nil {
			panic(err)
		}
	}))
	defer httpServer.Close()

	client := confluence.NewClient(httpServer.URL, confluence.BasicAuth{})
	server := New(client)

	result, err := server.HandleTool(context.Background(), "confluence_create_page", map[string]interface{}{
		"space_key": "TEST",
		"title":     "New Page",
		"blocks": []interface{}{
			map[string]interface{}{
				"type":  "heading",
				"level": float64(1),
				"text":  "Welcome",
			},
			map[string]interface{}{
				"type": "paragraph",
				"text": "Hello, World!",
			},
		},
	})

	if err != nil {
		t.Fatalf("HandleTool() error = %v", err)
	}

	if result.IsError {
		t.Errorf("HandleTool() returned error: %v", result.Content)
	}

	var response map[string]interface{}
	if err := json.Unmarshal([]byte(result.Content[0].Text), &response); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	if response["status"] != "created" {
		t.Errorf("Response status = %v, want created", response["status"])
	}

	if response["page_id"] != "67890" {
		t.Errorf("Response page_id = %v, want 67890", response["page_id"])
	}
}

func TestHandleDeletePage(t *testing.T) {
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer httpServer.Close()

	client := confluence.NewClient(httpServer.URL, confluence.BasicAuth{})
	server := New(client)

	result, err := server.HandleTool(context.Background(), "confluence_delete_page", map[string]interface{}{
		"page_id": "12345",
	})

	if err != nil {
		t.Fatalf("HandleTool() error = %v", err)
	}

	if result.IsError {
		t.Errorf("HandleTool() returned error: %v", result.Content)
	}

	var response map[string]interface{}
	if err := json.Unmarshal([]byte(result.Content[0].Text), &response); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	if response["status"] != "deleted" {
		t.Errorf("Response status = %v, want deleted", response["status"])
	}
}

func TestHandleSearchPages(t *testing.T) {
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": "1", "type": "page", "status": "current", "title": "Page 1"},
				{"id": "2", "type": "page", "status": "current", "title": "Page 2"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			panic(err)
		}
	}))
	defer httpServer.Close()

	client := confluence.NewClient(httpServer.URL, confluence.BasicAuth{})
	server := New(client)

	result, err := server.HandleTool(context.Background(), "confluence_search_pages", map[string]interface{}{
		"cql":   "space=TEST",
		"limit": float64(10),
	})

	if err != nil {
		t.Fatalf("HandleTool() error = %v", err)
	}

	if result.IsError {
		t.Errorf("HandleTool() returned error: %v", result.Content)
	}

	var response map[string]interface{}
	if err := json.Unmarshal([]byte(result.Content[0].Text), &response); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	if response["count"] != float64(2) {
		t.Errorf("Response count = %v, want 2", response["count"])
	}
}

func TestParseBlocks(t *testing.T) {
	tests := []struct {
		name       string
		blocks     []interface{}
		wantBlocks int
		wantErr    bool
	}{
		{
			name:       "empty blocks",
			blocks:     []interface{}{},
			wantBlocks: 0,
			wantErr:    false,
		},
		{
			name: "paragraph",
			blocks: []interface{}{
				map[string]interface{}{"type": "paragraph", "text": "Hello"},
			},
			wantBlocks: 1,
			wantErr:    false,
		},
		{
			name: "heading",
			blocks: []interface{}{
				map[string]interface{}{"type": "heading", "level": float64(2), "text": "Title"},
			},
			wantBlocks: 1,
			wantErr:    false,
		},
		{
			name: "table",
			blocks: []interface{}{
				map[string]interface{}{
					"type":    "table",
					"headers": []interface{}{"A", "B"},
					"rows":    []interface{}{[]interface{}{"1", "2"}},
				},
			},
			wantBlocks: 1,
			wantErr:    false,
		},
		{
			name: "bullet_list",
			blocks: []interface{}{
				map[string]interface{}{
					"type":  "bullet_list",
					"items": []interface{}{"Item 1", "Item 2"},
				},
			},
			wantBlocks: 1,
			wantErr:    false,
		},
		{
			name: "numbered_list",
			blocks: []interface{}{
				map[string]interface{}{
					"type":  "numbered_list",
					"items": []interface{}{"First", "Second"},
				},
			},
			wantBlocks: 1,
			wantErr:    false,
		},
		{
			name: "macro",
			blocks: []interface{}{
				map[string]interface{}{
					"type":   "macro",
					"name":   "info",
					"params": map[string]interface{}{"title": "Note"},
					"body":   "Content",
				},
			},
			wantBlocks: 1,
			wantErr:    false,
		},
		{
			name: "code_block",
			blocks: []interface{}{
				map[string]interface{}{
					"type":     "code_block",
					"language": "go",
					"code":     "func main() {}",
				},
			},
			wantBlocks: 1,
			wantErr:    false,
		},
		{
			name: "horizontal_rule",
			blocks: []interface{}{
				map[string]interface{}{"type": "horizontal_rule"},
			},
			wantBlocks: 1,
			wantErr:    false,
		},
		{
			name: "unknown type",
			blocks: []interface{}{
				map[string]interface{}{"type": "unknown"},
			},
			wantBlocks: 0,
			wantErr:    true,
		},
		{
			name: "multiple blocks",
			blocks: []interface{}{
				map[string]interface{}{"type": "heading", "level": float64(1), "text": "Title"},
				map[string]interface{}{"type": "paragraph", "text": "Content"},
				map[string]interface{}{"type": "bullet_list", "items": []interface{}{"A", "B"}},
			},
			wantBlocks: 3,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, err := parseBlocks(tt.blocks)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseBlocks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(page.Blocks) != tt.wantBlocks {
				t.Errorf("parseBlocks() blocks = %d, want %d", len(page.Blocks), tt.wantBlocks)
			}
		})
	}
}

func TestBlocksToJSON(t *testing.T) {
	blocks := []storage.Block{
		&storage.Heading{Level: 1, Text: "Title"},
		&storage.Paragraph{Text: "Content"},
		&storage.Table{
			Headers: []string{"A", "B"},
			Rows:    []storage.Row{{Cells: []storage.Cell{{Text: "1"}, {Text: "2"}}}},
		},
		&storage.BulletList{Items: []storage.ListItem{{Text: "Item 1"}}},
		&storage.NumberedList{Items: []storage.ListItem{{Text: "First"}}},
		&storage.Macro{Name: "info", Params: map[string]string{"title": "Note"}},
		&storage.CodeBlock{Language: "go", Code: "func main() {}"},
		&storage.HorizontalRule{},
	}

	result := blocksToJSON(blocks)

	if len(result) != len(blocks) {
		t.Errorf("blocksToJSON() returned %d items, want %d", len(result), len(blocks))
	}

	// Check first block (heading)
	heading, ok := result[0].(map[string]interface{})
	if !ok {
		t.Fatal("First block should be a map")
	}
	if heading["type"] != "heading" {
		t.Errorf("First block type = %v, want heading", heading["type"])
	}
}

func TestParseCell(t *testing.T) {
	tests := []struct {
		name      string
		input     interface{}
		wantText  string
		wantMacro bool
	}{
		{
			name:      "string cell",
			input:     "Hello",
			wantText:  "Hello",
			wantMacro: false,
		},
		{
			name:      "text object cell",
			input:     map[string]interface{}{"text": "World"},
			wantText:  "World",
			wantMacro: false,
		},
		{
			name: "macro cell",
			input: map[string]interface{}{
				"macro": map[string]interface{}{
					"name":   "status",
					"params": map[string]interface{}{"colour": "Green"},
				},
			},
			wantText:  "",
			wantMacro: true,
		},
		{
			name:      "empty input",
			input:     nil,
			wantText:  "",
			wantMacro: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cell := parseCell(tt.input)
			if cell.Text != tt.wantText {
				t.Errorf("parseCell() Text = %v, want %v", cell.Text, tt.wantText)
			}
			hasMacro := cell.Macro != nil
			if hasMacro != tt.wantMacro {
				t.Errorf("parseCell() has Macro = %v, want %v", hasMacro, tt.wantMacro)
			}
		})
	}
}
