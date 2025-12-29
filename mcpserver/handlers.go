package mcpserver

import (
	"context"
	"fmt"

	"github.com/agentplexus/mcp-confluence/storage"
)

func (s *Server) handleReadPage(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	pageID, ok := input["page_id"].(string)
	if !ok || pageID == "" {
		return nil, fmt.Errorf("page_id is required")
	}

	page, info, err := s.client.GetPageStorage(ctx, pageID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"page_id":   info.ID,
		"title":     info.Title,
		"version":   info.Version,
		"space_key": info.SpaceKey,
		"blocks":    blocksToJSON(page.Blocks),
	}, nil
}

func (s *Server) handleReadPageXHTML(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	pageID, ok := input["page_id"].(string)
	if !ok || pageID == "" {
		return nil, fmt.Errorf("page_id is required")
	}

	xhtml, info, err := s.client.GetPageStorageRaw(ctx, pageID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"page_id":   info.ID,
		"title":     info.Title,
		"version":   info.Version,
		"space_key": info.SpaceKey,
		"xhtml":     xhtml,
	}, nil
}

func (s *Server) handleUpdatePage(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	pageID, _ := input["page_id"].(string)
	title, _ := input["title"].(string)
	blocksRaw, _ := input["blocks"].([]interface{})

	if pageID == "" {
		return nil, fmt.Errorf("page_id is required")
	}
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}

	page, err := parseBlocks(blocksRaw)
	if err != nil {
		return nil, fmt.Errorf("invalid blocks: %w", err)
	}

	// Get current version
	_, info, err := s.client.GetPageStorageRaw(ctx, pageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current version: %w", err)
	}

	if err := s.client.UpdatePageStorage(ctx, pageID, page, info.Version, title); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"status":  "updated",
		"page_id": pageID,
		"title":   title,
		"version": info.Version + 1,
	}, nil
}

func (s *Server) handleUpdatePageXHTML(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	pageID, _ := input["page_id"].(string)
	title, _ := input["title"].(string)
	xhtml, _ := input["xhtml"].(string)

	if pageID == "" {
		return nil, fmt.Errorf("page_id is required")
	}
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if xhtml == "" {
		return nil, fmt.Errorf("xhtml is required")
	}

	// Get current version
	_, info, err := s.client.GetPageStorageRaw(ctx, pageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current version: %w", err)
	}

	if err := s.client.UpdatePageStorageRaw(ctx, pageID, xhtml, info.Version, title); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"status":  "updated",
		"page_id": pageID,
		"title":   title,
		"version": info.Version + 1,
	}, nil
}

func (s *Server) handleCreatePage(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	spaceKey, _ := input["space_key"].(string)
	title, _ := input["title"].(string)
	parentID, _ := input["parent_id"].(string)
	blocksRaw, _ := input["blocks"].([]interface{})

	if spaceKey == "" {
		return nil, fmt.Errorf("space_key is required")
	}
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}

	page, err := parseBlocks(blocksRaw)
	if err != nil {
		return nil, fmt.Errorf("invalid blocks: %w", err)
	}

	pageID, err := s.client.CreatePage(ctx, spaceKey, title, page, parentID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"status":    "created",
		"page_id":   pageID,
		"title":     title,
		"space_key": spaceKey,
	}, nil
}

func (s *Server) handleCreateTable(_ context.Context, input map[string]interface{}) (interface{}, error) {
	headersRaw, _ := input["headers"].([]interface{})
	rowsRaw, _ := input["rows"].([]interface{})

	table := &storage.Table{
		Headers: make([]string, 0, len(headersRaw)),
		Rows:    make([]storage.Row, 0, len(rowsRaw)),
	}

	for _, h := range headersRaw {
		if str, ok := h.(string); ok {
			table.Headers = append(table.Headers, str)
		}
	}

	for _, rowRaw := range rowsRaw {
		row := storage.Row{Cells: []storage.Cell{}}
		cells, ok := rowRaw.([]interface{})
		if !ok {
			continue
		}
		for _, cellRaw := range cells {
			cell := parseCell(cellRaw)
			row.Cells = append(row.Cells, cell)
		}
		table.Rows = append(table.Rows, row)
	}

	// Validate the table
	if err := storage.ValidateBlock(table); err != nil {
		return nil, fmt.Errorf("table validation failed: %w", err)
	}

	// Render to XHTML for preview
	xhtml, err := storage.RenderBlock(table)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"block": map[string]interface{}{
			"type":    "table",
			"headers": table.Headers,
			"rows":    table.Rows,
		},
		"xhtml": xhtml,
	}, nil
}

func (s *Server) handleDeletePage(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	pageID, ok := input["page_id"].(string)
	if !ok || pageID == "" {
		return nil, fmt.Errorf("page_id is required")
	}

	if err := s.client.DeletePage(ctx, pageID); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"status":  "deleted",
		"page_id": pageID,
	}, nil
}

func (s *Server) handleSearchPages(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	cql, ok := input["cql"].(string)
	if !ok || cql == "" {
		return nil, fmt.Errorf("cql is required")
	}

	limit := 25
	if l, ok := input["limit"].(float64); ok {
		limit = int(l)
	}

	pages, err := s.client.SearchPages(ctx, cql, limit)
	if err != nil {
		return nil, err
	}

	results := make([]map[string]interface{}, len(pages))
	for i, p := range pages {
		results[i] = map[string]interface{}{
			"page_id": p.ID,
			"title":   p.Title,
			"type":    p.Type,
			"status":  p.Status,
		}
	}

	return map[string]interface{}{
		"count":   len(results),
		"results": results,
	}, nil
}
