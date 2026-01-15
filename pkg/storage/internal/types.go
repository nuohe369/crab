package internal

// FileInfo file information
type FileInfo struct {
	Key          string // File path/key
	Size         int64  // File size (bytes)
	ContentType  string // MIME type
	LastModified int64  // Last modified time (Unix timestamp)
}
