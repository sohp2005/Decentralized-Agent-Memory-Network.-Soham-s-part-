package main

// i've mentioned fromat for the fxns in comments above it
import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// Health check: GET /health
func (g *Gateway) HealthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Agent API Gateway is healthy ✅")
}

// Structured read:
// GET /memory?key=soham
func (g *Gateway) GetStructuredHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "missing ?key=", http.StatusBadRequest)
		return
	}

	val, err := g.GetStructured(key)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(val))
}

// Structured write:
// POST /memory/update
// body: { "key":"soham", "value":"legendary" }
func (g *Gateway) StoreStructuredHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if body.Key == "" {
		http.Error(w, "missing key", http.StatusBadRequest)
		return
	}

	if err := g.StoreStructured(body.Key, body.Value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Unstructured store:
// POST /memory/unstructured
//
//	body: {
//	  "id": "doc1",
//	  "content": "raw text or JSON",
//	  "vector": [0.1, 0.2, 0.3],
//	  "index_global": true
//	}
func (g *Gateway) StoreUnstructuredHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		ID          string    `json:"id"`
		Content     string    `json:"content"`
		Vector      []float32 `json:"vector"`
		IndexGlobal bool      `json:"index_global"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(body.ID) == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	if len(body.Vector) == 0 {
		http.Error(w, "vector is required", http.StatusBadRequest)
		return
	}

	if err := g.StoreUnstructured(body.ID, body.Content, body.Vector, body.IndexGlobal); err != nil {
		log.Printf("[HANDLER] StoreUnstructured error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Unstructured search:
// POST /memory/unstructured/search
//
//	body: {
//	  "scope": "LOCAL" | "GLOBAL",
//	  "vector": [0.1, 0.2, 0.3],
//	  "top_k": 5
//	}
func (g *Gateway) SearchUnstructuredHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Scope  string    `json:"scope"`
		Vector []float32 `json:"vector"`
		TopK   int       `json:"top_k"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	if len(body.Vector) == 0 {
		http.Error(w, "vector is required", http.StatusBadRequest)
		return
	}

	ids, scores, err := g.SearchUnstructured(body.Scope, body.Vector, body.TopK)
	if err != nil {
		log.Printf("[HANDLER] SearchUnstructured error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]any{
		"ids":    ids,
		"scores": scores,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
