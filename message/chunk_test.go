package message

import (
	"bytes"
	"testing"
)

func TestNewChunk(t *testing.T) {
	chunk := NewChunk("msg1", 0, 3, []byte("test data"))
	
	if chunk.MessageID != "msg1" {
		t.Errorf("Expected MessageID 'msg1', got '%s'", chunk.MessageID)
	}
	if chunk.Seq != 0 {
		t.Errorf("Expected Seq 0, got %d", chunk.Seq)
	}
	if chunk.Total != 3 {
		t.Errorf("Expected Total 3, got %d", chunk.Total)
	}
}

func TestChunkSerializeDeserialize(t *testing.T) {
	chunk := NewChunk("msg1", 0, 1, []byte("test"))
	
	serialized, err := chunk.Serialize()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	deserialized, err := DeserializeChunk(serialized)
	if err != nil {
		t.Fatalf("Failed to deserialize: %v", err)
	}

	if deserialized.MessageID != chunk.MessageID {
		t.Error("MessageID mismatch")
	}
	if !bytes.Equal(deserialized.Data, chunk.Data) {
		t.Error("Data mismatch")
	}
}

func TestChunkValidate(t *testing.T) {
	tests := []struct {
		name    string
		chunk   *Chunk
		wantErr bool
	}{
		{
			name:    "valid chunk",
			chunk:   NewChunk("msg1", 0, 1, []byte("data")),
			wantErr: false,
		},
		{
			name:    "empty message ID",
			chunk:   NewChunk("", 0, 1, []byte("data")),
			wantErr: true,
		},
		{
			name:    "invalid sequence",
			chunk:   NewChunk("msg1", -1, 1, []byte("data")),
			wantErr: true,
		},
		{
			name:    "empty data",
			chunk:   NewChunk("msg1", 0, 1, []byte{}),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.chunk.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSplitMessage(t *testing.T) {
	data := []byte("This is a test message that will be split into chunks")
	chunkSize := 10

	chunks, err := SplitMessage("msg1", data, chunkSize)
	if err != nil {
		t.Fatalf("Failed to split message: %v", err)
	}

	expectedChunks := (len(data) + chunkSize - 1) / chunkSize
	if len(chunks) != expectedChunks {
		t.Errorf("Expected %d chunks, got %d", expectedChunks, len(chunks))
	}

	// Verify all chunks are valid
	for i, chunk := range chunks {
		if chunk.Seq != i {
			t.Errorf("Chunk %d has wrong sequence number: %d", i, chunk.Seq)
		}
		if chunk.Total != expectedChunks {
			t.Errorf("Chunk %d has wrong total: %d", i, chunk.Total)
		}
	}
}

func TestChunkAssembler(t *testing.T) {
	data := []byte("Complete message to be chunked and reassembled")
	chunks, err := SplitMessage("msg1", data, 10)
	if err != nil {
		t.Fatalf("Failed to split message: %v", err)
	}

	assembler := NewChunkAssembler()

	// Add all chunks
	for _, chunk := range chunks {
		if err := assembler.AddChunk(chunk); err != nil {
			t.Fatalf("Failed to add chunk: %v", err)
		}
	}

	// Check if complete
	if !assembler.IsComplete("msg1") {
		t.Error("Message should be complete")
	}

	// Assemble
	assembled, err := assembler.Assemble("msg1")
	if err != nil {
		t.Fatalf("Failed to assemble: %v", err)
	}

	if !bytes.Equal(data, assembled) {
		t.Errorf("Assembled data doesn't match original.\nExpected: %s\nGot: %s", data, assembled)
	}
}

func TestChunkAssemblerPartial(t *testing.T) {
	chunks, _ := SplitMessage("msg1", []byte("test message"), 5)
	
	assembler := NewChunkAssembler()
	
	// Add only first chunk
	assembler.AddChunk(chunks[0])
	
	if assembler.IsComplete("msg1") {
		t.Error("Message should not be complete with only one chunk")
	}
	
	_, err := assembler.Assemble("msg1")
	if err == nil {
		t.Error("Should not be able to assemble incomplete message")
	}
}
