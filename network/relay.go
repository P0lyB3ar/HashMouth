package network

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"
)

// RelayNode represents a node that can relay messages
type RelayNode struct {
	ID           string
	Addr         string
	LastSeen     time.Time
	Reliability  float64 // 0.0 to 1.0
	IsRelay      bool    // Willing to relay for others
}

// RelayNetwork manages the relay network
type RelayNetwork struct {
	relayNodes map[string]*RelayNode
	mu         sync.RWMutex
}

// RelayMessage wraps a message with routing info
type RelayMessage struct {
	MessageID   string   `json:"message_id"`
	NextHop     string   `json:"next_hop"`      // Next node in the path
	FinalDest   string   `json:"final_dest"`    // Ultimate destination
	HopsLeft    int      `json:"hops_left"`     // Remaining hops
	Payload     []byte   `json:"payload"`       // Encrypted payload
	Path        []string `json:"path,omitempty"` // For debugging (remove in production)
	Timestamp   int64    `json:"timestamp"`
}

// NewRelayNetwork creates a new relay network
func NewRelayNetwork() *RelayNetwork {
	return &RelayNetwork{
		relayNodes: make(map[string]*RelayNode),
	}
}

// RegisterRelayNode adds a node as available relay
func (rn *RelayNetwork) RegisterRelayNode(id, addr string) {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	
	rn.relayNodes[id] = &RelayNode{
		ID:          id,
		Addr:        addr,
		LastSeen:    time.Now(),
		Reliability: 1.0,
		IsRelay:     true,
	}
	log.Printf("🔄 Registered relay node: %s", id)
}

// UnregisterRelayNode removes a relay node
func (rn *RelayNetwork) UnregisterRelayNode(id string) {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	delete(rn.relayNodes, id)
	log.Printf("❌ Unregistered relay node: %s", id)
}

// GetRelayNodes returns all available relay nodes
func (rn *RelayNetwork) GetRelayNodes() []*RelayNode {
	rn.mu.RLock()
	defer rn.mu.RUnlock()
	
	nodes := make([]*RelayNode, 0, len(rn.relayNodes))
	for _, node := range rn.relayNodes {
		if node.IsRelay && time.Since(node.LastSeen) < 5*time.Minute {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// BuildRelayPath creates a random path through relay nodes
func (rn *RelayNetwork) BuildRelayPath(minHops, maxHops int, excludeNodes []string) ([]string, error) {
	rn.mu.RLock()
	defer rn.mu.RUnlock()
	
	// Filter available nodes
	available := make([]string, 0)
	excludeMap := make(map[string]bool)
	for _, node := range excludeNodes {
		excludeMap[node] = true
	}
	
	for id, node := range rn.relayNodes {
		if !excludeMap[id] && node.IsRelay && time.Since(node.LastSeen) < 5*time.Minute {
			available = append(available, id)
		}
	}
	
	if len(available) < minHops {
		return nil, errors.New("not enough relay nodes available")
	}
	
	// Determine path length
	pathLength := minHops
	if maxHops > minHops && len(available) >= maxHops {
		rangeVal := maxHops - minHops + 1
		offset, _ := rand.Int(rand.Reader, big.NewInt(int64(rangeVal)))
		pathLength = minHops + int(offset.Int64())
	}
	
	if pathLength > len(available) {
		pathLength = len(available)
	}
	
	// Select random nodes
	path := make([]string, 0, pathLength)
	used := make(map[int]bool)
	
	for len(path) < pathLength {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(available))))
		index := int(idx.Int64())
		
		if !used[index] {
			used[index] = true
			path = append(path, available[index])
		}
	}
	
	return path, nil
}

// CreateRelayMessage creates a message to be relayed
func CreateRelayMessage(finalDest string, payload []byte, path []string) (*RelayMessage, error) {
	if len(path) == 0 {
		return nil, errors.New("path cannot be empty")
	}
	
	msgID := generateMessageID()
	
	return &RelayMessage{
		MessageID: msgID,
		NextHop:   path[0],
		FinalDest: finalDest,
		HopsLeft:  len(path),
		Payload:   payload,
		Path:      path, // For debugging
		Timestamp: time.Now().Unix(),
	}, nil
}

// ProcessRelayMessage handles an incoming relay message
func (rn *RelayNetwork) ProcessRelayMessage(msg *RelayMessage, currentNodeID string) (*RelayMessage, bool, error) {
	// Check if we're the final destination
	if msg.FinalDest == currentNodeID {
		log.Printf("📬 Received message at final destination: %s", currentNodeID)
		return msg, true, nil // true = final destination
	}
	
	// Check if we should relay
	if msg.HopsLeft <= 0 {
		return nil, false, errors.New("message exceeded hop limit")
	}
	
	// Update for next hop
	msg.HopsLeft--
	
	// Find next hop in path
	if len(msg.Path) > 0 {
		// Remove current hop from path
		for i, node := range msg.Path {
			if node == currentNodeID && i+1 < len(msg.Path) {
				msg.NextHop = msg.Path[i+1]
				break
			}
		}
	}
	
	log.Printf("🔄 Relaying message %s to %s (hops left: %d)", msg.MessageID, msg.NextHop, msg.HopsLeft)
	return msg, false, nil // false = not final destination, keep relaying
}

// Serialize converts relay message to JSON
func (rm *RelayMessage) Serialize() ([]byte, error) {
	return json.Marshal(rm)
}

// DeserializeRelayMessage converts JSON to relay message
func DeserializeRelayMessage(data []byte) (*RelayMessage, error) {
	var msg RelayMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// generateMessageID creates a unique message ID
func generateMessageID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// UpdateNodeStatus updates the last seen time for a node
func (rn *RelayNetwork) UpdateNodeStatus(nodeID string) {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	
	if node, exists := rn.relayNodes[nodeID]; exists {
		node.LastSeen = time.Now()
	}
}

// GetRelayNodeAddr gets the address of a relay node
func (rn *RelayNetwork) GetRelayNodeAddr(nodeID string) (string, error) {
	rn.mu.RLock()
	defer rn.mu.RUnlock()
	
	if node, exists := rn.relayNodes[nodeID]; exists {
		return node.Addr, nil
	}
	return "", errors.New("relay node not found")
}

// CleanupStaleNodes removes nodes that haven't been seen recently
func (rn *RelayNetwork) CleanupStaleNodes() {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	
	cutoff := time.Now().Add(-10 * time.Minute)
	for id, node := range rn.relayNodes {
		if node.LastSeen.Before(cutoff) {
			delete(rn.relayNodes, id)
			log.Printf("🧹 Cleaned up stale relay node: %s", id)
		}
	}
}

// StartCleanupRoutine starts periodic cleanup of stale nodes
func (rn *RelayNetwork) StartCleanupRoutine() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		
		for range ticker.C {
			rn.CleanupStaleNodes()
		}
	}()
}
