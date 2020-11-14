package util

import "strings"

// InIgnoreCase 判断元素是否在字符串数组中
func InIgnoreCase(val string, items []string) bool {
	for _, item := range items {
		if strings.EqualFold(val, item) {
			return true
		}
	}

	return false
}
