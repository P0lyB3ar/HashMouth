package crypto

import (
	"crypto/rand"
	"crypto/ed25519"
)

// GenerateIdentityKeyPair generates a new Ed25519 keypair for identity
func GenerateIdentityKeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return pub, priv, nil
}

// GenerateSymmetricKey returns a random 32-byte key for ChaCha20-Poly1305
func GenerateSymmetricKey() ([]byte, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}
