// Package snowflake provides distributed unique ID generation using the Snowflake algorithm
// Package snowflake 提供使用雪花算法的分布式唯一 ID 生成
package snowflake

import (
	"sync"
	"time"
)

const (
	epoch          = int64(1704067200000)       // 2024-01-01 00:00:00 UTC | 起始时间戳
	machineBits    = 10                         // Machine ID bits | 机器 ID 位数
	sequenceBits   = 12                         // Sequence bits | 序列号位数
	machineMax     = -1 ^ (-1 << machineBits)   // Max machine ID | 最大机器 ID
	sequenceMax    = -1 ^ (-1 << sequenceBits)  // Max sequence | 最大序列号
	machineShift   = sequenceBits               // Machine ID shift | 机器 ID 左移位数
	timestampShift = machineBits + sequenceBits // Timestamp shift | 时间戳左移位数
)

// Node represents a Snowflake ID generator
// Node 表示雪花 ID 生成器
type Node struct {
	mu        sync.Mutex // Mutex for concurrent access | 并发访问互斥锁
	machineID int64      // Machine ID | 机器 ID
	sequence  int64      // Sequence number | 序列号
	lastTime  int64      // Last timestamp | 上次时间戳
}

var defaultNode *Node // Default node instance | 默认节点实例

// Init initializes the default node
// Init 初始化默认节点
func Init(machineID int64) error {
	if machineID < 0 || machineID > machineMax {
		machineID = machineID & machineMax
	}
	defaultNode = &Node{machineID: machineID}
	return nil
}

// Generate generates an ID using the default node
// Generate 使用默认节点生成 ID
func Generate() int64 {
	if defaultNode == nil {
		Init(1)
	}
	return defaultNode.Generate()
}

// Generate generates a unique ID
// Generate 生成唯一 ID
func (n *Node) Generate() int64 {
	n.mu.Lock()
	defer n.mu.Unlock()

	now := time.Now().UnixMilli()

	if now == n.lastTime {
		// Same millisecond, increment sequence | 同一毫秒内，递增序列号
		n.sequence = (n.sequence + 1) & sequenceMax
		if n.sequence == 0 {
			// Sequence overflow, wait for next millisecond | 序列号溢出，等待下一毫秒
			for now <= n.lastTime {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		// New millisecond, reset sequence | 新的毫秒，重置序列号
		n.sequence = 0
	}

	n.lastTime = now

	// Combine timestamp, machine ID, and sequence | 组合时间戳、机器 ID 和序列号
	return ((now - epoch) << timestampShift) |
		(n.machineID << machineShift) |
		n.sequence
}

// Parse parses a Snowflake ID into its components
// Parse 解析雪花 ID 为各个组成部分
func Parse(id int64) (timestamp time.Time, machineID, sequence int64) {
	machineID = (id >> machineShift) & machineMax
	sequence = id & sequenceMax
	ts := (id >> timestampShift) + epoch
	timestamp = time.UnixMilli(ts)
	return
}

// NewNode creates a new Snowflake node
// NewNode 创建新的雪花节点
func NewNode(machineID int64) *Node {
	if machineID < 0 || machineID > machineMax {
		machineID = machineID & machineMax
	}
	return &Node{machineID: machineID}
}
