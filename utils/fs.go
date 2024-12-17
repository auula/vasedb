package utils

import (
	"fmt"
	"os"
)

// IsExist checked directory is exist
func IsExist(dirPath string) bool {
	// 使用 os.Stat 检查目录是否存在
	_, err := os.Stat(dirPath)

	// 如果 err 不为 nil 并且是目录不存在错误返回 false
	// 如果 err 为 nil 或者是其他类型的错误，权限问题则返回 true
	return !(err != nil && os.IsNotExist(err))
}

// IsDir check if the path is a directory
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// CloseFile 封装了文件的 Sync 和 Close 操作，减少重复代码
func CloseFile(file *os.File) error {
	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}
	return nil
}
