// Command mcp-confluence runs a Confluence MCP server that exposes tools for
// reading, creating, and updating Confluence pages using structured content blocks.
//
// The server reads configuration from environment variables:
//   - CONFLUENCE_BASE_URL: The base URL of your Confluence instance (e.g., https://example.atlassian.net/wiki)
//   - CONFLUENCE_USERNAME: Your Confluence username (email)
//   - CONFLUENCE_API_TOKEN: Your Confluence API token
//
// Example usage:
//
//	export CONFLUENCE_BASE_URL=https://example.atlassian.net/wiki
//	export CONFLUENCE_USERNAME=user@example.com
//	export CONFLUENCE_API_TOKEN=your-api-token
//	go run ./cmd/mcp-confluence
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/agentplexus/mcp-confluence/confluence"
	"github.com/agentplexus/mcp-confluence/mcpserver"
)

const serverName = "mcp-confluence"
const serverVersion = "0.1.0"

// MCP JSON-RPC message types
type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *rpcError   `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func main() {
	// Load configuration from environment
	baseURL := os.Getenv("CONFLUENCE_BASE_URL")
	username := os.Getenv("CONFLUENCE_USERNAME")
	apiToken := os.Getenv("CONFLUENCE_API_TOKEN")

	if baseURL == "" {
		log.Fatal("CONFLUENCE_BASE_URL environment variable is required")
	}
	if username == "" {
		log.Fatal("CONFLUENCE_USERNAME environment variable is required")
	}
	if apiToken == "" {
		log.Fatal("CONFLUENCE_API_TOKEN environment variable is required")
	}

	// Create Confluence client
	auth := confluence.BasicAuth{
		Username: username,
		Token:    apiToken,
	}
	client := confluence.NewClient(baseURL, auth)

	// Create MCP server
	server := mcpserver.New(client)

	// Run the stdio transport
	if err := runStdio(server); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func runStdio(server *mcpserver.Server) error {
	reader := bufio.NewReader(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	log.SetOutput(os.Stderr) // Log to stderr to keep stdout clean for JSON-RPC

	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read error: %w", err)
		}

		var req jsonRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			sendError(encoder, nil, -32700, "Parse error")
			continue
		}

		// Notifications (requests without an id) don't get responses
		if req.ID == nil {
			handleNotification(&req)
			continue
		}

		response := handleRequest(server, &req)
		if err := encoder.Encode(response); err != nil {
			log.Printf("Write error: %v", err)
		}
	}
}

func handleNotification(req *jsonRPCRequest) {
	switch req.Method {
	case "initialized":
		// Client acknowledges initialization complete - nothing to do
	case "cancelled":
		// Client cancelled a request - nothing to do for now
	default:
		log.Printf("Unknown notification: %s", req.Method)
	}
}

func handleRequest(server *mcpserver.Server, req *jsonRPCRequest) *jsonRPCResponse {
	switch req.Method {
	case "initialize":
		return handleInitialize(req)
	case "tools/list":
		return handleToolsList(server, req)
	case "tools/call":
		return handleToolsCall(server, req)
	case "ping":
		return &jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  map[string]string{},
		}
	default:
		return &jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &rpcError{Code: -32601, Message: "Method not found"},
		}
	}
}

func handleInitialize(req *jsonRPCRequest) *jsonRPCResponse {
	return &jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]string{
				"name":    serverName,
				"version": serverVersion,
			},
		},
	}
}

func handleToolsList(server *mcpserver.Server, req *jsonRPCRequest) *jsonRPCResponse {
	tools := server.Tools()

	toolsList := make([]map[string]interface{}, len(tools))
	for i, tool := range tools {
		toolsList[i] = map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
		}
	}

	return &jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"tools": toolsList,
		},
	}
}

func handleToolsCall(server *mcpserver.Server, req *jsonRPCRequest) *jsonRPCResponse {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &rpcError{Code: -32602, Message: "Invalid params"},
		}
	}

	result, err := server.HandleTool(context.Background(), params.Name, params.Arguments)
	if err != nil {
		return &jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &rpcError{Code: -32000, Message: err.Error()},
		}
	}

	return &jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

func sendError(encoder *json.Encoder, id interface{}, code int, message string) {
	if err := encoder.Encode(&jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &rpcError{Code: code, Message: message},
	}); err != nil {
		log.Printf("Failed to send error response: %v", err)
	}
}
