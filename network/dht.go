package network

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// DHT implements a simple distributed hash table for peer discovery
type DHT struct {
	nodeID      string
	port        int
	peers       map[string]*DHTNode
	buckets     map[string][]*DHTNode
	mu          sync.RWMutex
	listener    *net.UDPConn
	stopCh      chan struct{}
	peerCh      chan *DHTNode
}

type DHTNode struct {
	ID       string
	Addr     string
	Port     int
	LastSeen time.Time
}

type DHTMessage struct {
	Type     string      `json:"type"`     // "ping", "find_node", "announce", "peers"
	NodeID   string      `json:"node_id"`
	InfoHash string      `json:"info_hash,omitempty"`
	Peers    []*DHTNode  `json:"peers,omitempty"`
	Data     interface{} `json:"data,omitempty"`
}

// Public DHT bootstrap nodes (like BitTorrent uses)
var BootstrapNodes = []string{
	"router.bittorrent.com:6881",
	"dht.transmissionbt.com:6881",
	"router.utorrent.com:6881",
	"dht.libtorrent.org:25401",
	"dht.aelitis.com:6881",
}

// HashMouth-specific bootstrap nodes (you can add your own)
var HashMouthBootstrap = []string{
	// Add your own bootstrap nodes here
	// "bootstrap1.hashmouth.io:6881",
	// "bootstrap2.hashmouth.io:6881",
}

func NewDHT(port int) (*DHT, error) {
	// Generate random node ID
	nodeID := generateNodeID()

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	listener, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	dht := &DHT{
		nodeID:   nodeID,
		port:     port,
		peers:    make(map[string]*DHTNode),
		buckets:  make(map[string][]*DHTNode),
		listener: listener,
		stopCh:   make(chan struct{}),
		peerCh:   make(chan *DHTNode, 100),
	}

	go dht.listen()
	go dht.maintainPeers()

	return dht, nil
}

func generateNodeID() string {
	b := make([]byte, 20)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// Bootstrap connects to known DHT nodes
func (dht *DHT) Bootstrap() error {
	log.Printf("🌐 Bootstrapping DHT...")

	// Try HashMouth bootstrap nodes first
	for _, addr := range HashMouthBootstrap {
		if err := dht.ping(addr); err == nil {
			log.Printf("✅ Connected to HashMouth bootstrap: %s", addr)
		}
	}

	// Try public DHT bootstrap nodes
	connected := 0
	for _, addr := range BootstrapNodes {
		if err := dht.ping(addr); err == nil {
			log.Printf("✅ Connected to public DHT: %s", addr)
			connected++
			if connected >= 3 {
				break // Connect to at least 3 bootstrap nodes
			}
		}
	}

	if connected == 0 && len(HashMouthBootstrap) == 0 {
		log.Printf("⚠️  No bootstrap nodes available, running in standalone mode")
		return fmt.Errorf("no bootstrap nodes available")
	}

	// Start finding peers
	go dht.findPeers()

	return nil
}

func (dht *DHT) ping(addr string) error {
	msg := DHTMessage{
		Type:   "ping",
		NodeID: dht.nodeID,
	}

	return dht.sendMessage(addr, msg)
}

func (dht *DHT) sendMessage(addr string, msg DHTMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}

	_, err = dht.listener.WriteToUDP(data, udpAddr)
	return err
}

func (dht *DHT) listen() {
	buffer := make([]byte, 65536)

	for {
		select {
		case <-dht.stopCh:
			return
		default:
			dht.listener.SetReadDeadline(time.Now().Add(1 * time.Second))
			n, addr, err := dht.listener.ReadFromUDP(buffer)
			if err != nil {
				continue
			}

			go dht.handleMessage(buffer[:n], addr)
		}
	}
}

func (dht *DHT) handleMessage(data []byte, addr *net.UDPAddr) {
	var msg DHTMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return
	}

	switch msg.Type {
	case "ping":
		dht.handlePing(msg, addr)
	case "find_node":
		dht.handleFindNode(msg, addr)
	case "announce":
		dht.handleAnnounce(msg, addr)
	case "peers":
		dht.handlePeers(msg)
	}
}

func (dht *DHT) handlePing(msg DHTMessage, addr *net.UDPAddr) {
	// Add peer
	peer := &DHTNode{
		ID:       msg.NodeID,
		Addr:     addr.IP.String(),
		Port:     addr.Port,
		LastSeen: time.Now(),
	}

	dht.addPeer(peer)

	// Send pong
	response := DHTMessage{
		Type:   "pong",
		NodeID: dht.nodeID,
	}
	dht.sendMessage(fmt.Sprintf("%s:%d", addr.IP.String(), addr.Port), response)
}

func (dht *DHT) handleFindNode(msg DHTMessage, addr *net.UDPAddr) {
	// Return known peers
	peers := dht.getClosestPeers(msg.NodeID, 8)

	response := DHTMessage{
		Type:   "peers",
		NodeID: dht.nodeID,
		Peers:  peers,
	}
	dht.sendMessage(fmt.Sprintf("%s:%d", addr.IP.String(), addr.Port), response)
}

func (dht *DHT) handleAnnounce(msg DHTMessage, addr *net.UDPAddr) {
	// Node is announcing itself
	peer := &DHTNode{
		ID:       msg.NodeID,
		Addr:     addr.IP.String(),
		Port:     addr.Port,
		LastSeen: time.Now(),
	}

	dht.addPeer(peer)
	log.Printf("📢 Peer announced: %s (%s:%d)", peer.ID[:8], peer.Addr, peer.Port)
}

func (dht *DHT) handlePeers(msg DHTMessage) {
	// Received peer list
	for _, peer := range msg.Peers {
		peer.LastSeen = time.Now()
		dht.addPeer(peer)
		
		// Notify about new peer
		select {
		case dht.peerCh <- peer:
		default:
		}
	}
}

func (dht *DHT) addPeer(peer *DHTNode) {
	dht.mu.Lock()
	defer dht.mu.Unlock()

	key := fmt.Sprintf("%s:%d", peer.Addr, peer.Port)
	if existing, exists := dht.peers[key]; exists {
		existing.LastSeen = time.Now()
	} else {
		dht.peers[key] = peer
		log.Printf("➕ New peer discovered: %s (%s:%d)", peer.ID[:8], peer.Addr, peer.Port)
	}
}

func (dht *DHT) getClosestPeers(targetID string, count int) []*DHTNode {
	dht.mu.RLock()
	defer dht.mu.RUnlock()

	peers := make([]*DHTNode, 0, count)
	for _, peer := range dht.peers {
		if time.Since(peer.LastSeen) < 5*time.Minute {
			peers = append(peers, peer)
			if len(peers) >= count {
				break
			}
		}
	}
	return peers
}

func (dht *DHT) findPeers() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-dht.stopCh:
			return
		case <-ticker.C:
			dht.mu.RLock()
			peerList := make([]*DHTNode, 0, len(dht.peers))
			for _, peer := range dht.peers {
				peerList = append(peerList, peer)
			}
			dht.mu.RUnlock()

			// Ask random peers for more peers
			for _, peer := range peerList {
				if time.Since(peer.LastSeen) < 2*time.Minute {
					msg := DHTMessage{
						Type:   "find_node",
						NodeID: dht.nodeID,
					}
					addr := fmt.Sprintf("%s:%d", peer.Addr, peer.Port)
					dht.sendMessage(addr, msg)
				}
			}
		}
	}
}

func (dht *DHT) maintainPeers() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-dht.stopCh:
			return
		case <-ticker.C:
			dht.mu.Lock()
			// Remove stale peers
			for key, peer := range dht.peers {
				if time.Since(peer.LastSeen) > 10*time.Minute {
					delete(dht.peers, key)
					log.Printf("🧹 Removed stale peer: %s", peer.ID[:8])
				}
			}
			dht.mu.Unlock()
		}
	}
}

// Announce announces this node to the DHT
func (dht *DHT) Announce() {
	msg := DHTMessage{
		Type:   "announce",
		NodeID: dht.nodeID,
	}

	// Announce to all known peers
	dht.mu.RLock()
	peers := make([]*DHTNode, 0, len(dht.peers))
	for _, peer := range dht.peers {
		peers = append(peers, peer)
	}
	dht.mu.RUnlock()

	for _, peer := range peers {
		addr := fmt.Sprintf("%s:%d", peer.Addr, peer.Port)
		dht.sendMessage(addr, msg)
	}

	log.Printf("📢 Announced to %d peers", len(peers))
}

// GetPeers returns all known peers
func (dht *DHT) GetPeers() []*DHTNode {
	dht.mu.RLock()
	defer dht.mu.RUnlock()

	peers := make([]*DHTNode, 0, len(dht.peers))
	for _, peer := range dht.peers {
		if time.Since(peer.LastSeen) < 5*time.Minute {
			peers = append(peers, peer)
		}
	}
	return peers
}

// GetPeerChannel returns channel for new peer notifications
func (dht *DHT) GetPeerChannel() <-chan *DHTNode {
	return dht.peerCh
}

// Stop stops the DHT
func (dht *DHT) Stop() {
	close(dht.stopCh)
	dht.listener.Close()
}

// GetNodeID returns this node's ID
func (dht *DHT) GetNodeID() string {
	return dht.nodeID
}

// GetPeerCount returns the number of known peers
func (dht *DHT) GetPeerCount() int {
	dht.mu.RLock()
	defer dht.mu.RUnlock()
	return len(dht.peers)
}
