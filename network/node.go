package network

import (
	"fmt"
	"net"
	"sync"
)

// Peer represents a remote node
type Peer struct {
	ID   string
	Addr string
}

// P2PNode represents a running node
type P2PNode struct {
	ID        string
	Addr      string
	Peers     map[string]*Peer
	listener  net.Listener
	SendFunc  func(peer *Peer, data []byte)
	ReceiveCh chan []byte
	mutex     sync.Mutex
}

// NewNode creates a node with a listening port
func NewNode(id, addr string) *P2PNode {
	return &P2PNode{
		ID:        id,
		Addr:      addr,
		Peers:     make(map[string]*Peer),
		ReceiveCh: make(chan []byte, 100),
	}
}

// Start listening TCP
func (n *P2PNode) Listen() error {
	ln, err := net.Listen("tcp", n.Addr)
	if err != nil {
		return err
	}
	n.listener = ln
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				continue
			}
			go n.handleConn(conn)
		}
	}()
	return nil
}

func (n *P2PNode) handleConn(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 65535)
	for {
		nRead, err := conn.Read(buf)
		if err != nil {
			return
		}
		data := make([]byte, nRead)
		copy(data, buf[:nRead])
		n.ReceiveCh <- data
	}
}

// Connect to peer
func (n *P2PNode) ConnectPeer(id, addr string) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.Peers[id] = &Peer{ID: id, Addr: addr}
}

// SendMessage sends raw bytes to a peer
func (n *P2PNode) SendMessage(peer *Peer, data []byte) {
	go func() {
		conn, err := net.Dial("tcp", peer.Addr)
		if err != nil {
			fmt.Printf("[%s] failed to connect to %s: %v\n", n.ID, peer.ID, err)
			return
		}
		defer conn.Close()
		conn.Write(data)
	}()
}
