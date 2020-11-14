package util

import (
	"os"
	"strings"
)

// InIgnoreCase 判断元素是否在字符串数组中
func InIgnoreCase(val string, items []string) bool {
	for _, item := range items {
		if strings.EqualFold(val, item) {
			return true
		}
	}

	return false
}

// FileExist 判断文件是否存在
func FileExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}