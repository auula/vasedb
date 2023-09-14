package utils

import (
	"os"
	"testing"
)

func TestIsDirExist(t *testing.T) {
	// 测试存在的目录
	existingDir := os.TempDir()
	exists := IsExist(existingDir)
	if !exists {
		t.Errorf("Expected directory %s to exist, but it does not.", existingDir)
	}

	// 测试不存在的目录
	nonExistingDir := "/aaa/bbb/cccc/directory"
	exists = IsExist(nonExistingDir)
	if exists {
		t.Errorf("Expected directory %s to not exist, but it does.", nonExistingDir)
	}

	// 测试无效路径
	invalidPath := "/invalid/path"
	exists = IsExist(invalidPath)
	if exists {
		t.Errorf("Expected directory %s to not exist, but it does.", invalidPath)
	}
}
