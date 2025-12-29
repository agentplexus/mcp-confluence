// Package confluence provides a client for the Confluence REST API.
// It integrates with the storage package for reading and writing pages
// using Confluence Storage Format.
package confluence

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/agentplexus/mcp-confluence/storage"
)

// Client is a Confluence REST API client.
type Client struct {
	baseURL    string
	httpClient *http.Client
	auth       AuthMethod
}

// AuthMethod represents an authentication method.
type AuthMethod interface {
	Apply(req *http.Request)
}

// BasicAuth implements basic authentication using API tokens.
type BasicAuth struct {
	Username string
	Token    string // API token (not password)
}

// Apply implements AuthMethod.
func (b BasicAuth) Apply(req *http.Request) {
	req.SetBasicAuth(b.Username, b.Token)
}

// BearerAuth implements bearer token authentication.
type BearerAuth struct {
	Token string
}

// Apply implements AuthMethod.
func (b BearerAuth) Apply(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+b.Token)
}

// NewClient creates a new Confluence client.
func NewClient(baseURL string, auth AuthMethod, opts ...Option) *Client {
	c := &Client{
		baseURL:    baseURL,
		httpClient: http.DefaultClient,
		auth:       auth,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// APIError represents an error returned by the Confluence API.
type APIError struct {
	StatusCode int
	Message    string
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("confluence API error %d: %s", e.StatusCode, e.Message)
}

// PageInfo contains metadata about a Confluence page.
type PageInfo struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Status   string `json:"status"`
	Title    string `json:"title"`
	Version  int    `json:"version"`
	SpaceKey string `json:"spaceKey,omitempty"`
}

// GetPageStorage retrieves a page's content as Storage Format, parsed to IR.
func (c *Client) GetPageStorage(ctx context.Context, pageID string) (*storage.Page, *PageInfo, error) {
	xhtml, info, err := c.GetPageStorageRaw(ctx, pageID)
	if err != nil {
		return nil, nil, err
	}

	page, err := storage.Parse(xhtml)
	if err != nil {
		return nil, info, fmt.Errorf("parse error: %w", err)
	}

	return page, info, nil
}

// GetPageStorageRaw retrieves a page's raw Storage XHTML.
func (c *Client) GetPageStorageRaw(ctx context.Context, pageID string) (string, *PageInfo, error) {
	u := fmt.Sprintf("%s/rest/api/content/%s?expand=body.storage,version,space", c.baseURL, pageID)

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("Accept", "application/json")
	c.auth.Apply(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    "failed to get page",
			Body:       string(body),
		}
	}

	var result struct {
		ID     string `json:"id"`
		Type   string `json:"type"`
		Status string `json:"status"`
		Title  string `json:"title"`
		Body   struct {
			Storage struct {
				Value string `json:"value"`
			} `json:"storage"`
		} `json:"body"`
		Version struct {
			Number int `json:"number"`
		} `json:"version"`
		Space struct {
			Key string `json:"key"`
		} `json:"space"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", nil, fmt.Errorf("json decode error: %w", err)
	}

	info := &PageInfo{
		ID:       result.ID,
		Type:     result.Type,
		Status:   result.Status,
		Title:    result.Title,
		Version:  result.Version.Number,
		SpaceKey: result.Space.Key,
	}

	return result.Body.Storage.Value, info, nil
}

// UpdatePageStorage updates a page with the given IR content.
func (c *Client) UpdatePageStorage(ctx context.Context, pageID string, page *storage.Page, version int, title string) error {
	xhtml, err := storage.Render(page)
	if err != nil {
		return fmt.Errorf("render error: %w", err)
	}

	if err := storage.Validate(xhtml); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	return c.UpdatePageStorageRaw(ctx, pageID, xhtml, version, title)
}

// UpdatePageStorageRaw updates a page with raw Storage XHTML.
func (c *Client) UpdatePageStorageRaw(ctx context.Context, pageID, xhtml string, version int, title string) error {
	if err := storage.Validate(xhtml); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	u := fmt.Sprintf("%s/rest/api/content/%s", c.baseURL, pageID)

	payload := map[string]interface{}{
		"type":  "page",
		"title": title,
		"body": map[string]interface{}{
			"storage": map[string]string{
				"value":          xhtml,
				"representation": "storage",
			},
		},
		"version": map[string]int{
			"number": version + 1,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", u, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	c.auth.Apply(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read error response body: %w", err)
		}
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    "failed to update page",
			Body:       string(respBody),
		}
	}

	return nil
}

// CreatePage creates a new page with IR content.
func (c *Client) CreatePage(ctx context.Context, spaceKey, title string, page *storage.Page, parentID string) (string, error) {
	xhtml, err := storage.Render(page)
	if err != nil {
		return "", fmt.Errorf("render error: %w", err)
	}

	if err := storage.Validate(xhtml); err != nil {
		return "", fmt.Errorf("validation error: %w", err)
	}

	return c.CreatePageRaw(ctx, spaceKey, title, xhtml, parentID)
}

// CreatePageRaw creates a new page with raw Storage XHTML.
func (c *Client) CreatePageRaw(ctx context.Context, spaceKey, title, xhtml, parentID string) (string, error) {
	if err := storage.Validate(xhtml); err != nil {
		return "", fmt.Errorf("validation error: %w", err)
	}

	u := fmt.Sprintf("%s/rest/api/content", c.baseURL)

	payload := map[string]interface{}{
		"type":  "page",
		"title": title,
		"space": map[string]string{"key": spaceKey},
		"body": map[string]interface{}{
			"storage": map[string]string{
				"value":          xhtml,
				"representation": "storage",
			},
		},
	}

	if parentID != "" {
		payload["ancestors"] = []map[string]string{{"id": parentID}}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", u, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	c.auth.Apply(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", &APIError{
			StatusCode: resp.StatusCode,
			Message:    "failed to create page",
			Body:       string(respBody),
		}
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("json decode error: %w", err)
	}

	return result.ID, nil
}

// DeletePage deletes a page by ID.
func (c *Client) DeletePage(ctx context.Context, pageID string) error {
	u := fmt.Sprintf("%s/rest/api/content/%s", c.baseURL, pageID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", u, nil)
	if err != nil {
		return err
	}
	c.auth.Apply(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read error response body: %w", err)
		}
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    "failed to delete page",
			Body:       string(respBody),
		}
	}

	return nil
}

// GetSpace retrieves information about a space.
func (c *Client) GetSpace(ctx context.Context, spaceKey string) (*SpaceInfo, error) {
	u := fmt.Sprintf("%s/rest/api/space/%s", c.baseURL, spaceKey)

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	c.auth.Apply(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    "failed to get space",
			Body:       string(body),
		}
	}

	var result SpaceInfo
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("json decode error: %w", err)
	}

	return &result, nil
}

// SpaceInfo contains metadata about a Confluence space.
type SpaceInfo struct {
	ID          int    `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// SearchPages searches for pages matching the given CQL query.
func (c *Client) SearchPages(ctx context.Context, cql string, limit int) ([]PageInfo, error) {
	u := fmt.Sprintf("%s/rest/api/content/search?cql=%s&limit=%d", c.baseURL, cql, limit)

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	c.auth.Apply(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    "failed to search pages",
			Body:       string(body),
		}
	}

	var result struct {
		Results []struct {
			ID     string `json:"id"`
			Type   string `json:"type"`
			Status string `json:"status"`
			Title  string `json:"title"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("json decode error: %w", err)
	}

	pages := make([]PageInfo, len(result.Results))
	for i, r := range result.Results {
		pages[i] = PageInfo{
			ID:     r.ID,
			Type:   r.Type,
			Status: r.Status,
			Title:  r.Title,
		}
	}

	return pages, nil
}
