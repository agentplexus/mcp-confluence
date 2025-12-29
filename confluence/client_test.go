package confluence

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agentplexus/mcp-confluence/storage"
)

func TestNewClient(t *testing.T) {
	auth := BasicAuth{Username: "user", Token: "token"}
	client := NewClient("https://example.atlassian.net/wiki", auth)

	if client.baseURL != "https://example.atlassian.net/wiki" {
		t.Errorf("NewClient() baseURL = %v, want https://example.atlassian.net/wiki", client.baseURL)
	}

	if client.httpClient != http.DefaultClient {
		t.Errorf("NewClient() httpClient should be default client")
	}
}

func TestNewClientWithOptions(t *testing.T) {
	auth := BasicAuth{Username: "user", Token: "token"}
	customClient := &http.Client{}
	client := NewClient("https://example.atlassian.net/wiki", auth, WithHTTPClient(customClient))

	if client.httpClient != customClient {
		t.Errorf("NewClient() with WithHTTPClient should set custom client")
	}
}

func TestBasicAuthApply(t *testing.T) {
	auth := BasicAuth{Username: "user", Token: "token"}
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	auth.Apply(req)

	username, password, ok := req.BasicAuth()
	if !ok {
		t.Fatal("BasicAuth.Apply() should set basic auth")
	}
	if username != "user" {
		t.Errorf("BasicAuth.Apply() username = %v, want user", username)
	}
	if password != "token" {
		t.Errorf("BasicAuth.Apply() password = %v, want token", password)
	}
}

func TestBearerAuthApply(t *testing.T) {
	auth := BearerAuth{Token: "mytoken"}
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	auth.Apply(req)

	authHeader := req.Header.Get("Authorization")
	if authHeader != "Bearer mytoken" {
		t.Errorf("BearerAuth.Apply() Authorization = %v, want Bearer mytoken", authHeader)
	}
}

func TestAPIError(t *testing.T) {
	err := &APIError{
		StatusCode: 404,
		Message:    "page not found",
		Body:       `{"message": "Page does not exist"}`,
	}

	expected := "confluence API error 404: page not found"
	if err.Error() != expected {
		t.Errorf("APIError.Error() = %v, want %v", err.Error(), expected)
	}
}

func TestGetPageStorageRaw(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/rest/api/content/12345" {
			t.Errorf("Expected path /rest/api/content/12345, got %s", r.URL.Path)
		}

		response := map[string]interface{}{
			"id":     "12345",
			"type":   "page",
			"status": "current",
			"title":  "Test Page",
			"body": map[string]interface{}{
				"storage": map[string]string{
					"value": "<p>Hello, World!</p>",
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
	defer server.Close()

	client := NewClient(server.URL, BasicAuth{Username: "user", Token: "token"})
	xhtml, info, err := client.GetPageStorageRaw(context.Background(), "12345")

	if err != nil {
		t.Fatalf("GetPageStorageRaw() error = %v", err)
	}

	if xhtml != "<p>Hello, World!</p>" {
		t.Errorf("GetPageStorageRaw() xhtml = %v, want <p>Hello, World!</p>", xhtml)
	}

	if info.ID != "12345" {
		t.Errorf("GetPageStorageRaw() info.ID = %v, want 12345", info.ID)
	}

	if info.Title != "Test Page" {
		t.Errorf("GetPageStorageRaw() info.Title = %v, want Test Page", info.Title)
	}

	if info.Version != 5 {
		t.Errorf("GetPageStorageRaw() info.Version = %v, want 5", info.Version)
	}

	if info.SpaceKey != "TEST" {
		t.Errorf("GetPageStorageRaw() info.SpaceKey = %v, want TEST", info.SpaceKey)
	}
}

func TestGetPageStorageRaw_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		if _, err := w.Write([]byte(`{"message": "Page not found"}`)); err != nil {
			panic(err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, BasicAuth{Username: "user", Token: "token"})
	_, _, err := client.GetPageStorageRaw(context.Background(), "99999")

	if err == nil {
		t.Fatal("GetPageStorageRaw() should return error for 404")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("GetPageStorageRaw() error should be *APIError, got %T", err)
	}

	if apiErr.StatusCode != 404 {
		t.Errorf("APIError.StatusCode = %v, want 404", apiErr.StatusCode)
	}
}

func TestGetPageStorage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
				"number": 1,
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
	defer server.Close()

	client := NewClient(server.URL, BasicAuth{Username: "user", Token: "token"})
	page, info, err := client.GetPageStorage(context.Background(), "12345")

	if err != nil {
		t.Fatalf("GetPageStorage() error = %v", err)
	}

	if len(page.Blocks) != 2 {
		t.Errorf("GetPageStorage() blocks = %d, want 2", len(page.Blocks))
	}

	if info.Title != "Test Page" {
		t.Errorf("GetPageStorage() info.Title = %v, want Test Page", info.Title)
	}
}

func TestUpdatePageStorageRaw(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}

		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			panic(err)
		}

		if payload["title"] != "Updated Title" {
			t.Errorf("Expected title 'Updated Title', got %v", payload["title"])
		}

		version := payload["version"].(map[string]interface{})
		if version["number"] != float64(6) {
			t.Errorf("Expected version 6, got %v", version["number"])
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"id": "12345"}`)); err != nil {
			panic(err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, BasicAuth{Username: "user", Token: "token"})
	err := client.UpdatePageStorageRaw(context.Background(), "12345", "<p>Updated content</p>", 5, "Updated Title")

	if err != nil {
		t.Fatalf("UpdatePageStorageRaw() error = %v", err)
	}
}

func TestUpdatePageStorageRaw_ValidationError(t *testing.T) {
	client := NewClient("http://example.com", BasicAuth{Username: "user", Token: "token"})
	err := client.UpdatePageStorageRaw(context.Background(), "12345", "<div>Invalid</div>", 5, "Title")

	if err == nil {
		t.Fatal("UpdatePageStorageRaw() should return validation error for forbidden tag")
	}
}

func TestUpdatePageStorage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"id": "12345"}`)); err != nil {
			panic(err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, BasicAuth{Username: "user", Token: "token"})

	page := &storage.Page{
		Blocks: []storage.Block{
			&storage.Paragraph{Text: "Updated content"},
		},
	}

	err := client.UpdatePageStorage(context.Background(), "12345", page, 5, "Updated Title")

	if err != nil {
		t.Fatalf("UpdatePageStorage() error = %v", err)
	}
}

func TestCreatePageRaw(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			panic(err)
		}

		if payload["title"] != "New Page" {
			t.Errorf("Expected title 'New Page', got %v", payload["title"])
		}

		space := payload["space"].(map[string]interface{})
		if space["key"] != "TEST" {
			t.Errorf("Expected space key 'TEST', got %v", space["key"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte(`{"id": "67890"}`)); err != nil {
			panic(err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, BasicAuth{Username: "user", Token: "token"})
	pageID, err := client.CreatePageRaw(context.Background(), "TEST", "New Page", "<p>Content</p>", "")

	if err != nil {
		t.Fatalf("CreatePageRaw() error = %v", err)
	}

	if pageID != "67890" {
		t.Errorf("CreatePageRaw() pageID = %v, want 67890", pageID)
	}
}

func TestCreatePageRaw_WithParent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			panic(err)
		}

		ancestors := payload["ancestors"].([]interface{})
		if len(ancestors) != 1 {
			t.Errorf("Expected 1 ancestor, got %d", len(ancestors))
		}

		ancestor := ancestors[0].(map[string]interface{})
		if ancestor["id"] != "11111" {
			t.Errorf("Expected parent ID '11111', got %v", ancestor["id"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte(`{"id": "67890"}`)); err != nil {
			panic(err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, BasicAuth{Username: "user", Token: "token"})
	_, err := client.CreatePageRaw(context.Background(), "TEST", "New Page", "<p>Content</p>", "11111")

	if err != nil {
		t.Fatalf("CreatePageRaw() with parent error = %v", err)
	}
}

func TestCreatePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte(`{"id": "67890"}`)); err != nil {
			panic(err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, BasicAuth{Username: "user", Token: "token"})

	page := &storage.Page{
		Blocks: []storage.Block{
			&storage.Heading{Level: 1, Text: "Welcome"},
			&storage.Paragraph{Text: "Hello, World!"},
		},
	}

	pageID, err := client.CreatePage(context.Background(), "TEST", "New Page", page, "")

	if err != nil {
		t.Fatalf("CreatePage() error = %v", err)
	}

	if pageID != "67890" {
		t.Errorf("CreatePage() pageID = %v, want 67890", pageID)
	}
}

func TestDeletePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, BasicAuth{Username: "user", Token: "token"})
	err := client.DeletePage(context.Background(), "12345")

	if err != nil {
		t.Fatalf("DeletePage() error = %v", err)
	}
}

func TestGetSpace(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/space/TEST" {
			t.Errorf("Expected path /rest/api/space/TEST, got %s", r.URL.Path)
		}

		response := SpaceInfo{
			ID:   123,
			Key:  "TEST",
			Name: "Test Space",
			Type: "global",
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			panic(err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, BasicAuth{Username: "user", Token: "token"})
	space, err := client.GetSpace(context.Background(), "TEST")

	if err != nil {
		t.Fatalf("GetSpace() error = %v", err)
	}

	if space.Key != "TEST" {
		t.Errorf("GetSpace() Key = %v, want TEST", space.Key)
	}

	if space.Name != "Test Space" {
		t.Errorf("GetSpace() Name = %v, want Test Space", space.Name)
	}
}

func TestSearchPages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	defer server.Close()

	client := NewClient(server.URL, BasicAuth{Username: "user", Token: "token"})
	pages, err := client.SearchPages(context.Background(), "space=TEST", 10)

	if err != nil {
		t.Fatalf("SearchPages() error = %v", err)
	}

	if len(pages) != 2 {
		t.Errorf("SearchPages() returned %d pages, want 2", len(pages))
	}

	if pages[0].Title != "Page 1" {
		t.Errorf("SearchPages() pages[0].Title = %v, want Page 1", pages[0].Title)
	}
}
