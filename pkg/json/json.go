// Package json provides a unified JSON encoding/decoding interface.
// It wraps github.com/bytedance/sonic for better performance.
package json

import (
	"io"

	"github.com/bytedance/sonic"
)

var (
	// Marshal returns the JSON encoding of v.
	Marshal = sonic.Marshal

	// MarshalIndent is like Marshal but with indentation.
	MarshalIndent = sonic.MarshalIndent

	// Unmarshal parses the JSON-encoded data and stores the result in v.
	Unmarshal = sonic.Unmarshal

	// MarshalString returns the JSON encoding of v as a string.
	MarshalString = sonic.MarshalString

	// UnmarshalString parses the JSON-encoded string and stores the result in v.
	UnmarshalString = sonic.UnmarshalString
)

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) sonic.Decoder {
	return sonic.ConfigDefault.NewDecoder(r)
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) sonic.Encoder {
	return sonic.ConfigDefault.NewEncoder(w)
}

// Valid reports whether data is a valid JSON encoding.
func Valid(data []byte) bool {
	return sonic.Valid(data)
}
