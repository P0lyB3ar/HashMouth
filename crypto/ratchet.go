package crypto

import (
    "crypto/rand"
    "errors"
    "golang.org/x/crypto/curve25519"
)

// RatchetSession holds the state for a single session with a peer
type RatchetSession struct {
    DHPrivate []byte // our ephemeral private key
    DHPublic  []byte // our ephemeral public key
    PeerPub   []byte // peer's ephemeral public key
    RootKey   []byte // shared secret root
    ChainKey  []byte // evolving chain key for message encryption
}

// NewRatchetSession creates a new session with a peer
func NewRatchetSession(peerPub []byte) (*RatchetSession, error) {
    priv := make([]byte, 32)
    _, err := rand.Read(priv)
    if err != nil {
        return nil, err
    }
    pub, err := curve25519.X25519(priv, curve25519.Basepoint)
    if err != nil {
        return nil, err
    }

    // Derive initial shared secret
    if len(peerPub) != 32 {
        return nil, errors.New("invalid peer public key")
    }
    shared, err := curve25519.X25519(priv, peerPub)
    if err != nil {
        return nil, err
    }

    session := &RatchetSession{
        DHPrivate: priv,
        DHPublic:  pub,
        PeerPub:   peerPub,
        RootKey:   shared,
        ChainKey:  shared, // simple start, will evolve per message
    }
    return session, nil
}

// RatchetStep derives a new chain key (simplified)
// In a real implementation, use HMAC or KDF (like HKDF)
func (r *RatchetSession) RatchetStep() {
    newKey := make([]byte, len(r.ChainKey))
    copy(newKey, r.ChainKey)
    for i := range newKey {
        newKey[i] ^= 0x55 // very simple placeholder for demo
    }
    r.ChainKey = newKey
}

// GetNextKey returns the current chain key to encrypt the next message
func (r *RatchetSession) GetNextKey() []byte {
    r.RatchetStep()
    return r.ChainKey
}
