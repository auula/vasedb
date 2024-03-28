package vfs

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/auula/vasedb/utils"
)

var (
	// Folders 标准目录结构
	folders = []string{"etc", "temp", "data", "index"}
)

// SetupFS build vasedb file system
func SetupFS(path string, perm fs.FileMode) error {

	// 拼接文件路径
	for _, dir := range folders {
		// 检查目录是否存在
		if !utils.IsExist(filepath.Join(path, dir)) {
			// 不存在创建对应的目录
			err := os.MkdirAll(filepath.Join(path, dir), perm)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
