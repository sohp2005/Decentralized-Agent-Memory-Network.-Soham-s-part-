package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPPersistenceClient implements PersistenceService via HTTP.
type HTTPPersistenceClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewHTTPPersistenceClient(baseURL string) *HTTPPersistenceClient {
	return &HTTPPersistenceClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func (p *HTTPPersistenceClient) ExecuteQuery(query string, params ...any) (string, error) {
	u := fmt.Sprintf("%s/query", p.BaseURL)

	payload := map[string]any{
		"query":  query,
		"params": params,
	}
	b, _ := json.Marshal(payload)

	resp, err := p.HTTPClient.Post(u, "application/json", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("query failed: %s", string(body))
	}
	// persistence_service returns JSON string already
	return string(body), nil
}

func (p *HTTPPersistenceClient) ExecuteWrite(sqlStmt string, params ...any) error {
	u := fmt.Sprintf("%s/write", p.BaseURL)

	payload := map[string]any{
		"sql":    sqlStmt,
		"params": params,
	}
	b, _ := json.Marshal(payload)

	resp, err := p.HTTPClient.Post(u, "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("write failed: %s", string(body))
	}
	return nil
}

// AddVector sends a vector to the persistence service to be stored locally
// (SQLite + in-memory index), matching /vector/add.
func (p *HTTPPersistenceClient) AddVector(id string, vector []float32) error {
	u := fmt.Sprintf("%s/vector/add", p.BaseURL)

	payload := struct {
		ID     string    `json:"id"`
		Vector []float32 `json:"vector"`
	}{
		ID:     id,
		Vector: vector,
	}

	b, _ := json.Marshal(payload)

	resp, err := p.HTTPClient.Post(u, "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("add vector failed: %s", string(body))
	}
	return nil
}

// SearchVector calls /vector/search on the persistence service and returns
// the ids + scores from the local vector index.
func (p *HTTPPersistenceClient) SearchVector(vector []float32, topK int) ([]string, []float32, error) {
	u := fmt.Sprintf("%s/vector/search", p.BaseURL)

	payload := struct {
		Vector []float32 `json:"vector"`
		TopK   int       `json:"top_k"`
	}{
		Vector: vector,
		TopK:   topK,
	}

	b, _ := json.Marshal(payload)

	resp, err := p.HTTPClient.Post(u, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, nil, fmt.Errorf("vector search failed: %s", string(body))
	}

	var parsed struct {
		IDs    []string  `json:"ids"`
		Scores []float32 `json:"scores"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, nil, err
	}

	return parsed.IDs, parsed.Scores, nil
}
