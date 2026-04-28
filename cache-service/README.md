


# Cache Service



# Cache Service (Ristretto)

The Cache Service is a high-performance local cache built on **Ristretto**.  
It exposes a minimal REST interface used exclusively by the Agent API Gateway.

This service is stateless except for in-memory data, and is restarted safely.

---

## 🔎 Responsibilities

- Store structured memory in hot local cache  
- Serve read requests at microsecond latency  
- Invalidate keys on broadcast  
- Reduce load on Persistence Service  
- Improve read throughput under high concurrency


## Run
`go run .`