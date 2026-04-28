# Persistence Service

The Persistence Service provides long-term, durable storage for both:

- Structured memory (key–value)
- Unstructured memory (text + vector embeddings)

Backed by **SQLite** + a custom in-process vector search index.

This service ensures memory survives restarts and powers local semantic search.

---

## 🔎 Responsibilities

### Structured Memory
- Upsert key–value data
- Fetch values on cache miss

### Unstructured Memory
- Store content + embedding vectors
- Perform LOCAL vector search
- Support global search orchestration (future)


## Run
`go run .`
