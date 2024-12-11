package vfs

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/auula/vasedb/conf"
	"github.com/auula/vasedb/utils"
)

var (
	// Folders 标准目录结构
	folders = []string{"etc", "temp", "data", "index"}
)

// 这个 SetupFS 函数只需要检查文件系统格式合规不

// SetupFS build vasedb file system
func SetupFS(path string) (*Storage, error) {

	// 拼接文件路径
	for _, dir := range folders {
		// 检查目录是否存在
		if !utils.IsExist(filepath.Join(path, dir)) {
			// 不存在创建对应的目录
			err := os.MkdirAll(filepath.Join(path, dir), conf.FsPerm)
			if err != nil {
				return nil, err
			}
		}
	}

	// 文件检查格式和目录检查成功之后恢复文件
	// 1. 对 data 目录下的文件进行排序
	// 2. 拿到最大那个文件，并且检查文件大小
	// 3. 文件小于阀值，就打开返回，大于创建一个新的文件

	var files []*os.File
	// 遍历目录获取文件
	err := filepath.Walk(filepath.Join(path, "data"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// 检查文件名是否匹配 000x.vsdb
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".vsdb") {
			// 限制文件名格式，长度为 8
			if len(info.Name()) == 8 && strings.HasPrefix(info.Name(), "000") {
				// 以读取模式打开文件
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

	// 获取最大文件，构造 Storage 的实现
	checkAndCreateNewFile(files)

	return nil, nil
}

type DataFile struct {
	path string
	file *os.File
}

// NewDataFile 创建一个新的文件实例
func NewDataFile(path string) (*DataFile, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, conf.FsPerm)
	if err != nil {
		return nil, err
	}

	return &DataFile{
		path: path,
		file: file,
	}, nil
}

// Open 打开文件
func (fs *DataFile) Open() error {
	var err error
	fs.file, err = os.OpenFile(fs.path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	return nil
}

// Read 从文件中读取内容
func (fs *DataFile) Read(p []byte) (int, error) {
	if fs.file == nil {
		return 0, os.ErrInvalid
	}
	return fs.file.Read(p)
}

// Write 将数据写入文件
func (fs *DataFile) Write(data []byte) (int, error) {
	if fs.file == nil {
		return 0, os.ErrInvalid
	}
	return fs.file.Write(data)
}

// Close 关闭文件
func (fs *DataFile) Close() error {
	if fs.file != nil {
		err := fs.file.Close()
		if err != nil {
			return err
		}
		fs.file = nil
		return nil
	}
	return nil
}

func checkAndCreateNewFile(files []*os.File) (*os.File, error) {
	// 确保文件列表不为空
	if len(files) == 0 {
		return nil, errors.New("no data files found")
	}

	// 获取最大文件
	lastTimeFile := files[len(files)-1]

	// 获取最大文件的 FileInfo
	fileInfo, err := lastTimeFile.Stat()
	if err != nil {
		return nil, err
	}

	// 如果文件大小超过预设的文件大小，则创建新文件
	if fileInfo.Size() >= conf.Settings.FileSize {
		// 创建新的文件，生成一个递增数据文件名称
		newFilePath := filepath.Join(conf.Settings.Path, "")
		newFile, err := os.Create(newFilePath)
		if err != nil {
			return nil, err
		}
		return newFile, nil
	}

	// 如果没有超过大小，返回上一次最后一次使用的文件
	return lastTimeFile, nil
}
