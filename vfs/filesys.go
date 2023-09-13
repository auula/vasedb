package vfs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/auula/vasedb/clog"
	"github.com/auula/vasedb/conf"
)

// InitFS build vasedb file system
func InitFS(path string) error {

	// 拼接文件路径
	for _, dir := range conf.Dirs {
		// 检查目录是否存在
		if dirExist(filepath.Join(path, dir)) {
			clog.Info(fmt.Sprintf("Initial %s checked successful", dir))
		} else {
			// 不存在创建对应的目录
			if err := os.MkdirAll(filepath.Join(path, dir), conf.Permissions); err != nil {
				return err
			}
		}
	}

	clog.Info("Initial storage successful")
	return nil
}

func dirExist(dirPath string) bool {
	// 使用 os.Stat 检查目录是否存在
	_, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		return false
	} else {
		return true
	}
}
