# Decentralized Agent Memory Network (DAMN)

> **Note to Readers:** This `README` is divided into two sections. **Part 1** details the high-level architecture of the complete Decentralized Agent Memory Network (DAMN). **Part 2** specifically highlights my individual contributions and the microservices contained within this repository.

---

## Part 1: The Complete DAMN Architecture

The Decentralized Agent Memory Network (DAMN) is a distributed system that enables a large number of software agents, spread across multiple servers, to manage, share, and query their collective memory in a peer-to-peer (P2P) fashion. It operates entirely without a central coordinating authority.

### System Intuition
Think of DAMN as a hive mind where every agent holds its own piece of the puzzle, but can effortlessly query the entire network to find missing information. 

The full network relies on several core pillars:
* **P2P Networking & Discovery:** Agents discover each other using a Kademlia Distributed Hash Table (DHT) to resolve data keys to specific network addresses.
* **Inter-Agent Communication:** All peer-to-peer data access is handled via a strictly defined gRPC protocol. 
* **Cache Invalidation via Gossip:** When an agent updates its data, it broadcasts an `InvalidationNotice` across a decentralized Pub/Sub network (Gossip protocol), ensuring peer agents purge stale data from their local caches.
* **Global & Local Search:** Agents maintain their own local vector indexes for private unstructured data, but also connect to a centralized Distributed Vector Store (like Milvus) for global, network-wide semantic searches.
* **Zero-Trust Security:** Peer requests are protected by mutual TLS (mTLS) via SPIFFE for identity verification, and an Open Policy Agent (OPA) interceptor for Role-Based Access Control (RBAC).

---

## Part 2: My Contribution - Core Agent Microservices

This repository contains my specific contributions to the DAMN project: the internal microservice architecture that powers a single Server Node. 

Instead of building a monolithic agent, I designed the node's internal mechanics as decoupled, containerized Go microservices. This repository implements the local storage, caching, and routing mechanisms that allow an agent's internal logic to securely interact with the broader P2P network.

### 🏗️ Directory Structure & Services

The node is broken down into three primary microservices, orchestrated locally via Docker Compose:

#### 1. `agent_api_gateway/`
This service acts as the brain's central router. It provides a single, unified gRPC interface for the agent's internal logic to interact with all memory and network functions.
* **Key Responsibilities:**
  * Consolidates complex multi-step workflows (like cache-miss-and-fetch) into simple API calls.
  * Maintains distinct client stubs (`cache_client.go`, `persistence_client.go`, `invalidation_client.go`) to route data flow efficiently.

#### 2. `cache-service/`
A high-speed, in-memory KV caching layer designed to minimize latency when reading data owned by other agents.
* **Key Responsibilities:**
  * Drastically reduces network RPC calls by returning instant cache hits for recently fetched peer data.
  * Implements `Get`, `Put` (with TTL), and `Invalidate` endpoints triggered by the Pub/Sub gossip network.

#### 3. `persistence_service/`
The repository layer that abstracts the physical storage of the agent's private memory. 
* **Key Responsibilities:**
  * **Structured Memory:** Manages an embedded SQLite database (`persistence.db` running in WAL mode) to execute safe, injection-resistant SQL queries on private tabular data.
  * **Unstructured Memory:** Wraps an embedded local vector index (`vector_index.go` using HNSWlib) to persist high-dimensional embeddings and perform highly optimized, local nearest-neighbor semantic searches.

### 🛠️ Tech Stack
* **Language:** Go
* **Communication:** gRPC & Protocol Buffers
* **Relational Storage:** SQLite (Embedded, WAL mode)
* **Vector Search:** Local HNSW-based vector index
* **Deployment:** Docker & Docker Compose

### Getting Started

To spin up the internal agent node components and test the local microservice interactions:

```bash
# Start the API Gateway, Cache, and Persistence services locally
docker-compose up --build
```
