package snowflake

import (
	"sync"
	"testing"
	"time"
)

func TestGenerate(t *testing.T) {
	Init(1)

	id := Generate()
	if id <= 0 {
		t.Errorf("Generated ID should be positive, got %d", id)
	}
}

func TestGenerateUnique(t *testing.T) {
	Init(1)

	ids := make(map[int64]bool)
	count := 10000

	for i := 0; i < count; i++ {
		id := Generate()
		if ids[id] {
			t.Fatalf("Duplicate ID generated: %d", id)
		}
		ids[id] = true
	}
}

func TestGenerateConcurrent(t *testing.T) {
	Init(1)

	var mu sync.Mutex
	ids := make(map[int64]bool)
	var wg sync.WaitGroup

	workers := 10
	idsPerWorker := 1000

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			localIDs := make([]int64, idsPerWorker)
			for j := 0; j < idsPerWorker; j++ {
				localIDs[j] = Generate()
			}
			mu.Lock()
			for _, id := range localIDs {
				if ids[id] {
					t.Errorf("Duplicate ID: %d", id)
				}
				ids[id] = true
			}
			mu.Unlock()
		}()
	}

	wg.Wait()

	if len(ids) != workers*idsPerWorker {
		t.Errorf("Expected %d unique IDs, got %d", workers*idsPerWorker, len(ids))
	}
}

func TestParse(t *testing.T) {
	Init(5)

	before := time.Now()
	id := Generate()
	after := time.Now()

	timestamp, machineID, sequence := Parse(id)

	if machineID != 5 {
		t.Errorf("MachineID mismatch: got %d, want 5", machineID)
	}

	if sequence < 0 {
		t.Errorf("Sequence should be non-negative, got %d", sequence)
	}

	if timestamp.Before(before.Add(-time.Second)) || timestamp.After(after.Add(time.Second)) {
		t.Errorf("Timestamp out of range: %v", timestamp)
	}
}

func TestNewNode(t *testing.T) {
	node1 := NewNode(1)
	node2 := NewNode(2)

	id1 := node1.Generate()
	id2 := node2.Generate()

	_, machineID1, _ := Parse(id1)
	_, machineID2, _ := Parse(id2)

	if machineID1 != 1 {
		t.Errorf("Node1 machineID mismatch: got %d, want 1", machineID1)
	}

	if machineID2 != 2 {
		t.Errorf("Node2 machineID mismatch: got %d, want 2", machineID2)
	}
}

func TestMachineIDOverflow(t *testing.T) {
	// machineMax is 1023 (10 bits)
	node := NewNode(9999)
	id := node.Generate()

	_, machineID, _ := Parse(id)

	// Should be masked to valid range
	if machineID < 0 || machineID > machineMax {
		t.Errorf("MachineID out of range: %d", machineID)
	}
}

func BenchmarkGenerate(b *testing.B) {
	Init(1)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Generate()
	}
}

func BenchmarkGenerateParallel(b *testing.B) {
	Init(1)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Generate()
		}
	})
}
