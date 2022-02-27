// Open Source: MIT License
// Author: Leon Ding <ding@ibyte.me>
// Date: 2022/2/26 - 10:32 下午 - UTC/GMT+08:00

package bottle

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path"
	"sync"
	"time"
)

// Global universal block
var (

	// Data storage directory
	dataRoot = ""

	// Currently writable file
	active *os.File

	// Concurrent lock
	mutex sync.RWMutex

	// Global indexes
	index map[uint64]*record

	// Current data file meter
	dataFileIdentifier int64 = -1

	// Old data file mapping
	fileList map[int64]*os.File

	// Data file name extension
	dataFileSuffix = ".data"

	// Index file name extension
	indexFileSuffix = ".index"

	// FRW Read-only Opens a file in write - only mode
	FRW = os.O_RDWR | os.O_APPEND | os.O_CREATE

	// FR Open the file in read-only mode
	FR = os.O_RDONLY

	// Perm Default file operation permission
	Perm = os.FileMode(0755)

	// Default max file size
	defaultMaxFileSize int64 = 2 << 8 << 20 // 2 << 8 = 512 << 20 = 536870912 kb

	// Default garbage collection merge threshold value
	defaultMergeThresholdValue = 1024

	// Index delete key count
	deleteKeyCount = 0

	// Default configuration file format
	defaultConfigFileSuffix = ".yaml"

	// HashedFunc Default Hashed function
	HashedFunc Hashed

	// Secret encryption key
	Secret = []byte("ME:QQ:2420498526")

	// itemPadding binary encoding header padding
	itemPadding uint32 = 20

	// Global data encoder
	encoder *Encoder

	// Write file offset
	writeOffset uint32 = 0
)

// Higher-order function blocks
var (

	// Opens a file by specifying a mode
	openDataFile = func(flag int) (*os.File, error) {
		return os.OpenFile(fileSuffixFunc(dataFileSuffix), flag, Perm)
	}

	// Builds the specified file name extension
	fileSuffixFunc = func(suffix string) string {
		return fmt.Sprintf("%s%d%s", dataRoot, dataFileIdentifier, suffix)
	}
)

// record Mapping Data Record
type record struct {
	FID        int64  // data file id
	Size       uint32 // data record size
	Offset     uint32 // data record offset
	Timestamp  uint32 // data record create timestamp
	ExpireTime uint32 // data record expire time
}

func Open(opt Option) error {

	opt.Validation()

	initialize()

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

// Load 通过配置文件加载
func Load(file string) error {

	if path.Ext(file) != defaultConfigFileSuffix {
		return errors.New("the current configuration file format is invalid")
	}

	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(file)

	if err := v.ReadInConfig(); err != nil {
		return err
	}

	var opt Option
	if err := v.Unmarshal(&opt); err != nil {
		return err
	}

	return Open(opt)
}

// Action Operation add-on
type Action struct {
	TTL time.Time // Survival time
}

// TTL You can set a timeout for the key in seconds
func TTL(second uint32) func(action *Action) {
	return func(action *Action) {
		action.TTL = time.Now().Add(time.Duration(second) * time.Second)
	}
}

// Put Add key-value data to the storage engine
// actionFunc You can set the timeout period
func Put(key, value []byte, actionFunc ...func(action *Action)) (err error) {

	var (
		action Action
		size   int
	)

	if len(actionFunc) > 0 {
		for _, fn := range actionFunc {
			fn(&action)
		}
	}

	fileInfo, _ := active.Stat()

	if fileInfo.Size() >= defaultMaxFileSize {
		if err := exchangeFile(); err != nil {
			return err
		}
	}

	sum64 := HashedFunc.Sum64(key)

	mutex.Lock()
	defer mutex.Unlock()

	timestamp := time.Now().Unix()

	if size, err = encoder.Write(NewItem(key, value, uint64(timestamp))); err != nil {
		return err
	}

	index[sum64] = &record{
		FID:        dataFileIdentifier,
		Size:       uint32(size),
		Offset:     writeOffset,
		Timestamp:  uint32(timestamp),
		ExpireTime: uint32(action.TTL.Unix()),
	}

	writeOffset += uint32(size)

	return nil
}

func Get(key []byte) *Data {
	var data Data

	mutex.RLock()
	defer mutex.RUnlock()

	sum64 := HashedFunc.Sum64(key)

	if index[sum64] == nil {
		data.Err = errors.New("the current key does not exist")
		return &data
	}

	if index[sum64].ExpireTime <= uint32(time.Now().Unix()) {
		data.Err = errors.New("the current key has expired")
		return &data
	}

	if item, err := encoder.Read(index[sum64]); err != nil {
		data.Err = err
	} else {
		data.Item = item
	}

	return &data
}

// Create a new active file
func createActiveFile() error {
	if file, err := buildDataFile(); err == nil {
		active = file
		fileList[dataFileIdentifier] = active
		return nil
	}
	return errors.New("failed to create writable data file")
}

// Build a new datastore file
func buildDataFile() (*os.File, error) {
	mutex.Lock()
	defer mutex.Unlock()
	dataFileIdentifier = time.Now().Unix()
	writeOffset = 0
	return openDataFile(FRW)
}

// File archiving is triggered when the data file is full
func exchangeFile() error {
	mutex.Lock()
	defer mutex.Unlock()
	_ = active.Close()
	if file, err := openDataFile(FR); err == nil {
		fileList[dataFileIdentifier] = file
	}
	return createActiveFile()
}

func initialize() {
	HashedFunc = DefaultHashFunc()
	encoder = DefaultEncoder()
	index = make(map[uint64]*record)
	fileList = make(map[int64]*os.File)
}
