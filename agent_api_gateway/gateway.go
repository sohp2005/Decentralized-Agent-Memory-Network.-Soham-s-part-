package main

import (
	"fmt"
	"log"
	"strings"
)

// ---- Local Services (your microservices) ----

// CacheService talks to the cache-service HTTP API.
type CacheService interface {
	Get(key string) (string, bool)
	Put(key, value string) error
	Invalidate(key string) error
}

// PersistenceService talks to the persistence_service HTTP API.
type PersistenceService interface {
	// Structured
	ExecuteQuery(query string, params ...any) (string, error)
	ExecuteWrite(sqlStmt string, params ...any) error

	// Unstructured (vector) operations
	AddVector(id string, vector []float32) error
	SearchVector(vector []float32, topK int) ([]string, []float32, error)
}

// ---- Distributed Services (stubs for now) ----

// DiscoveryService will eventually use DHT to resolve key -> owner agent.
type DiscoveryService interface {
	ResolveOwner(key string) (string, error)
}

// PeerService will eventually do gRPC calls to other agents.
type PeerService interface {
	FetchStructured(address, key string) (string, error)
}

// InvalidationService will eventually broadcast invalidations over pub/sub.
type InvalidationService interface {
	BroadcastInvalidation(key string) error
}

// GlobalSearchService will eventually hit the global vector DB.
type GlobalSearchService interface {
	// Global semantic search
	SearchGlobal(vector []float32, topK int) ([]string, []float32, error)
	// Index new vector globally
	IndexGlobal(id string, vector []float32, ownerAgentID string) error
}

// Gateway is the LLD-style orchestrator for an agent.
type Gateway struct {
	Cache        CacheService
	Persistence  PersistenceService
	Discovery    DiscoveryService
	Peer         PeerService
	Invalidation InvalidationService
	GlobalSearch GlobalSearchService

	AgentID string // this agent's ID, used for global indexing metadata
}

// NewGateway wires all the dependencies together.
func NewGateway(
	cache CacheService,
	persistence PersistenceService,
	discovery DiscoveryService,
	peer PeerService,
	invalidation InvalidationService,
	global GlobalSearchService,
	agentID string,
) *Gateway {
	return &Gateway{
		Cache:        cache,
		Persistence:  persistence,
		Discovery:    discovery,
		Peer:         peer,
		Invalidation: invalidation,
		GlobalSearch: global,
		AgentID:      agentID,
	}
}

//strucured memory

func (g *Gateway) GetStructured(key string) (string, error) {
	// 1. Cache check
	if val, ok := g.Cache.Get(key); ok {
		return val, nil
	}

	// 2. Resolve owner via discovery (DHT)
	owner, err := g.Discovery.ResolveOwner(key)
	if err != nil {
		return "", err
	}

	var jsonResult string

	// 3. Local vs remote
	if owner == "" || owner == "local" {
		const q = `SELECT value FROM kv WHERE key = ?`
		jsonResult, err = g.Persistence.ExecuteQuery(q, key)
		if err != nil {
			return "", err
		}
	} else {
		jsonResult, err = g.Peer.FetchStructured(owner, key)
		if err != nil {
			return "", err
		}
	}

	if jsonResult == "" {
		return "", fmt.Errorf("not found")
	}

	// 4. Cache populate
	if err := g.Cache.Put(key, jsonResult); err != nil {
		log.Printf("[GATEWAY] Cache put failed for %s: %v", key, err)
	}

	return jsonResult, nil
}

func (g *Gateway) StoreStructured(key, value string) error {
	const stmt = `INSERT OR REPLACE INTO kv (key, value) VALUES (?, ?)`

	// 1. Local DB write
	if err := g.Persistence.ExecuteWrite(stmt, key, value); err != nil {
		return err
	}

	// 2. Local cache invalidate
	if err := g.Cache.Invalidate(key); err != nil {
		log.Printf("[GATEWAY] Cache invalidate failed for %s: %v", key, err)
	}

	// 3. Distributed invalidation (best-effort)
	if err := g.Invalidation.BroadcastInvalidation(key); err != nil {
		log.Printf("[GATEWAY] BroadcastInvalidation failed for %s: %v", key, err)
	}

	return nil
}

// ApplyInvalidation is what the invalidation network should call on this agent.
func (g *Gateway) ApplyInvalidation(key string) {
	log.Printf("[GATEWAY] Applying invalidation for key=%s", key)
	if err := g.Cache.Invalidate(key); err != nil {
		log.Printf("[GATEWAY] Cache invalidate failed in ApplyInvalidation: %v", err)
	}
}

//UNSTRUCTURED MEMORY (VECTORS)

func (g *Gateway) StoreUnstructured(id, content string, vector []float32, indexGlobal bool) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}

	// 1. Structured part: store the content in kv as a simple doc table.
	if content != "" {
		const stmt = `INSERT OR REPLACE INTO kv (key, value) VALUES (?, ?)`
		if err := g.Persistence.ExecuteWrite(stmt, id, content); err != nil {
			return err
		}
	}

	// 2. Local vector index.
	if len(vector) > 0 {
		if err := g.Persistence.AddVector(id, vector); err != nil {
			return err
		}
	}

	// 3. Global indexing (best-effort).
	if indexGlobal && len(vector) > 0 && g.GlobalSearch != nil {
		if err := g.GlobalSearch.IndexGlobal(id, vector, g.AgentID); err != nil {
			log.Printf("[GATEWAY] Global index failed for %s: %v", id, err)
		}
	}

	return nil
}
func (g *Gateway) SearchUnstructured(scope string, queryVector []float32, topK int) ([]string, []float32, error) {
	if len(queryVector) == 0 {
		return nil, nil, fmt.Errorf("query vector is empty")
	}
	if topK <= 0 {
		topK = 5
	}

	switch strings.ToUpper(strings.TrimSpace(scope)) {
	case "", "LOCAL":
		// Local vector search via persistence service.
		return g.Persistence.SearchVector(queryVector, topK)

	case "GLOBAL":
		if g.GlobalSearch == nil {
			return nil, nil, fmt.Errorf("global search service not configured")
		}
		return g.GlobalSearch.SearchGlobal(queryVector, topK)

	default:
		return nil, nil, fmt.Errorf("unknown scope %q (expected LOCAL or GLOBAL)", scope)
	}
}
