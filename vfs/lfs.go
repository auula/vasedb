package vfs

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/auula/vasedb/conf"
	"github.com/auula/vasedb/utils"
)

const (
	etc   = "etc"
	temp  = "temp"
	data  = "data"
	index = "index"
)

var (
	// dirs 标准目录结构
	dirs = []string{etc, temp, data, index}

	// Data file name extension
	dataFileSuffix = ".vsdb"

	// index file name extension
	indexFileSuffix = ".vsix"
)

// 这个 SetupFS 函数只需要检查文件系统格式合规不

// SetupFS build vasedb file system
func SetupFS(path string) (*LogStructuredFS, error) {

	// 拼接文件路径
	for _, dir := range dirs {
		// 检查目录是否存在
		if !utils.IsExist(filepath.Join(path, dir)) {
			// 不存在创建对应的目录
			err := os.MkdirAll(filepath.Join(path, dir), conf.FsPerm)
			if err != nil {
				return nil, err
			}
		}
	}

	var files []*os.File
	// 遍历目录获取文件
	err := filepath.Walk(filepath.Join(path, data), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// 检查文件名是否匹配 000x.vsdb
		if !info.IsDir() && strings.HasSuffix(info.Name(), dataFileSuffix) {
			// 限制文件名格式，长度为 8，为什么不用时间戳
			if len(info.Name()) == 8 && strings.HasPrefix(info.Name(), "000") {
				// 改成检查数据文件格式是否合法
				file, err := os.Open(path)
				if err != nil {
					return err
				}
				files = append(files, file)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// 对文件名排序
	sort.Slice(files, func(i, j int) bool {
		return strings.Compare(
			filepath.Base(files[i].Name()),
			filepath.Base(files[j].Name())) < 0
	})

	return nil, nil
}

// INode represents a file system node with metadata.
type INode struct {
	ID          uint16    // Unique identifier for the INode
	Offset      uint32    // Offset within the file
	CreatedTime time.Time // Creation time of the INode
	EexpireTime time.Time // Expiration time of the INode
}

// LogStructuredFS represents the virtual file storage system.
type LogStructuredFS struct {
	Indexs      map[uint64]*INode   // Index mapping for INode references
	BlockGroup  map[uint16]*os.File // Archived files keyed by unique file ID
	ActiveBlock *os.File            // Currently active file for writing
}
