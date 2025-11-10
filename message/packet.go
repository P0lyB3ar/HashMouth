package message

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"time"
)

// PacketType defines the type of packet
type PacketType int

const (
	PacketTypeData PacketType = iota
	PacketTypeAck
	PacketTypeHandshake
	PacketTypeKeyExchange
)

// Packet represents a network packet with metadata
type Packet struct {
	Type      PacketType `json:"type"`
	Sender    string     `json:"sender"`     // Sender ID
	Recipient string     `json:"recipient"`  // Recipient ID
	Timestamp int64      `json:"timestamp"`  // Unix timestamp
	Nonce     []byte     `json:"nonce"`      // Random nonce for replay protection
	Payload   []byte     `json:"payload"`    // Encrypted payload
	Signature []byte     `json:"signature"`  // Ed25519 signature
}

// NewPacket creates a new packet
func NewPacket(pktType PacketType, sender, recipient string, payload []byte) *Packet {
	return &Packet{
		Type:      pktType,
		Sender:    sender,
		Recipient: recipient,
		Timestamp: time.Now().Unix(),
		Payload:   payload,
	}
}

// Sign signs the packet with a private key
func (p *Packet) Sign(privateKey ed25519.PrivateKey) error {
	if len(privateKey) != ed25519.PrivateKeySize {
		return errors.New("invalid private key size")
	}

	// Create signature over all fields except the signature itself
	data, err := p.signableData()
	if err != nil {
		return err
	}

	p.Signature = ed25519.Sign(privateKey, data)
	return nil
}

// Verify verifies the packet signature
func (p *Packet) Verify(publicKey ed25519.PublicKey) error {
	if len(publicKey) != ed25519.PublicKeySize {
		return errors.New("invalid public key size")
	}
	if len(p.Signature) == 0 {
		return errors.New("packet is not signed")
	}

	data, err := p.signableData()
	if err != nil {
		return err
	}

	if !ed25519.Verify(publicKey, data, p.Signature) {
		return errors.New("signature verification failed")
	}

	return nil
}

// signableData returns the data that should be signed
func (p *Packet) signableData() ([]byte, error) {
	// Create a copy without signature
	temp := &Packet{
		Type:      p.Type,
		Sender:    p.Sender,
		Recipient: p.Recipient,
		Timestamp: p.Timestamp,
		Nonce:     p.Nonce,
		Payload:   p.Payload,
	}
	return json.Marshal(temp)
}

// Serialize converts packet to JSON bytes
func (p *Packet) Serialize() ([]byte, error) {
	return json.Marshal(p)
}

// DeserializePacket converts JSON bytes back to Packet
func DeserializePacket(data []byte) (*Packet, error) {
	var packet Packet
	if err := json.Unmarshal(data, &packet); err != nil {
		return nil, err
	}
	return &packet, nil
}

// Validate checks if the packet is valid
func (p *Packet) Validate() error {
	if p.Sender == "" {
		return errors.New("sender cannot be empty")
	}
	if p.Recipient == "" {
		return errors.New("recipient cannot be empty")
	}
	if p.Timestamp == 0 {
		return errors.New("timestamp cannot be zero")
	}
	if len(p.Payload) == 0 {
		return errors.New("payload cannot be empty")
	}
	return nil
}

// IsExpired checks if the packet is too old (replay protection)
func (p *Packet) IsExpired(maxAge time.Duration) bool {
	packetTime := time.Unix(p.Timestamp, 0)
	return time.Since(packetTime) > maxAge
}

// PacketQueue manages a queue of packets
type PacketQueue struct {
	packets []*Packet
	maxSize int
}

// NewPacketQueue creates a new packet queue
func NewPacketQueue(maxSize int) *PacketQueue {
	return &PacketQueue{
		packets: make([]*Packet, 0, maxSize),
		maxSize: maxSize,
	}
}

// Enqueue adds a packet to the queue
func (pq *PacketQueue) Enqueue(packet *Packet) error {
	if len(pq.packets) >= pq.maxSize {
		return errors.New("queue is full")
	}
	pq.packets = append(pq.packets, packet)
	return nil
}

// Dequeue removes and returns the first packet
func (pq *PacketQueue) Dequeue() (*Packet, error) {
	if len(pq.packets) == 0 {
		return nil, errors.New("queue is empty")
	}
	packet := pq.packets[0]
	pq.packets = pq.packets[1:]
	return packet, nil
}

// Size returns the current queue size
func (pq *PacketQueue) Size() int {
	return len(pq.packets)
}

// IsEmpty checks if the queue is empty
func (pq *PacketQueue) IsEmpty() bool {
	return len(pq.packets) == 0
}

// Clear removes all packets from the queue
func (pq *PacketQueue) Clear() {
	pq.packets = make([]*Packet, 0, pq.maxSize)
}
