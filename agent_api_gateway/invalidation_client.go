package main

import "log"

// StubInvalidation implements InvalidationService but only logs.
// In the real system this will call the Pub/Sub invalidation network i hope
type StubInvalidation struct{}

func (s StubInvalidation) BroadcastInvalidation(key string) error {
	log.Printf("[GATEWAY][Invalidation] Broadcasting invalidation: %s\n", key)
	return nil // In real version, will call pubsub service
}
