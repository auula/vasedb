package utils

import "os"

// IsExist checked directory is exist
func IsExist(dirPath string) bool {
	// 使用 os.Stat 检查目录是否存在
	_, err := os.Stat(dirPath)
	if err != nil && os.IsNotExist(err) {
		// 如果 err 不为 nil 并且是目录不存在错误返回 false
		return false
	}
	// 如果 err 为 nil 或者是其他类型的错误，权限问题则返回 true
	return true
}
