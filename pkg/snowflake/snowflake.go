// Package snowflake provides distributed unique ID generation using the Snowflake algorithm
// Package snowflake 提供使用雪花算法的分布式唯一 ID 生成
package snowflake

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strconv"
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

// SnowflakeID is a custom type for Snowflake IDs with automatic conversion support
// SnowflakeID 是雪花 ID 的自定义类型，支持自动转换
type SnowflakeID int64

// FromDB converts database value to SnowflakeID (called by XORM when reading from database)
// FromDB 将数据库值转换为 SnowflakeID（XORM 从数据库读取时调用）
func (s *SnowflakeID) FromDB(b []byte) error {
	if b == nil || len(b) == 0 {
		*s = 0
		return nil
	}
	// Parse the database value (stored as bigint) to int64
	id, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return errors.New("failed to parse snowflake ID: " + err.Error())
	}
	*s = SnowflakeID(id)
	return nil
}

// ToDB converts SnowflakeID to database value (called by XORM when writing to database)
// ToDB 将 SnowflakeID 转换为数据库值（XORM 写入数据库时调用）
func (s SnowflakeID) ToDB() (driver.Value, error) {
	return int64(s), nil
}

// String returns the string representation of SnowflakeID
// String 返回 SnowflakeID 的字符串表示
func (s SnowflakeID) String() string {
	return strconv.FormatInt(int64(s), 10)
}

// Int64 returns the int64 value of SnowflakeID
// Int64 返回 SnowflakeID 的 int64 值
func (s SnowflakeID) Int64() int64 {
	return int64(s)
}

// MarshalJSON implements json.Marshaler interface to serialize as string
// MarshalJSON 实现 json.Marshaler 接口，序列化为字符串
func (s SnowflakeID) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// UnmarshalJSON implements json.Unmarshaler interface to deserialize from string or number
// UnmarshalJSON 实现 json.Unmarshaler 接口，从字符串或数字反序列化
func (s *SnowflakeID) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		id, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return errors.New("invalid snowflake ID string: " + err.Error())
		}
		*s = SnowflakeID(id)
		return nil
	}

	// Try to unmarshal as number
	var num int64
	if err := json.Unmarshal(data, &num); err == nil {
		*s = SnowflakeID(num)
		return nil
	}

	return errors.New("snowflake ID must be a string or number")
}

// IsZero checks if the SnowflakeID is zero
// IsZero 检查 SnowflakeID 是否为零
func (s SnowflakeID) IsZero() bool {
	return s == 0
}

// Valid checks if the SnowflakeID is valid (non-zero)
// Valid 检查 SnowflakeID 是否有效（非零）
func (s SnowflakeID) Valid() bool {
	return s != 0
}
