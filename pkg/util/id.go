package util

import "strconv"

// IDStr converts int64 ID to string (solves JS precision issue)
func IDStr(id int64) string {
	return strconv.FormatInt(id, 10)
}

// IDsStr batch converts int64 IDs to strings
func IDsStr(ids []int64) []string {
	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = strconv.FormatInt(id, 10)
	}
	return result
}
