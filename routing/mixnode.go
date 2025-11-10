package routing

import (
	"crypto/rand"
	"errors"
	"math/big"
	"sync"
	"time"
)

// MixNode represents a node that mixes and delays packets for anonymity
type MixNode struct {
	ID            string
	mu            sync.Mutex
	packetQueue   [][]byte
	maxQueueSize  int
	minDelay      time.Duration
	maxDelay      time.Duration
	batchSize     int
	processingCh  chan []byte
	outputCh      chan []byte
	stopCh        chan struct{}
}

// NewMixNode creates a new mix node
func NewMixNode(id string, maxQueueSize, batchSize int, minDelay, maxDelay time.Duration) (*MixNode, error) {
	if maxQueueSize <= 0 {
		return nil, errors.New("max queue size must be positive")
	}
	if batchSize <= 0 {
		return nil, errors.New("batch size must be positive")
	}
	if minDelay < 0 || maxDelay < minDelay {
		return nil, errors.New("invalid delay configuration")
	}

	return &MixNode{
		ID:           id,
		packetQueue:  make([][]byte, 0, maxQueueSize),
		maxQueueSize: maxQueueSize,
		minDelay:     minDelay,
		maxDelay:     maxDelay,
		batchSize:    batchSize,
		processingCh: make(chan []byte, maxQueueSize),
		outputCh:     make(chan []byte, maxQueueSize),
		stopCh:       make(chan struct{}),
	}, nil
}

// Start begins processing packets
func (mn *MixNode) Start() {
	go mn.processLoop()
	go mn.batchLoop()
}

// Stop stops the mix node
func (mn *MixNode) Stop() {
	close(mn.stopCh)
}

// AddPacket adds a packet to the mix node queue
func (mn *MixNode) AddPacket(packet []byte) error {
	mn.mu.Lock()
	defer mn.mu.Unlock()

	if len(mn.packetQueue) >= mn.maxQueueSize {
		return errors.New("queue is full")
	}

	mn.packetQueue = append(mn.packetQueue, packet)
	return nil
}

// GetOutput returns the output channel for processed packets
func (mn *MixNode) GetOutput() <-chan []byte {
	return mn.outputCh
}

// processLoop handles individual packet processing with delays
func (mn *MixNode) processLoop() {
	for {
		select {
		case <-mn.stopCh:
			return
		case packet := <-mn.processingCh:
			// Apply random delay
			delay := mn.randomDelay()
			time.Sleep(delay)
			mn.outputCh <- packet
		}
	}
}

// batchLoop processes packets in batches
func (mn *MixNode) batchLoop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-mn.stopCh:
			return
		case <-ticker.C:
			mn.processBatch()
		}
	}
}

// processBatch takes a batch of packets and shuffles them
func (mn *MixNode) processBatch() {
	mn.mu.Lock()
	if len(mn.packetQueue) == 0 {
		mn.mu.Unlock()
		return
	}

	// Determine batch size
	batchSize := mn.batchSize
	if batchSize > len(mn.packetQueue) {
		batchSize = len(mn.packetQueue)
	}

	// Extract batch
	batch := make([][]byte, batchSize)
	copy(batch, mn.packetQueue[:batchSize])
	mn.packetQueue = mn.packetQueue[batchSize:]
	mn.mu.Unlock()

	// Shuffle batch
	shuffled, err := mn.shuffleBatch(batch)
	if err != nil {
		// On error, just process in order
		shuffled = batch
	}

	// Send to processing channel
	for _, packet := range shuffled {
		select {
		case mn.processingCh <- packet:
		case <-mn.stopCh:
			return
		}
	}
}

// shuffleBatch randomly shuffles a batch of packets
func (mn *MixNode) shuffleBatch(batch [][]byte) ([][]byte, error) {
	shuffled := make([][]byte, len(batch))
	copy(shuffled, batch)

	// Fisher-Yates shuffle
	for i := len(shuffled) - 1; i > 0; i-- {
		jBig, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return nil, err
		}
		j := int(jBig.Int64())
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled, nil
}

// randomDelay generates a random delay between min and max
func (mn *MixNode) randomDelay() time.Duration {
	if mn.minDelay == mn.maxDelay {
		return mn.minDelay
	}

	delayRange := mn.maxDelay - mn.minDelay
	randomOffset, err := rand.Int(rand.Reader, big.NewInt(int64(delayRange)))
	if err != nil {
		return mn.minDelay
	}

	return mn.minDelay + time.Duration(randomOffset.Int64())
}

// QueueSize returns the current queue size
func (mn *MixNode) QueueSize() int {
	mn.mu.Lock()
	defer mn.mu.Unlock()
	return len(mn.packetQueue)
}

// Stats returns statistics about the mix node
type MixNodeStats struct {
	QueueSize     int
	MaxQueueSize  int
	BatchSize     int
	MinDelay      time.Duration
	MaxDelay      time.Duration
	ProcessedChan int
	OutputChan    int
}

// GetStats returns current statistics
func (mn *MixNode) GetStats() MixNodeStats {
	mn.mu.Lock()
	defer mn.mu.Unlock()

	return MixNodeStats{
		QueueSize:     len(mn.packetQueue),
		MaxQueueSize:  mn.maxQueueSize,
		BatchSize:     mn.batchSize,
		MinDelay:      mn.minDelay,
		MaxDelay:      mn.maxDelay,
		ProcessedChan: len(mn.processingCh),
		OutputChan:    len(mn.outputCh),
	}
}

// MixNetwork represents a network of mix nodes
type MixNetwork struct {
	nodes map[string]*MixNode
	mu    sync.RWMutex
}

// NewMixNetwork creates a new mix network
func NewMixNetwork() *MixNetwork {
	return &MixNetwork{
		nodes: make(map[string]*MixNode),
	}
}

// AddNode adds a mix node to the network
func (mn *MixNetwork) AddNode(node *MixNode) error {
	mn.mu.Lock()
	defer mn.mu.Unlock()

	if _, exists := mn.nodes[node.ID]; exists {
		return errors.New("node already exists")
	}

	mn.nodes[node.ID] = node
	node.Start()
	return nil
}

// RemoveNode removes a mix node from the network
func (mn *MixNetwork) RemoveNode(nodeID string) error {
	mn.mu.Lock()
	defer mn.mu.Unlock()

	node, exists := mn.nodes[nodeID]
	if !exists {
		return errors.New("node not found")
	}

	node.Stop()
	delete(mn.nodes, nodeID)
	return nil
}

// GetNode retrieves a mix node by ID
func (mn *MixNetwork) GetNode(nodeID string) (*MixNode, error) {
	mn.mu.RLock()
	defer mn.mu.RUnlock()

	node, exists := mn.nodes[nodeID]
	if !exists {
		return nil, errors.New("node not found")
	}

	return node, nil
}

// GetAllNodeIDs returns all node IDs in the network
func (mn *MixNetwork) GetAllNodeIDs() []string {
	mn.mu.RLock()
	defer mn.mu.RUnlock()

	ids := make([]string, 0, len(mn.nodes))
	for id := range mn.nodes {
		ids = append(ids, id)
	}
	return ids
}

// NodeCount returns the number of nodes in the network
func (mn *MixNetwork) NodeCount() int {
	mn.mu.RLock()
	defer mn.mu.RUnlock()
	return len(mn.nodes)
}
