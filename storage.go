// MIT License

// Copyright (c) 2022 Leon Ding

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Bottle It is the storage instance
// Is the implementation of the entire storage engine
// through the data store read delete operation interface.

package bottle

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	// Default file operation permission
	perm = os.FileMode(0755)

	// Default data file path
	dataPath = ""

	// Index and data file name extensions
	dataFileSuffix  = ".bdf"
	indexFileSuffix = ".idx"

	// Data segment encryption key
	secret = []byte("1234567890123456")

	globalConfig *Options

	// Default file size
	defaultMaxFileSize = 2 << 8 << 20 // 2 << 8 = 512 << 20 = 536870912 kb

	// Open the file in read-only mode
	fileOnlyRead = os.O_RDONLY
	// Read-only Opens a file in write - only mode
	fileOnlyReadANDWrite = os.O_RDWR | os.O_APPEND | os.O_CREATE

	ErrEntityDataBufToFile  = errors.New("error 203: error writing entity record data from buffer to file")
	ErrCreateActiveFileFail = errors.New("error 104: failed to create a writable and readable active file")
	ErrSourceDataEncodeFail = errors.New("error 201: source data fails to be encrypted by the encoder")
	ErrSourceDataDecodeFail = errors.New("error 202: source data failed to be decrypted by encoder")
	ErrPathNotAvailable     = errors.New("error 102: the current directory path is unavailable")
	ErrCreateDirectoryFail  = errors.New("error 103: failed to create a data store directory")
	ErrKeyNoData            = errors.New("error 301: the queried key does not have data")
	ErrNoDataEntityWasFound = errors.New("error 204: no data entity was found")
	ErrPathIsExists         = errors.New("error 101: an empty path is illegal")
	ErrIndexEncode          = errors.New("error 401: error saving index")
)

const (
	// EntityPadding : 1 uint32 = 4 byte = 8 bit
	// 5 field * uint32 = 4 * 4 = 16 byte = 16 * 8 = 128 bit
	EntityPadding = 1 << 4
)

var (
	fileLists  map[uint64]*os.File // List of global read-only files
	hashedFunc Hashed              // A function used to compute a key hash
	encoder    *Encoder            // Data recording codec
	tm         *TimeoutMgr         // Global expire time manager
)

// Record Mapping record entity
type Record struct {
	FID        uint64
	Size       uint32
	Offset     uint32
	Timestamp  uint32
	ExpireTime uint32
}

// Storage The storage engine operates on objects
// This structure is responsible for interacting with operating system files
// A valid Bottle stores engine objects
type Storage struct {
	*bottle
}

// Bottle Data directory operation client
type bottle struct {
	af           *activeFile        // The current writable file
	index        map[uint64]*Record // Global dictionary location to record mapping
	offset       uint32             // The file records the last offset
	mutex        *sync.RWMutex      // Concurrent control lock
	garbageTruck chan uint64        // The expired key cleans up the message channel
	GcState      bool               // The running status of garbage collection
}

// Compaction The compression process
type Compaction struct {
}

// TimeoutMgr Timeout manager
type TimeoutMgr struct {
	index map[uint64]*time.Timer
	mutex *sync.RWMutex
}

type indexItem struct {
	fid        uint64
	idx        uint64
	Size       uint32
	Offset     uint32
	CRC32      uint32
	ExpireTime uint32
}

// Add a  time task to the timeout manager
func addTimer(idx uint64, t *time.Timer) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	tm.index[idx] = t
}

// Stop the timeout by index ID
func (m *TimeoutMgr) Stop(idx uint64) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	if m.index[idx] != nil {
		m.index[idx].Stop()
		delete(m.index, idx)
	}
}

// Action Data manipulation attachment options
type Action struct {
	TTL time.Time
}

// TTL You can set a timeout for the key in seconds
func TTL(sec uint32) func(action *Action) {
	return func(action *Action) {
		action.TTL = time.Now().Add(time.Duration(sec) * time.Second)
	}
}

type activeFile struct {
	fid uint64
	*os.File
}

func createActiveFile(path string, storage Storage) error {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()

	storage.af = new(activeFile)
	storage.af.fid = hashedFunc.Sum64([]byte(uuid.NewString()))

	filePath := dataFilePath(path, storage.af)

	if file, err := os.OpenFile(filePath, fileOnlyReadANDWrite, perm); err != nil {
		return err
	} else {
		storage.af.File = file
		fileLists[storage.af.fid] = storage.af.File
	}
	return nil
}

// Open the destination path file in read-only mode
func openOnlyReadFile(path string) (*os.File, error) {
	return os.OpenFile(path, fileOnlyRead, perm)
}

// Open the specified directory and initializes.
// Used when initializing the data folder for the first time.
// The corresponding directory as the data store archive destination,
// if the target directory already has data files,
// the program automatically restores the index map and initializes it.
func Open(path string) (*Storage, error) {
	var storage Storage

	if path == "" {
		return nil, ErrPathIsExists
	}

	// The first one does not determine whether there is a backslash
	path = pathBackslashes(path)

	// Record the location of the data file
	dataPath = strings.TrimSpace(path)

	if ok, err := pathNotExist(path); err != nil {
		return nil, ErrPathNotAvailable
	} else if ok {
		// Read the following index file, whether there is an index file to view
		// If there is an index, it is returned to memory
		recoveryData()
	} else {
		// Create folder if it does not exist
		if err := os.MkdirAll(path, perm); err != nil {
			return nil, ErrCreateDirectoryFail
		}
	}

	// Initialize read/write locks only once
	storage.initialize()

	// Folder does not exist
	// Create a writable file start key index
	if err := createActiveFile(path, storage); err != nil {
		return nil, ErrCreateActiveFileFail
	}

	// Restore data file from index file, restore memory index
	return &storage, nil
}

// Item a data entity struct
type Item struct {
	item
}

// Item a data item
type item struct {
	CRC32      uint32 // Cyclic check code
	KeySize    uint32 // The size of the key
	ValueSize  uint32 // The size of the value
	TimeStamp  uint32 // Create timestamp
	Key, Value []byte // Key string, value serialization
}

// NewEntity build a data entity
func NewEntity(key, value []byte, timestamp uint32) *Item {
	var entity Item
	entity.Key = key
	entity.Value = value
	entity.TimeStamp = timestamp
	entity.KeySize = uint32(len(key))
	entity.ValueSize = uint32(len(value))
	return &entity
}

// Put values to the storage engine by key
func (s *Storage) Put(key, value []byte, secs ...func(action *Action)) (err error) {

	var (
		action Action
		size   int
	)

	sum64 := hashedFunc.Sum64(key)

	// If the user sets a timeout period then the timeout calculation is performed
	if len(secs) > 0 {
		for _, do := range secs {
			do(&action)
		}
		// Create coroutines to initiate scheduled cleanup
		go func() {
			tm.Stop(sum64)
			// 映射到一个索引time管理器
			// 此处有bug如果key重复put定时器会无限增多,一对多了
			// 我知道如果是其他人做的话是惰性删除，存在真实清理偏差，所以我采用的实时信号通知的方式删除
			timer := time.NewTimer(time.Until(action.TTL))
			addTimer(sum64, timer)
			<-timer.C
			s.garbageTruck <- sum64
			timer.Stop()
		}()
	}

	s.mutex.Lock()
	timestamp := time.Now().Unix()
	if size, err = encoder.Write(NewEntity(key, value, uint32(timestamp)), s.af); err != nil {
		return err
	}
	s.index[sum64] = &Record{
		FID:        s.af.fid,
		Size:       uint32(size),
		Offset:     s.offset,
		Timestamp:  uint32(timestamp),
		ExpireTime: uint32(action.TTL.Unix()),
	}
	s.offset += uint32(size)
	s.mutex.Unlock()

	return
}

// Get retrieves the corresponding value by key
func (s *Storage) Get(key []byte) (*Item, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	sum64 := hashedFunc.Sum64(key)
	if !indexIDExist(sum64, s.index) {
		return nil, ErrKeyNoData
	}
	return encoder.Read(s.index[sum64])
}

// Remove the corresponding value by key
func (s *Storage) Remove(key []byte) {
	sum64 := hashedFunc.Sum64(key)
	// 通知gc工作线程
	s.garbageTruck <- sum64
	tm.Stop(sum64)
}

// Sync memory index and record files are all written to disk
func (s *Storage) Sync() error {

	return saveIndexToFile(s.index)
}

// CTX Safely close context
var CTX = &Context{
	Exit: make(chan struct{}),
}

type Context struct {
	Exit chan struct{}
}

// Close current active file
// safely shut down the storage engine
func (s *Storage) Close(ctx Context) error {
	ctx.Exit <- struct{}{}
	return s.af.Close()
}

// Initialize storage
func (s *Storage) initialize() {
	s.bottle = new(bottle)
	s.offset = uint32(0)
	s.mutex = new(sync.RWMutex)
	s.index = make(map[uint64]*Record)
	s.garbageTruck = make(chan uint64, 10)
	tm = &TimeoutMgr{
		index: make(map[uint64]*time.Timer),
		mutex: new(sync.RWMutex),
	}
	go s.ActionTruck(context.Background(), 1)
	fileLists = make(map[uint64]*os.File)
	encoder = AESEncoder()
	hashedFunc = DefaultHashFunc()
}

// SetIndexSize initialize the size of the memory index
func (s *Storage) SetIndexSize(size uint16) {
	s.mutex.Lock()
	s.index = make(map[uint64]*Record, size)
	s.mutex.Unlock()
}

// ActionTruck garbage collection does not work by default
// sleep: garbage collection idle time
func (s *Storage) ActionTruck(ctx context.Context, sleep int) {
	changeState(s, true)
	for s.GcState {
		select {
		case <-ctx.Done():
			changeState(s, false)
			return
			// If the remaining keys have not expired, record them separately
		case sum64 := <-s.garbageTruck:
			// fmt.Println("清理:", sum64)
			if indexIDExist(sum64, s.index) {
				s.mutex.Lock()
				delete(s.index, sum64)
				s.mutex.Unlock()
			}
		default:
			time.Sleep(time.Duration(sleep) * time.Second)
		}
	}
}

// indexIDExist check index id whether exist
func indexIDExist(sum64 uint64, index map[uint64]*Record) bool {
	if _, ok := index[sum64]; ok {
		return true
	}
	return false
}

// changeState modify the GC running status
func changeState(s *Storage, state bool) {
	s.mutex.Lock()
	s.GcState = state
	s.mutex.Unlock()
}

// pathBackslashes Check directory ending backslashes
func pathBackslashes(path string) string {
	if !strings.HasSuffix(path, "/") {
		return fmt.Sprintf("%s/", path)
	}
	return path
}

// Build the data file address path
func dataFilePath(path string, file *activeFile) string {
	return fmt.Sprintf("%s%d%s", path, file.fid, dataFileSuffix)
}

// Indexes can be recovered from multiple files in parallel
func indexFilePath(path string) string {
	if ok, _ := pathNotExist(path); ok {
		_ = os.MkdirAll(fmt.Sprintf("%sindexs/", path), perm)
	}
	return fmt.Sprintf("%sindexs/%s%s", path, uuid.NewString(), indexFileSuffix)
}

//func hintFilePath(path string) string {
//	return fmt.Sprintf("%s%s%s", path, currentFile.fid, hintFileSuffix)
//}

// Recover data from data files
func recoveryData() {
}

// Checks whether the target path exists
func pathNotExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
