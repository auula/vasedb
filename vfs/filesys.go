package vfs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/auula/vasedb/clog"
	"github.com/auula/vasedb/conf"
	"github.com/auula/vasedb/utils"
)

// InitFS build vasedb file system
func InitFS(path string) error {

	// 拼接文件路径
	for _, dir := range conf.Dirs {
		// 检查目录是否存在
		if utils.IsExist(filepath.Join(path, dir)) {
			clog.Info(fmt.Sprintf("Initial %s checked successful", dir))
		} else {
			// 不存在创建对应的目录
			err := os.MkdirAll(filepath.Join(path, dir), conf.Permissions)
			if err != nil {
				return err
			}
		}
	}

	clog.Info("Initial storage successful")
	return nil
}
