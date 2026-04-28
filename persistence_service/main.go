package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

var store *Persistence

func main() {
	var err error
	store, err = NewPersistence("persistence.db", 3)
	if err != nil {
		log.Fatalf("failed to start persistence service: %v", err)
	}
	http.HandleFunc("/health", health)
	http.HandleFunc("/query", handleQuery)
	http.HandleFunc("/write", handleWrite)
	http.HandleFunc("/vector/add", handleAddVector)
	http.HandleFunc("/vector/search", handleSearchVector)
	log.Println("💾 Persistence Service running on :8090")
	log.Fatal(http.ListenAndServe(":8090", nil))
}

func health(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintln(w, "Persistence service is healthy")
}

func handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Query  string `json:"query"`
		Params []any  `json:"params"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad json", 400)
		return
	}

	if body.Query == "" {
		http.Error(w, "missing 'query' field", 400)
		return
	}

	jsonStr, err := store.ExecuteQuery(body.Query, body.Params...)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, jsonStr)
}

func handleWrite(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		SQL    string `json:"sql"`
		Params []any  `json:"params"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad json", 400)
		return
	}

	if err := store.ExecuteWrite(body.SQL, body.Params...); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleAddVector(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		ID     string    `json:"id"`
		Vector []float32 `json:"vector"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad json", 400)
		return
	}

	if body.ID == "" {
		http.Error(w, "missing 'id' field", 400)
		return
	}

	// 1. Serialize vector to bytes
	vectorBytes, err := json.Marshal(body.Vector)
	if err != nil {
		http.Error(w, "failed to serialize vector", 500)
		return
	}

	// 2. Save to SQLite
	sqlStmt := `INSERT OR REPLACE INTO vectors (id, vector) VALUES (?, ?)`
	if err := store.ExecuteWrite(sqlStmt, body.ID, vectorBytes); err != nil {
		http.Error(w, "db write failed: "+err.Error(), 500)
		return
	}

	// 3. Add to in-memory index
	if err := store.vectors.Add(body.ID, body.Vector); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	fmt.Fprintf(w, "Vector added for %s\n", body.ID)
}

func handleSearchVector(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Vector []float32 `json:"vector"`
		TopK   int       `json:"top_k"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad json", 400)
		return
	}

	if body.TopK <= 0 {
		body.TopK = 5
	}

	ids, scores, err := store.vectors.Search(body.Vector, body.TopK)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	resp := map[string]any{
		"ids":    ids,
		"scores": scores,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
