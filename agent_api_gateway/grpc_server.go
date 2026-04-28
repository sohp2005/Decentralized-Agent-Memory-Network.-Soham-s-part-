package main

import (
	"context"
	"log"

	pb "agent_api_gateway/api/memoryv1" // update to your actual module path
)

// MemoryServiceServer implements memory.v1.MemoryService using the Gateway.
type MemoryServiceServer struct {
	pb.UnimplementedMemoryServiceServer
	gw *Gateway
}

// NewMemoryServiceServer constructs a gRPC server backed by the Gateway.
func NewMemoryServiceServer(gw *Gateway) *MemoryServiceServer {
	return &MemoryServiceServer{gw: gw}
}

// GetStructured maps directly to Gateway.GetStructured.
func (s *MemoryServiceServer) GetStructured(
	ctx context.Context,
	req *pb.GetStructuredRequest,
) (*pb.GetStructuredResponse, error) {
	val, err := s.gw.GetStructured(req.GetKey())
	if err != nil {
		return nil, err
	}
	return &pb.GetStructuredResponse{
		JsonValue: val,
	}, nil
}

// StoreStructured maps directly to Gateway.StoreStructured.
func (s *MemoryServiceServer) StoreStructured(
	ctx context.Context,
	req *pb.StoreStructuredRequest,
) (*pb.StoreStructuredResponse, error) {
	if err := s.gw.StoreStructured(req.GetKey(), req.GetValue()); err != nil {
		return nil, err
	}
	return &pb.StoreStructuredResponse{}, nil
}

// StoreUnstructured calls Gateway.StoreUnstructured.
func (s *MemoryServiceServer) StoreUnstructured(
	ctx context.Context,
	req *pb.StoreUnstructuredRequest,
) (*pb.StoreUnstructuredResponse, error) {
	// Convert repeated float to []float32 (proto already uses float32).
	vector := make([]float32, len(req.Vector))
	copy(vector, req.Vector)

	err := s.gw.StoreUnstructured(
		req.GetId(),
		req.GetContent(),
		vector,
		req.GetIndexGlobal(),
	)
	if err != nil {
		return nil, err
	}
	return &pb.StoreUnstructuredResponse{}, nil
}

// SearchUnstructured chooses LOCAL / GLOBAL via Gateway.SearchUnstructured.
func (s *MemoryServiceServer) SearchUnstructured(
	ctx context.Context,
	req *pb.SearchUnstructuredRequest,
) (*pb.SearchUnstructuredResponse, error) {
	vector := make([]float32, len(req.Vector))
	copy(vector, req.Vector)

	ids, scores, err := s.gw.SearchUnstructured(req.GetScope(), vector, int(req.GetTopK()))
	if err != nil {
		return nil, err
	}
	return &pb.SearchUnstructuredResponse{
		Ids:    ids,
		Scores: scores,
	}, nil
}

// kinda optional... auto recommended
func (s *MemoryServiceServer) MustLogReady(addr string) {
	log.Printf("[gRPC] MemoryService listening on %s", addr)
}
