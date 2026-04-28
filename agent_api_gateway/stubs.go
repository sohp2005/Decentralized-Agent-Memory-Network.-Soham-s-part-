package main

//Check this file once pls
import "log"

// StubDiscovery implements DiscoveryService but only logs for now.
type StubDiscovery struct{}

func (s StubDiscovery) ResolveOwner(key string) (string, error) {
	log.Printf("[STUB][Discovery] ResolveOwner(%s)\n", key)
	// For now, pretend everything is local.
	return "local", nil
}

// StubPeer implements PeerService but does nothing yet.
type StubPeer struct{}

func (s StubPeer) FetchStructured(address, key string) (string, error) {
	log.Printf("[STUB][Peer] FetchStructured(addr=%s, key=%s)\n", address, key)
	// No remote fetch yet.
	return "", nil
}

// StubGlobalSearch implements GlobalSearchService.
type StubGlobalSearch struct{}

func (s StubGlobalSearch) SearchGlobal(vector []float32, topK int) ([]string, []float32, error) {
	log.Printf("[STUB][GlobalSearch] SearchGlobal(vec=%v, topK=%d)\n", vector, topK)
	// No real global DB yet; return empty result set.
	return []string{}, []float32{}, nil
}

func (s StubGlobalSearch) IndexGlobal(id string, vector []float32, ownerAgentID string) error {
	log.Printf("[STUB][GlobalSearch] IndexGlobal(id=%s, owner=%s, vec_dim=%d)\n",
		id, ownerAgentID, len(vector))
	return nil
}
