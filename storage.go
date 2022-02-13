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

package bigmap

import (
	"hash/fnv"
	"os"
	"sync"
	"time"
)

var (
	currentActiveFile *ActiveFile         // The current writable file
	fileList          map[uint64]*os.File // The file handle corresponding to the file ID is read-only
	indexMap          map[uint64]*Record  // Global dictionary location to record mapping
	rubbishList       []uint64            // The key marked for deletion is stored here
	LastOffset        uint64              // The file records the last offset
	globalLock        *sync.RWMutex       // Concurrent control lock
	onceFunc          sync.Once           // A function wrapper for execution once
	HashedFunc        Hashed              // A function used to compute a key hash
)

// Record 映射记录实体
type Record struct {
	FID       string
	Size      uint32
	Offset    uint64
	Timestamp uint64
}

// Entity 数据实体
type Entity struct {
	CRC   int32  // 循环校验码
	KS    int8   // 键的大小
	VS    int16  // 值的大小
	TTL   uint64 // 超时时间戳，截止时间
	Key   string // 键字符串
	Value []byte // 值序列化
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

type Action struct{}

// Open the specified directory and initializes.
// Used when initializing the data folder for the first time.
// The corresponding directory as the data store archive destination,
// if the target directory already has data files,
// the program automatically restores the index map and initializes it.
func Open(path string) error {

	if path == "" {
		return ErrPathIsExists
	}

	if ok, err := PathExists(path); err != nil {
		return ErrPathNotAvailable
	} else if ok {
		// Folder exists
		// 1. Read the following index file, whether there is an index file to view
		// 2. If there is an index, it is returned to memory
		recoveryIndex()
	} else {
		// 不存在就创建文件夹
		if err := os.MkdirAll(path, perm); err != nil {
			return ErrCreateDirectoryFail
		}
	}
	// Folder does not exist
	// Create a writable file start key index
	if err := createActiveFile(path); err != nil {
		return ErrCreateActiveFileFail
	}

	// Initialize read/write locks only once
	onceFunc.Do(Initialize)

	// Restore data file from index file, restore memory index
	return nil
}

// Save values to the storage engine by key
func Save(key string, value []byte, as ...func(*Action) *Action) (err error) {
	sum64 := HashedFunc.Sum64(key)
	globalLock.Lock()
	indexMap[sum64] = &Record{
		FID:       currentActiveFile.fid,
		Size:      offset64,
		Offset:    LastOffset,
		Timestamp: uint64(time.Now().UnixNano()),
	}
	LastOffset += offset64
	globalLock.Unlock()
	return
}

// Get retrieves the corresponding value by key
func Get(key []byte) (bytes []byte, err error) {

	return []byte{}, nil
}

// Remove the corresponding value by key
func Remove(key []byte) (err error) {

	return
}

// FlushAll memory index and record files are all written to disk
// safely shut down the storage engine
func FlushAll() {

}

// Close current active file
func Close() error {
	return currentActiveFile.Close()
}

// Hashed is responsible for generating unsigned, 64-bit hash of provided string. Harsher should minimize collisions
// (generating same hash for different strings) and while performance is also important fast functions are preferable (i.e.
// you can use FarmHash family).
type Hashed interface {
	Sum64(string) uint64
}

// DefaultHashFunc returns a new 64-bit FNV-1a Hashed which makes no memory allocations.
// Its Sum64 method will lay the value out in big-endian byte order.
// See https://en.wikipedia.org/wiki/Fowler–Noll–Vo_hash_function
func DefaultHashFunc() Hashed {
	return fnv64a{}
}

type fnv64a struct{}

const (
	// offset64 FNVa offset basis. See https://en.wikipedia.org/wiki/Fowler–Noll–Vo_hash_function#FNV-1a_hash
	offset64 = 14695981039346656037
	// prime64 FNVa prime value. See https://en.wikipedia.org/wiki/Fowler–Noll–Vo_hash_function#FNV-1a_hash
	prime64 = 1099511628211
)

// Sum64 gets the string and returns its uint64 hash value.
func (f fnv64a) Sum64(key string) uint64 {
	var hash uint64 = offset64
	for i := 0; i < len(key); i++ {
		hash ^= uint64(key[i])
		hash *= prime64
	}
	return hash
}

// Use the standard library to compute the hashing algorithm
func stdLibFnvSum64(key string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(key))
	return h.Sum64()
}

func Initialize() {
	globalLock = new(sync.RWMutex)
	HashedFunc = DefaultHashFunc()
	LastOffset = uint64(0)
	indexMap = make(map[uint64]*Record)
}
