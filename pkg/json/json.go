// Package json provides a unified JSON encoding/decoding interface.
// It wraps github.com/bytedance/sonic for better performance.
// Package json 提供统一的 JSON 编码/解码接口
// 它封装了 github.com/bytedance/sonic 以获得更好的性能
package json

import (
	"io"

	"github.com/bytedance/sonic"
)

var (
	// Marshal returns the JSON encoding of v.
	// Marshal 返回 v 的 JSON 编码
	Marshal = sonic.Marshal

	// MarshalIndent is like Marshal but with indentation.
	// MarshalIndent 类似 Marshal 但带缩进
	MarshalIndent = sonic.MarshalIndent

	// Unmarshal parses the JSON-encoded data and stores the result in v.
	// Unmarshal 解析 JSON 编码的数据并将结果存储在 v 中
	Unmarshal = sonic.Unmarshal

	// MarshalString returns the JSON encoding of v as a string.
	// MarshalString 返回 v 的 JSON 编码字符串
	MarshalString = sonic.MarshalString

	// UnmarshalString parses the JSON-encoded string and stores the result in v.
	// UnmarshalString 解析 JSON 编码的字符串并将结果存储在 v 中
	UnmarshalString = sonic.UnmarshalString
)

// NewDecoder returns a new decoder that reads from r.
// NewDecoder 返回从 r 读取的新解码器
func NewDecoder(r io.Reader) sonic.Decoder {
	return sonic.ConfigDefault.NewDecoder(r)
}

// NewEncoder returns a new encoder that writes to w.
// NewEncoder 返回写入 w 的新编码器
func NewEncoder(w io.Writer) sonic.Encoder {
	return sonic.ConfigDefault.NewEncoder(w)
}

// Valid reports whether data is a valid JSON encoding.
// Valid 报告 data 是否为有效的 JSON 编码
func Valid(data []byte) bool {
	return sonic.Valid(data)
}
