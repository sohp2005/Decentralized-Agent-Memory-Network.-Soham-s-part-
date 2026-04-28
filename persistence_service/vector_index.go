package main

import (
	"errors"
	"math"
	"sort"
	"sync"
)

// VectorIndex is a simple in-memory vector index.=
type VectorIndex struct {
	dim  int
	mu   sync.RWMutex
	data map[string][]float32
}

func NewVectorIndex(dim int) *VectorIndex {
	return &VectorIndex{
		dim:  dim,
		data: make(map[string][]float32),
	}
}

func (v *VectorIndex) Add(id string, vector []float32) error {
	if len(vector) != v.dim {
		return errors.New("vector dimension mismatch")
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	v.data[id] = vector
	return nil
}

func (v *VectorIndex) Search(query []float32, topK int) ([]string, []float32, error) {
	if len(query) != v.dim {
		return nil, nil, errors.New("query dimension mismatch")
	}

	v.mu.RLock()
	defer v.mu.RUnlock()

	type result struct {
		id    string
		score float32
	}

	var results []result

	for id, vec := range v.data {
		score := cosineSimilarity(vec, query)
		results = append(results, result{id: id, score: score})
	}

	//sort by descending similarity
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	if topK > len(results) {
		topK = len(results)
	}

	ids := make([]string, topK)
	scores := make([]float32, topK)

	for i := 0; i < topK; i++ {
		ids[i] = results[i].id
		scores[i] = results[i].score
	}

	return ids, scores, nil
}

func cosineSimilarity(a, b []float32) float32 {
	var dot, magA, magB float32

	for i := range a {
		dot += a[i] * b[i]
		magA += a[i] * a[i]
		magB += b[i] * b[i]
	}
	if magA == 0 || magB == 0 {
		return 0
	}
	return dot / (float32(math.Sqrt(float64(magA))) * float32(math.Sqrt(float64(magB))))
}
