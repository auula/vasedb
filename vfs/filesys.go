package vfs

import (
	"os"
	"path/filepath"

	"github.com/auula/vasedb/clog"
	"github.com/auula/vasedb/conf"
)

// InitFS build classdb file system
func InitFS(path string) error {
	// 拼接文件路径
	for _, dir := range conf.Dirs {
		if err := os.MkdirAll(filepath.Join(path, dir), conf.Settings.Permissions); err != nil {
			return err
		}
	}
	clog.Info("Initial storage successful")
	return nil
}
