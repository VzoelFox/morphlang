package agent

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const MemoryFile = ".vzoel.jules"

type Interaction struct {
	Index       int64     `json:"index"`
	Timestamp   time.Time `json:"timestamp"`
	User        string    `json:"user"`
	Assistant   string    `json:"assistant"`
	ContextHash string    `json:"context_hash"` // Optional: Hash of critical files (ROADMAP.md, etc)
	PrevHash    string    `json:"prev_hash"`
	Hash        string    `json:"hash"`
}

// CalculateHash generates SHA256 of the interaction content + prevHash.
func (i *Interaction) CalculateHash() string {
	// Simple concatenation logic.
	// For robustness, consider separating fields with a delimiter or using structured hash.
	record := fmt.Sprintf("%d%s%s%s%s%s", i.Index, i.Timestamp.UTC().String(), i.User, i.Assistant, i.ContextHash, i.PrevHash)
	h := sha256.New()
	h.Write([]byte(record))
	return hex.EncodeToString(h.Sum(nil))
}

// Memory represents the chain of interactions.
type Memory struct {
	Interactions []Interaction
	Filepath     string
}

// LoadMemory reads the memory file and verifies integrity.
func LoadMemory(path string) (*Memory, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Memory{Filepath: path}, nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var interactions []Interaction
	if len(content) > 0 {
		if err := json.Unmarshal(content, &interactions); err != nil {
			return nil, err
		}
	}

	mem := &Memory{Interactions: interactions, Filepath: path}
	if err := mem.Verify(); err != nil {
		return nil, err
	}

	return mem, nil
}

// Verify checks the hash chain integrity.
func (m *Memory) Verify() error {
	var prevHash string
	for i, interact := range m.Interactions {
		if interact.PrevHash != prevHash {
			return fmt.Errorf("integrity failure at index %d: prev_hash mismatch (expected %s, got %s)", i, prevHash, interact.PrevHash)
		}
		expected := interact.CalculateHash()
		if interact.Hash != expected {
			return fmt.Errorf("integrity failure at index %d: hash mismatch", i)
		}
		prevHash = interact.Hash
	}
	return nil
}

// Record appends a new interaction.
func (m *Memory) Record(user, assistant string) error {
	var prevHash string
	var index int64
	if len(m.Interactions) > 0 {
		last := m.Interactions[len(m.Interactions)-1]
		prevHash = last.Hash
		index = last.Index + 1
	}

	interact := Interaction{
		Index:     index,
		Timestamp: time.Now().UTC(), // Use UTC for consistency
		User:      user,
		Assistant: assistant,
		PrevHash:  prevHash,
	}
	interact.Hash = interact.CalculateHash()

	m.Interactions = append(m.Interactions, interact)
	return m.Save()
}

// Save writes the chain to disk.
func (m *Memory) Save() error {
	data, err := json.MarshalIndent(m.Interactions, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.Filepath, data, 0644)
}
