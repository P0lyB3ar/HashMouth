package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
)

// GenerateIdentityKeyPair generates a new Ed25519 keypair for identity
func GenerateIdentityKeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return pub, priv, nil
}

// Note: GenerateSymmetricKey is defined in crypto.go to avoid duplication
