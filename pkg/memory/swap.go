package memory

import (
	"fmt"
	"os"
	"sync"
)

const SWAP_FILE = ".morph_cache.z"

type SwapSystem struct {
	mu       sync.Mutex
	file     *os.File
	isActive bool
}

var Swap SwapSystem

func InitSwap() error {
	Swap.mu.Lock()
	defer Swap.mu.Unlock()

	// 0600 = Readable/Writable by owner only (not user accessible generally)
	f, err := os.OpenFile(SWAP_FILE, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return err
	}

	// Write Magic Header to make it "readable system"
	header := []byte("MORPH_SWAP_V1")
	if stat, _ := f.Stat(); stat.Size() == 0 {
		_, err = f.Write(header)
		if err != nil {
			return err
		}
	}

	Swap.file = f
	Swap.isActive = true
	return nil
}

// Spill writes data to swap file and returns the offset
func (s *SwapSystem) Spill(data []byte) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isActive {
		return -1, fmt.Errorf("swap not initialized")
	}

	info, err := s.file.Stat()
	if err != nil {
		return -1, err
	}
	offset := info.Size()

	_, err = s.file.Write(data)
	if err != nil {
		return -1, err
	}

	// Sync to ensure disk write
	// s.file.Sync() // Optional, slow but safe

	return offset, nil
}

// Restore reads data from swap file at offset
func (s *SwapSystem) Restore(offset int64, dest []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isActive {
		return fmt.Errorf("swap not initialized")
	}

	_, err := s.file.ReadAt(dest, offset)
	return err
}

// FreeCache clears the swap file (Manual Free)
func (s *SwapSystem) FreeCache() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.file != nil {
		s.file.Close()
	}

	err := os.Remove(SWAP_FILE)
	s.isActive = false
	return err
}
