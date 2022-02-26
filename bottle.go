// Open Source: MIT License
// Author: Leon Ding <ding@ibyte.me>
// Date: 2022/2/26 - 10:32 下午 - UTC/GMT+08:00

package bottle

import (
	"errors"
	"fmt"
	"os"
	"sync"
)

var (

	// Data storage directory
	dataRoot = ""

	// Currently writable file
	active *os.File

	// Concurrent lock
	mutex sync.Mutex

	// Global indexes
	index map[uint64]*record

	// Current data file meter
	dataFileCount int32 = 0

	// Old data file mapping
	fileList map[int32]*os.File

	// Data file name extension
	dataFileSuffix = ".data"

	// FRW Read-only Opens a file in write - only mode
	FRW = os.O_RDWR | os.O_APPEND | os.O_CREATE

	// FR Open the file in read-only mode
	FR = os.O_RDONLY

	// Perm Default file operation permission
	Perm = os.FileMode(0755)

	// defaultMaxFileSize file size
	defaultMaxFileSize int64 = 2 << 8 << 20 // 2 << 8 = 512 << 20 = 536870912 kb
)

// record Mapping Data Record
type record struct {
	FID        uint32 // data file id
	Size       uint32 // data record size
	Offset     uint32 // data record offset
	Timestamp  uint32 // data record create timestamp
	ExpireTime uint32 // data record expire time
}

func Open(opt Option) error {

	opt.Validation()

	if ok, err := pathExists(dataRoot); ok {
		// 目录存在 恢复数据

	} else {

		// 如果有错误说明上面传入的文件不是目录或者非法
		if err != nil {
			panic("The current path is invalid!!!")
		}

		// Create folder if it does not exist
		if err := os.MkdirAll(dataRoot, Perm); err != nil {
			panic("Failed to create a working directory!!!")
		}

		// 目录创建好了就可以创建活跃文件写数据
		return createActiveFile()
	}

	return nil
}

// Create a new active file
func createActiveFile() error {
	if file, err := buildDataFile(); err == nil {
		active = file
		return nil
	}
	return errors.New("failed to create writable data file")
}

// Build a new datastore file
func buildDataFile() (*os.File, error) {
	mutex.Lock()
	defer mutex.Unlock()
	dataFileCount += 1
	return os.OpenFile(fmt.Sprintf("%s%d%s", dataRoot, dataFileCount, dataFileSuffix), FRW, Perm)
}

// File archiving is triggered when the data file is full
func exchangeFile() error {
	mutex.Lock()
	defer mutex.Unlock()
	active.Close()
	if file, err := os.OpenFile(fmt.Sprintf("%s%d%s", dataRoot, dataFileCount, dataFileSuffix), FR, Perm); err == nil {
		fileList[dataFileCount] = file
	}
	return createActiveFile()
}
