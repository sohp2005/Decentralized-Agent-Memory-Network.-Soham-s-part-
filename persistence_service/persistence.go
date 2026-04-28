package main

import (
	"database/sql"
	"encoding/json"
	"log"
	_ "modernc.org/sqlite"
)

type Persistence struct {
	db      *sql.DB
	vectors *VectorIndex
}

func NewPersistence(path string, vectorDim int) (*Persistence, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	// Structured KV table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS kv (
        key TEXT PRIMARY KEY,
        value TEXT
    );`)
	if err != nil {
		return nil, err
	}

	// NEW: Vector table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS vectors (
        id TEXT PRIMARY KEY,
        vector BLOB NOT NULL
    );`)
	if err != nil {
		return nil, err
	}

	// Create NEW empty in-memory vector index
	vecIndex := NewVectorIndex(vectorDim)

	// --- REHYDRATE VECTOR INDEX ON STARTUP ---
	rows, err := db.Query(`SELECT id, vector FROM vectors`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id string
		var vectorBytes []byte

		if err := rows.Scan(&id, &vectorBytes); err != nil {
			log.Printf("Failed to scan vector row: %v", err)
			continue
		}

		var vec []float32
		if err := json.Unmarshal(vectorBytes, &vec); err != nil {
			log.Printf("Deserialize error for %s: %v", id, err)
			continue
		}

		if err := vecIndex.Add(id, vec); err != nil {
			log.Printf("Failed to add vector %s: %v", id, err)
			continue
		}

		count++
	}

	log.Printf("Rehydrated %d vectors into memory.", count)

	// Return initialized persistence layer
	return &Persistence{
		db:      db,
		vectors: vecIndex,
	}, nil
}

func (p *Persistence) ExecuteWrite(sqlStmt string, args ...any) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(sqlStmt, args...); err != nil {
		return err
	}

	return tx.Commit()
}

func (p *Persistence) ExecuteQuery(query string, args ...any) (string, error) {
	rows, err := p.db.Query(query, args...)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	return RowsToJSON(rows)
}

func RowsToJSON(rows *sql.Rows) (string, error) {
	cols, err := rows.Columns()
	if err != nil {
		return "", err
	}

	results := []map[string]any{}

	for rows.Next() {
		values := make([]any, len(cols))
		pointers := make([]any, len(cols))
		for i := range values {
			pointers[i] = &values[i]
		}

		if err := rows.Scan(pointers...); err != nil {
			return "", err
		}

		obj := make(map[string]any)
		for i, col := range cols {
			obj[col] = values[i]
		}

		results = append(results, obj)
	}

	out, err := json.Marshal(results)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
