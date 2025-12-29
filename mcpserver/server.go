// Package mcpserver provides an MCP (Model Context Protocol) server implementation
// for Confluence. It exposes tools for reading, creating, and updating Confluence pages
// using structured content blocks instead of raw XHTML, ensuring safe and valid output.
package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/agentplexus/mcp-confluence/confluence"
)

// Server is the MCP server for Confluence.
type Server struct {
	client *confluence.Client
}

// New creates a new MCP server with the given Confluence client.
func New(client *confluence.Client) *Server {
	return &Server{client: client}
}

// Tool represents an MCP tool definition.
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ToolHandler is a function that handles a tool call.
type ToolHandler func(ctx context.Context, input map[string]interface{}) (interface{}, error)

// ToolResult represents the result of a tool execution.
type ToolResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

// ContentBlock represents a content block in an MCP response.
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// HandleTool dispatches a tool call to the appropriate handler.
func (s *Server) HandleTool(ctx context.Context, name string, input map[string]interface{}) (*ToolResult, error) {
	var result interface{}
	var err error

	switch name {
	case "confluence_read_page":
		result, err = s.handleReadPage(ctx, input)
	case "confluence_read_page_xhtml":
		result, err = s.handleReadPageXHTML(ctx, input)
	case "confluence_update_page":
		result, err = s.handleUpdatePage(ctx, input)
	case "confluence_update_page_xhtml":
		result, err = s.handleUpdatePageXHTML(ctx, input)
	case "confluence_create_page":
		result, err = s.handleCreatePage(ctx, input)
	case "confluence_create_table":
		result, err = s.handleCreateTable(ctx, input)
	case "confluence_delete_page":
		result, err = s.handleDeletePage(ctx, input)
	case "confluence_search_pages":
		result, err = s.handleSearchPages(ctx, input)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}

	if err != nil {
		return &ToolResult{
			Content: []ContentBlock{{Type: "text", Text: err.Error()}},
			IsError: true,
		}, nil
	}

	text, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, err
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: string(text)}},
	}, nil
}
