# Agent API Gateway

The Agent API Gateway is the primary access point into an agent’s memory system.  
It exposes both **HTTP REST** and **gRPC** interfaces that unify:

- Structured key–value memory
- Unstructured (vector-indexed) memory
- Local cache lookups
- Local persistence
- Peer discovery + remote reads (future)
- Distributed cache invalidation (future)
- Global semantic search (future)

This service orchestrates all memory operations exactly as defined in the LLD.

---

## 🔎 Responsibilities

- Handle REST + gRPC requests from clients
- Serve structured memory reads/writes
- Serve unstructured memory storage + search
- Coordinate with:
    - Cache Service (`cache_client.go`)
    - Persistence Service (`persistence_client.go`)
- Abstract distributed components (stubs):
    - Peer fetch
    - DHT owner lookup
    - Global search
    - Invalidation broadcast

---


## Run
`go run .`