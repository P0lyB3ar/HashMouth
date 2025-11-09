package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"hashmouth/crypto"
	"hashmouth/network"
	"time"
)

var nodeKeys map[string][]byte

func main() {
	id := flag.String("id", "NodeA", "Node ID")
	addr := flag.String("listen", ":9000", "Listen address")
	flag.Parse()

	// Generate key
	nodeKeys = make(map[string][]byte)
	key, _ := crypto.GenerateSymmetricKey()
	nodeKeys[*id] = key

	node := network.NewNode(*id, *addr)
	node.Listen()
	fmt.Printf("[%s] Listening on %s\n", *id, *addr)

	// Example: connect to peers (manually for now)
	// node.ConnectPeer("NodeB0", "localhost:9001")

	// Receive loop
	for {
		select {
		case msg := <-node.ReceiveCh:
			fmt.Printf("[%s] Received raw: %x\n", *id, msg)
			pkt, err := crypto.Deserialize(msg)
			if err != nil {
				fmt.Println("deserialize error:", err)
				continue
			}
			plain, err := crypto.PeelOnion(pkt, nodeKeys[*id])
			if err != nil {
				fmt.Println("peel error:", err)
				continue
			}
			fmt.Printf("[%s] Decrypted: %s\n", *id, string(plain))
		}
	}
}

// Helper to create onion packet for a given key
func mustCreateOnionPacket(data, key []byte) *crypto.OnionPacket {
	pkt, err := crypto.CreateOnionPacket(data, key)
	if err != nil {
		panic(err)
	}
	return pkt
}

// Send a message to a path of nodes
func sendOnion(node *network.P2PNode, path []string, data []byte) {
	current := data
	for i := len(path) - 1; i >= 0; i-- {
		k := nodeKeys[path[i]]
		pkt := mustCreateOnionPacket(current, k)
		current = pkt.Serialize()
	}
	peerID := path[0]
	if peer, ok := node.Peers[peerID]; ok {
		node.SendMessage(peer, current)
	} else {
		fmt.Printf("[%s] No peer %s\n", node.ID, peerID)
	}
}

// Example chunk struct
type MessageChunk struct {
	Seq  int
	Data []byte
}

// Pack message chunk
func packChunk(seq int, text string) []byte {
	ch := MessageChunk{
		Seq:  seq,
		Data: []byte(text),
	}
	data, _ := json.Marshal(ch)
	return data
}
