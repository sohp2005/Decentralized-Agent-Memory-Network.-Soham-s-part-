package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

var cache *Cache

func handlePut(w http.ResponseWriter, r *http.Request) {
	var body map[string]string
	json.NewDecoder(r.Body).Decode(&body)
	key := body["key"]
	value := body["value"]
	cache.Put(key, value, 5*time.Minute)
	fmt.Fprintf(w, "Stored %s = %s\n", key, value)
}
func handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	value, ok := cache.Get(key)
	if !ok {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, "%s\n", value)
}
func handleInvalidate(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	cache.Invalidate(key)
	fmt.Fprintf(w, "Invalidated %s\n", key)
}
func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Cache Service is healthy")
}

func main() {
	cache = NewCache()
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/get", handleGet)
	http.HandleFunc("/put", handlePut)
	http.HandleFunc("/invalidate", handleInvalidate)
	port := ":8080"
	log.Printf("Cache Service starting on %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
