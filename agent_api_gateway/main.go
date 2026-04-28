package main

import (
	"log"
	"net"
	"net/http"
	"os"

	"google.golang.org/grpc"

	pb "agent_api_gateway/api/memoryv1"
	// update path after generating proto
)

func main() {
	cacheURL := getenv("CACHE_URL", "http://localhost:8080")
	persistURL := getenv("PERSIST_URL", "http://localhost:8090")
	httpPort := getenv("GATEWAY_HTTP_PORT", "8081")
	grpcPort := getenv("GATEWAY_GRPC_PORT", "50051")
	agentID := getenv("AGENT_ID", "agent-1")

	log.Printf("Starting Agent API Gateway (agent_id=%s)", agentID)
	log.Printf("HTTP on :%s, gRPC on :%s", httpPort, grpcPort)
	log.Printf("Using Cache Service at %s", cacheURL)
	log.Printf("Using Persistence Service at %s", persistURL)

	cacheClient := NewHTTPCacheClient(cacheURL)
	persistClient := NewHTTPPersistenceClient(persistURL)

	gw := NewGateway(
		cacheClient,
		persistClient,
		StubDiscovery{},
		StubPeer{},
		StubInvalidation{},
		StubGlobalSearch{},
		agentID,
	)

	//gRPC server (MemoryService)
	go func() {
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			log.Fatalf("failed to listen on gRPC port %s: %v", grpcPort, err)
		}

		grpcServer := grpc.NewServer()
		memSrv := NewMemoryServiceServer(gw)
		pb.RegisterMemoryServiceServer(grpcServer, memSrv)

		log.Printf("[gRPC] MemoryService listening on :%s", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	//HTTP REST server
	http.HandleFunc("/health", gw.HealthHandler)

	// Structured memory
	http.HandleFunc("/memory", gw.GetStructuredHandler)          // GET /memory?key=...
	http.HandleFunc("/memory/update", gw.StoreStructuredHandler) // POST /memory/update

	// Unstructured memory
	http.HandleFunc("/memory/unstructured", gw.StoreUnstructuredHandler)
	http.HandleFunc("/memory/unstructured/search", gw.SearchUnstructuredHandler)

	log.Printf("[HTTP] REST listening on :%s", httpPort)
	if err := http.ListenAndServe(":"+httpPort, nil); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
