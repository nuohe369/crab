package util

import (
	"strconv"
	"sync"
)

// String conversion cache for frequently used int64 to string conversions
// This is particularly useful for ID conversions in API responses
// 字符串转换缓存，用于频繁使用的 int64 到 string 的转换
// 这对于 API 响应中的 ID 转换特别有用

const (
	// Cache range for small integers (0-10000)
	// 小整数的缓存范围 (0-10000)
	cacheMin = 0
	cacheMax = 10000
)

var (
	// Pre-computed string cache for common integers
	// 常用整数的预计算字符串缓存
	int64StrCache     []string
	int64StrCacheOnce sync.Once
)

// initInt64StrCache initializes the string cache
// initInt64StrCache 初始化字符串缓存
func initInt64StrCache() {
	int64StrCache = make([]string, cacheMax-cacheMin+1)
	for i := cacheMin; i <= cacheMax; i++ {
		int64StrCache[i-cacheMin] = strconv.FormatInt(int64(i), 10)
	}
}

// Int64ToString converts int64 to string with caching for common values
// This provides significant performance improvement for frequently converted IDs
// Int64ToString 将 int64 转换为字符串，对常用值使用缓存
// 这为频繁转换的 ID 提供了显著的性能提升
func Int64ToString(n int64) string {
	// Initialize cache on first use | 首次使用时初始化缓存
	int64StrCacheOnce.Do(initInt64StrCache)

	// Use cache for common range | 对常用范围使用缓存
	if n >= cacheMin && n <= cacheMax {
		return int64StrCache[n-cacheMin]
	}

	// Fall back to standard conversion for values outside cache range
	// 对缓存范围外的值回退到标准转换
	return strconv.FormatInt(n, 10)
}

// StringToInt64 converts string to int64 (no caching needed for parsing)
// StringToInt64 将字符串转换为 int64（解析不需要缓存）
func StringToInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// MustStringToInt64 converts string to int64, returns 0 on error
// MustStringToInt64 将字符串转换为 int64，错误时返回 0
func MustStringToInt64(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

// Int64ToStringBatch converts multiple int64 values to strings efficiently
// Int64ToStringBatch 高效地将多个 int64 值转换为字符串
func Int64ToStringBatch(nums []int64) []string {
	result := make([]string, len(nums))
	for i, n := range nums {
		result[i] = Int64ToString(n)
	}
	return result
}
