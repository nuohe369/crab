package snowflake

import (
	"sync"
	"time"
)

const (
	epoch         = int64(1704067200000) // 2024-01-01 00:00:00 UTC
	machineBits   = 10
	sequenceBits  = 12
	machineMax    = -1 ^ (-1 << machineBits)
	sequenceMax   = -1 ^ (-1 << sequenceBits)
	machineShift  = sequenceBits
	timestampShift = machineBits + sequenceBits
)

// Node 雪花 ID 生成器
type Node struct {
	mu        sync.Mutex
	machineID int64
	sequence  int64
	lastTime  int64
}

var defaultNode *Node

// Init 初始化默认nodes
func Init(machineID int64) error {
	if machineID < 0 || machineID > machineMax {
		machineID = machineID & machineMax
	}
	defaultNode = &Node{machineID: machineID}
	return nil
}

// Generate 生成 ID(使用默认nodes)
func Generate() int64 {
	if defaultNode == nil {
		Init(1)
	}
	return defaultNode.Generate()
}

// Generate 生成 ID
func (n *Node) Generate() int64 {
	n.mu.Lock()
	defer n.mu.Unlock()

	now := time.Now().UnixMilli()

	if now == n.lastTime {
		n.sequence = (n.sequence + 1) & sequenceMax
		if n.sequence == 0 {
			for now <= n.lastTime {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		n.sequence = 0
	}

	n.lastTime = now

	return ((now - epoch) << timestampShift) |
		(n.machineID << machineShift) |
		n.sequence
}

// Parse 解析 ID
func Parse(id int64) (timestamp time.Time, machineID, sequence int64) {
	machineID = (id >> machineShift) & machineMax
	sequence = id & sequenceMax
	ts := (id >> timestampShift) + epoch
	timestamp = time.UnixMilli(ts)
	return
}

// NewNode 创建新nodes
func NewNode(machineID int64) *Node {
	if machineID < 0 || machineID > machineMax {
		machineID = machineID & machineMax
	}
	return &Node{machineID: machineID}
}
