package crypto

import (
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
)

// GenerateSymmetricKey generates a 32-byte key for ChaCha20-Poly1305
func GenerateSymmetricKey() ([]byte, error) {
	key := make([]byte, chacha20poly1305.KeySize)
	_, err := rand.Read(key)
	return key, err
}

// OnionPacket represents an encrypted layer
type OnionPacket struct {
	Payload []byte
}

// CreateOnionPacket encrypts a payload with a key
func CreateOnionPacket(plain, key []byte) (*OnionPacket, error) {
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := aead.Seal(nonce, nonce, plain, nil)
	return &OnionPacket{Payload: ciphertext}, nil
}

// PeelOnion decrypts a packet with a key
func PeelOnion(pkt *OnionPacket, key []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}
	if len(pkt.Payload) < aead.NonceSize() {
		return nil, errors.New("invalid payload")
	}
	nonce := pkt.Payload[:aead.NonceSize()]
	ciphertext := pkt.Payload[aead.NonceSize():]
	return aead.Open(nil, nonce, ciphertext, nil)
}

// Serialize packet
func (p *OnionPacket) Serialize() []byte {
	return p.Payload
}

// Deserialize packet
func Deserialize(data []byte) (*OnionPacket, error) {
	return &OnionPacket{Payload: data}, nil
}
