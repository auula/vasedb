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

// BigMap It is the storage instance
// Is the implementation of the entire storage engine
// through the data store read delete operation interface.

package bottle

import (
	"os"
	"strings"
	"sync"
	"time"
)

const (
	// EntityPadding : 1 uint32 = 4 byte = 8 bit
	// 5 field * uint32 = 4 * 4 = 16 byte = 16 * 8 = 128 bit
	EntityPadding = 1 << 4
)

var (
	fileLists   map[string]*os.File // List of global read-only files
	rubbishList []uint64            // The key marked for deletion is stored here
	onceFunc    sync.Once           // A function wrapper for execution once
	hashedFunc  Hashed              // A function used to compute a key hash
	encoder     *Encoder            // Data recording codec
)

// Record 映射记录实体
type Record struct {
	FID        string
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

type bottle struct {
	af     *ActiveFile        // The current writable file
	index  map[uint64]*Record // Global dictionary location to record mapping
	offset uint32             // The file records the last offset
	mutex  *sync.RWMutex      // Concurrent control lock
}

// Compaction 压缩进程
type Compaction struct {
	// 需要清理的可以通道
	rubbishs <-chan uint32
	// 需要删除的字典
	keydir map[uint32]*Record
	// 快速恢复索引使用
	hint map[uint32]*os.File
}

type Options struct {
	FileMaxSize int32  `json:"file_max_size,omitempty"`
	Path        string `json:"path,omitempty"`
	// 是否开启加密和秘钥
	Secret     []byte
	EnableSafe bool
}

type Seconds struct {
	TTL uint32
}

func TTL(sec uint32) func(seconds *Seconds) {
	return func(seconds *Seconds) {
		seconds.TTL = sec
	}
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

	if ok, err := PathExists(path); err != nil {
		return nil, ErrPathNotAvailable
	} else if ok {
		// Folder exists
		// 1. Read the following index file, whether there is an index file to view
		// 2. If there is an index, it is returned to memory
		recoveryIndex()
	} else {
		// 不存在就创建文件夹
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

	// Record the location of the data file
	dataPath = strings.TrimSpace(path)

	// Restore data file from index file, restore memory index
	return &storage, nil
}

// Entity a data entity struct
type Entity struct {
	entityItem
}

// entityItem a data item
type entityItem struct {
	CRC32      uint32 // 循环校验码
	KeySize    uint32 // 键的大小
	ValueSize  uint32 // 值的大小
	TimeStamp  uint32 // 创建时间戳
	Key, Value []byte // 键字符串,值序列化
}

// NewEntity build a data entity
func NewEntity(key, value []byte, timestamp uint32) *Entity {
	var entity Entity
	entity.TimeStamp = timestamp
	entity.Key = key
	entity.Value = value
	entity.KeySize = uint32(len(key))
	entity.ValueSize = uint32(len(value))
	return &entity
}

// Put values to the storage engine by key
func (s *Storage) Put(key, value []byte, secs ...func(seconds *Seconds)) (err error) {
	var (
		sec  Seconds
		size int
	)
	sum64 := hashedFunc.Sum64(key)
	// 如果用户设置了超时时间那么就要操作超时计算
	if len(secs) > 0 {
		for _, do := range secs {
			do(&sec)
		}
		sec.TTL = uint32(time.Now().Add(time.Duration(sec.TTL)).Unix())
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
		ExpireTime: sec.TTL,
	}
	s.offset += uint32(size)
	s.mutex.Unlock()
	return
}

// Get retrieves the corresponding value by key
func (s *Storage) Get(key []byte) (*Entity, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	sum64 := hashedFunc.Sum64(key)
	if _, isExist := s.index[sum64]; !isExist {
		return nil, ErrKeyNoData
	}
	return encoder.Read(s.index[sum64])
}

// Remove the corresponding value by key
func (s *Storage) Remove(key []byte) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	sum64 := hashedFunc.Sum64(key)
	// 如果存在就清理
	if _, isExist := s.index[sum64]; isExist {
		delete(s.index, sum64)
		rubbishList = append(rubbishList, sum64)
	}
}

// FlushAll memory index and record files are all written to disk
// safely shut down the storage engine
func FlushAll() {

}

// Close current active file
func (s *Storage) Close() error {
	return s.af.Close()
}

// Hashed is responsible for generating unsigned, 64-bit hash of provided string. Harsher should minimize collisions
// (generating same hash for different strings) and while performance is also important fast functions are preferable (i.e.
// you can use FarmHash family).
type Hashed interface {
	Sum64([]byte) uint64
}

// DefaultHashFunc returns a new 64-bit FNV-1a Hashed which makes no memory allocations.
// Its Sum64 method will lay the value out in big-endian byte order.
// See https://en.wikipedia.org/wiki/Fowler–Noll–Vo_hash_function
func DefaultHashFunc() Hashed {
	return fnv64a{}
}

type fnv64a struct{}

const (
	// offset64 FNVa offset basis.
	// See https://en.wikipedia.org/wiki/Fowler–Noll–Vo_hash_function#FNV-1a_hash
	offset64 = 14695981039346656037
	// prime64 FNVa prime value.
	// See https://en.wikipedia.org/wiki/Fowler–Noll–Vo_hash_function#FNV-1a_hash
	prime64 = 1099511628211
)

// Sum64 gets the string and returns its uint64 hash value.
func (f fnv64a) Sum64(key []byte) uint64 {
	var hash uint64 = offset64
	for i := 0; i < len(key); i++ {
		hash ^= uint64(key[i])
		hash *= prime64
	}
	return hash
}

func (s *Storage) initialize() {
	s.bottle = new(bottle)
	s.offset = uint32(0)
	s.mutex = new(sync.RWMutex)
	s.index = make(map[uint64]*Record)
	fileLists = make(map[string]*os.File)
	encoder = AESEncoder()
	hashedFunc = DefaultHashFunc()
}

// SetIndexSize initialize the size of the memory index
func (s *Storage) SetIndexSize(size uint16) {
	s.mutex.Lock()
	s.index = make(map[uint64]*Record, size)
	s.mutex.Unlock()
}
