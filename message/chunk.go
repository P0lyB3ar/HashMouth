package message

import (
	"encoding/json"
	"errors"
)

// Chunk represents a piece of a larger message
type Chunk struct {
	MessageID string `json:"message_id"` // Unique ID for the complete message
	Seq       int    `json:"seq"`        // Sequence number of this chunk
	Total     int    `json:"total"`      // Total number of chunks
	Data      []byte `json:"data"`       // Actual chunk data
}

// NewChunk creates a new message chunk
func NewChunk(messageID string, seq, total int, data []byte) *Chunk {
	return &Chunk{
		MessageID: messageID,
		Seq:       seq,
		Total:     total,
		Data:      data,
	}
}

// Serialize converts chunk to JSON bytes
func (c *Chunk) Serialize() ([]byte, error) {
	return json.Marshal(c)
}

// DeserializeChunk converts JSON bytes back to Chunk
func DeserializeChunk(data []byte) (*Chunk, error) {
	var chunk Chunk
	if err := json.Unmarshal(data, &chunk); err != nil {
		return nil, err
	}
	return &chunk, nil
}

// Validate checks if the chunk is valid
func (c *Chunk) Validate() error {
	if c.MessageID == "" {
		return errors.New("message ID cannot be empty")
	}
	if c.Seq < 0 || c.Seq >= c.Total {
		return errors.New("invalid sequence number")
	}
	if c.Total <= 0 {
		return errors.New("total chunks must be positive")
	}
	if len(c.Data) == 0 {
		return errors.New("chunk data cannot be empty")
	}
	return nil
}

// ChunkAssembler helps reassemble chunks into complete messages
type ChunkAssembler struct {
	chunks map[string]map[int]*Chunk // messageID -> seq -> chunk
}

// NewChunkAssembler creates a new chunk assembler
func NewChunkAssembler() *ChunkAssembler {
	return &ChunkAssembler{
		chunks: make(map[string]map[int]*Chunk),
	}
}

// AddChunk adds a chunk to the assembler
func (ca *ChunkAssembler) AddChunk(chunk *Chunk) error {
	if err := chunk.Validate(); err != nil {
		return err
	}

	if _, exists := ca.chunks[chunk.MessageID]; !exists {
		ca.chunks[chunk.MessageID] = make(map[int]*Chunk)
	}

	ca.chunks[chunk.MessageID][chunk.Seq] = chunk
	return nil
}

// IsComplete checks if all chunks for a message have been received
func (ca *ChunkAssembler) IsComplete(messageID string) bool {
	chunks, exists := ca.chunks[messageID]
	if !exists || len(chunks) == 0 {
		return false
	}

	// Get total from any chunk
	var total int
	for _, chunk := range chunks {
		total = chunk.Total
		break
	}

	// Check if we have all chunks
	if len(chunks) != total {
		return false
	}

	// Verify all sequence numbers are present
	for i := 0; i < total; i++ {
		if _, exists := chunks[i]; !exists {
			return false
		}
	}

	return true
}

// Assemble combines all chunks into the complete message
func (ca *ChunkAssembler) Assemble(messageID string) ([]byte, error) {
	if !ca.IsComplete(messageID) {
		return nil, errors.New("message is not complete")
	}

	chunks := ca.chunks[messageID]
	total := chunks[0].Total

	// Combine chunks in order
	var result []byte
	for i := 0; i < total; i++ {
		result = append(result, chunks[i].Data...)
	}

	// Clean up
	delete(ca.chunks, messageID)

	return result, nil
}

// SplitMessage splits a large message into chunks
func SplitMessage(messageID string, data []byte, chunkSize int) ([]*Chunk, error) {
	if chunkSize <= 0 {
		return nil, errors.New("chunk size must be positive")
	}
	if len(data) == 0 {
		return nil, errors.New("data cannot be empty")
	}

	total := (len(data) + chunkSize - 1) / chunkSize
	chunks := make([]*Chunk, 0, total)

	for i := 0; i < total; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(data) {
			end = len(data)
		}

		chunk := NewChunk(messageID, i, total, data[start:end])
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}
