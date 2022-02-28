// Open Source: MIT License
// Author: Leon Ding <ding@ibyte.me>
// Date: 2022/2/26 - 10:32 下午 - UTC/GMT+08:00

package bottle

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
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

	// Data recovery triggers the merge threshold
	totalDataSize int64 = 2 << 8 << 20 << 3 // 4GB

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
	openDataFile = func(flag int, dataFileIdentifier int64) (*os.File, error) {
		return os.OpenFile(fileSuffixFunc(dataFileSuffix, dataFileIdentifier), flag, Perm)
	}

	// Builds the specified file name extension
	fileSuffixFunc = func(suffix string, dataFileIdentifier int64) string {
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
		return recoverData()
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

	if size, err = encoder.Write(NewItem(key, value, uint64(timestamp)), active); err != nil {
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

func Remove(key []byte) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(index, HashedFunc.Sum64(key))
}

func Close() error {
	for _, file := range fileList {
		file.Close()
	}
	return saveIndexToFile()
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
	dataFileIdentifier = time.Now().Unix()
	writeOffset = 0
	return openDataFile(FRW, dataFileIdentifier)
}

// File archiving is triggered when the data file is full
func exchangeFile() error {
	mutex.Lock()
	defer mutex.Unlock()
	_ = active.Close()
	if file, err := openDataFile(FR, dataFileIdentifier); err == nil {
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

// Index item
type indexItem struct {
	idx uint64
	*record
}

// Save index files to the data directory
func saveIndexToFile() (err error) {

	var file *os.File
	defer file.Close()

	var channel = make(chan indexItem, 1024)

	go func() {
		for sum64, record := range index {
			channel <- indexItem{
				idx:    sum64,
				record: record,
			}
		}
		close(channel)
	}()

	if file, err = buildIndexFile(); err != nil {
		return
	}

	for v := range channel {
		if _, err = encoder.WriteIndex(v, file); err != nil {
			return
		}
	}

	return
}

func buildIndexFile() (*os.File, error) {
	// 索引文件夹
	indexDirectory := fmt.Sprintf("%sindexs/", dataRoot)

	// 不存在就创建
	if ok, _ := pathExists(indexDirectory); !ok {
		_ = os.MkdirAll(indexDirectory, Perm)
	}

	// 构建索引文件
	indexPath := fmt.Sprintf("%sindexs/%d%s", dataRoot, time.Now().Unix(), indexFileSuffix)
	return os.OpenFile(indexPath, FRW, Perm)
}

func recoverData() error {

	// 恢复数据流程
	// 1. 从数据文件夹里面把数据文件排序
	// 2. 找到最新的那个数据文件并且检测是否满了
	// 3. 如果满了创建新的可写文件，其他数据文件归档
	// 4. 并且把当先全局可写文件激活

	if dataTotalSize() >= totalDataSize {
		// 触发合并
		return merge()
	}

	if file, err := findLatestDataFile(); err == nil {
		info, _ := file.Stat()
		if info.Size() >= defaultMaxFileSize {
			if err := createActiveFile(); err != nil {
				return err
			}
		}
		active = file
		writeOffset = uint32(info.Size())
	}

	return buildIndex()
}

func merge() error {

	if err := readIndexItem(); err == nil {
		return err
	}

	var (
		size     int
		offset   uint32
		fileInfo os.FileInfo
	)

	newDataFileID := time.Now().Unix()

	// 创建迁移的目标数据文件
	file, _ := openDataFile(FRW, newDataFileID)

	fileInfo, _ = file.Stat()

	for idx, rec := range index {

		// 每轮检测迁移文件是否阀值了
		if fileInfo.Size() >= defaultMaxFileSize {
			newDataFileID = time.Now().Unix()
			file, _ = openDataFile(FRW, newDataFileID)
			fileInfo, _ = file.Stat()
		}

		item, _ := encoder.Read(rec)

		// 新文件ID
		rec.FID = newDataFileID

		// 把原来的内容写到新文件
		size, _ = encoder.Write(item, file)

		// 更新偏移量
		rec.Size = uint32(size)
		rec.Offset = offset
		index[idx] = rec

		offset += uint32(size)
	}

	return nil
}

func buildIndex() error {

	// 索引恢复流程
	// 1. 找到索引文件夹
	// 2. 从一堆文件夹里找到最新的那个索引文件
	// 3. 然后恢复索引到内存
	// 4. 并且打开索引映射的文件为只读状态

	if err := readIndexItem(); err == nil {
		return err
	}

	for _, record := range index {
		// Exclude the currently active file
		if dataFileIdentifier != record.FID {
			fp := fmt.Sprintf("%s%d%s", dataRoot, record.FID, dataFileSuffix)
			if file, err := os.OpenFile(fp, FR, Perm); err != nil {
				return err
			} else {
				// Open the original data file
				fileList[record.FID] = file
			}
		}
	}

	return nil
}

// Find the latest data files in the index folder
func findLatestIndexFile() (*os.File, error) {

	indexDirectory := fmt.Sprintf("%sindexs/", dataRoot)

	files, err := ioutil.ReadDir(indexDirectory)

	if err != nil {
		return nil, err
	}

	var indexes []fs.FileInfo

	for _, file := range files {
		if path.Ext(file.Name()) == indexFileSuffix {
			indexes = append(indexes, file)
		}
	}

	var ids []int

	for _, info := range indexes {
		id := strings.Split(info.Name(), ".")[0]
		i, err := strconv.Atoi(id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, i)
	}

	sort.Ints(ids)

	indexPath := fmt.Sprintf("%sindexs/%d%s", dataRoot, ids[len(ids)-1], indexFileSuffix)

	return os.OpenFile(indexPath, FR, Perm)
}

// Read index file contents into memory index
func readIndexItem() error {

	if file, err := findLatestIndexFile(); err == nil {
		defer func() {
			_ = file.Close()
		}()

		buf := make([]byte, 36)

		for {

			_, err := file.Read(buf)

			if err != nil && err != io.EOF {
				return err
			}

			if err == io.EOF {
				break
			}

			if err = encoder.ReadIndex(buf); err != nil {
				return err
			}

		}

		return nil
	}

	return errors.New("index reading failed")
}

// Find the latest data file from the data file
func findLatestDataFile() (*os.File, error) {

	files, _ := ioutil.ReadDir(dataRoot)

	var datafiles []fs.FileInfo

	for _, file := range files {
		if path.Ext(file.Name()) == dataFileSuffix {
			datafiles = append(datafiles, file)
		}
	}

	var ids []int

	for _, info := range datafiles {
		id := strings.Split(info.Name(), ".")[0]
		i, err := strconv.Atoi(id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, i)
	}

	sort.Ints(ids)

	activePath := fmt.Sprintf("%s%d%s", dataRoot, ids[len(ids)-1], dataFileSuffix)

	// Reset file counters and writable files and offsets
	dataFileIdentifier = int64(ids[len(ids)-1])

	return os.OpenFile(activePath, FRW, Perm)
}

// Calculate all data file sizes from the data folder
func dataTotalSize() int64 {

	files, _ := ioutil.ReadDir(dataRoot)

	var datafiles []fs.FileInfo

	for _, file := range files {
		if path.Ext(file.Name()) == dataFileSuffix {
			datafiles = append(datafiles, file)
		}
	}

	var totalSize int64

	for _, datafile := range datafiles {
		totalSize += datafile.Size()
	}

	return totalSize
}
