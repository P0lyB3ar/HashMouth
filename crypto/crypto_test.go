package crypto

import (
	"bytes"
	"testing"
)

func TestGenerateSymmetricKey(t *testing.T) {
	key, err := GenerateSymmetricKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	if len(key) != 32 {
		t.Errorf("Expected key length 32, got %d", len(key))
	}
}

func TestCreateAndPeelOnion(t *testing.T) {
	key, err := GenerateSymmetricKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	plaintext := []byte("Hello, World!")
	
	// Create onion packet
	packet, err := CreateOnionPacket(plaintext, key)
	if err != nil {
		t.Fatalf("Failed to create onion packet: %v", err)
	}

	// Peel onion packet
	decrypted, err := PeelOnion(packet, key)
	if err != nil {
		t.Fatalf("Failed to peel onion: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("Decrypted text doesn't match. Expected %s, got %s", plaintext, decrypted)
	}
}

func TestMultiLayerOnion(t *testing.T) {
	// Generate keys for 3 layers
	key1, _ := GenerateSymmetricKey()
	key2, _ := GenerateSymmetricKey()
	key3, _ := GenerateSymmetricKey()

	plaintext := []byte("Secret message")

	// Layer 3 (innermost)
	pkt3, err := CreateOnionPacket(plaintext, key3)
	if err != nil {
		t.Fatalf("Failed to create layer 3: %v", err)
	}

	// Layer 2
	pkt2, err := CreateOnionPacket(pkt3.Serialize(), key2)
	if err != nil {
		t.Fatalf("Failed to create layer 2: %v", err)
	}

	// Layer 1 (outermost)
	pkt1, err := CreateOnionPacket(pkt2.Serialize(), key1)
	if err != nil {
		t.Fatalf("Failed to create layer 1: %v", err)
	}

	// Peel layer 1
	data1, err := PeelOnion(pkt1, key1)
	if err != nil {
		t.Fatalf("Failed to peel layer 1: %v", err)
	}

	// Peel layer 2
	pkt2b, _ := Deserialize(data1)
	data2, err := PeelOnion(pkt2b, key2)
	if err != nil {
		t.Fatalf("Failed to peel layer 2: %v", err)
	}

	// Peel layer 3
	pkt3b, _ := Deserialize(data2)
	data3, err := PeelOnion(pkt3b, key3)
	if err != nil {
		t.Fatalf("Failed to peel layer 3: %v", err)
	}

	if !bytes.Equal(plaintext, data3) {
		t.Errorf("Final decrypted text doesn't match. Expected %s, got %s", plaintext, data3)
	}
}

func TestSerializeDeserialize(t *testing.T) {
	key, _ := GenerateSymmetricKey()
	plaintext := []byte("Test data")

	pkt, err := CreateOnionPacket(plaintext, key)
	if err != nil {
		t.Fatalf("Failed to create packet: %v", err)
	}

	serialized := pkt.Serialize()
	deserialized, err := Deserialize(serialized)
	if err != nil {
		t.Fatalf("Failed to deserialize: %v", err)
	}

	if !bytes.Equal(pkt.Payload, deserialized.Payload) {
		t.Error("Serialization/deserialization failed")
	}
}

func TestGenerateIdentityKeyPair(t *testing.T) {
	pub, priv, err := GenerateIdentityKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate identity keypair: %v", err)
	}

	if len(pub) != 32 {
		t.Errorf("Expected public key length 32, got %d", len(pub))
	}
	if len(priv) != 64 {
		t.Errorf("Expected private key length 64, got %d", len(priv))
	}
}
