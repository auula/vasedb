package vfs

import (
	"os"
	"path/filepath"

	"github.com/auula/vasedb/clog"
	"github.com/auula/vasedb/conf"
	"github.com/auula/vasedb/utils"
)

// SetupFS build vasedb file system
func SetupFS(path string, folders ...string) error {

	// 拼接文件路径
	for _, dir := range folders {
		// 检查目录是否存在
		if utils.IsExist(filepath.Join(path, dir)) {
			clog.Infof("Initial %s checked successful", dir)
		} else {
			// 不存在创建对应的目录
			err := os.MkdirAll(filepath.Join(path, dir), conf.Permissions)
			if err != nil {
				return err
			}
		}
	}

	// 不要写在这里如果这个包被单独拿出去使用，不能配置 clog 包使用
	// clog.Info("Initial storage successful")
	return nil
}
