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
	"os"
	"sync"
)

var (
	// Used when initializing the data folder for the first time
	currentActiveFileFile *ActiveFile         // The current writable file
	fileList              map[uint32]*os.File // The file handle corresponding to the file ID is read-only
	index                 map[uint32]*Record  // Global dictionary location to record mapping
	rubbishList           []uint32            // The key marked for deletion is stored here
	LastOffset            uint32              // The file records the last offset
	lock                  *sync.RWMutex       // Concurrent control lock
)

// Record 映射记录实体
type Record struct {
	FID       uint32
	Size      uint32
	Offset    uint32
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

// Open the specified directory and initializes
// the corresponding directory as the data store archive destination.
// if the target directory already has data files,
// the program automatically restores the index map and initializes it.
func Open(path string) error {
	if path == "" {
		return ErrPathIsExists
	}
	if ok, err := PathExists(path); err != nil {
		return ErrPathNotAvailable
	} else if ok {
		// 文件夹存在
		// 1. 读取下面的索引文件，是否有索引文件查看
		// 2. 如果有索引，则回复到内存里面
		recoveryIndex()
	} else {
		// 不存在就创建文件夹
		if err := os.MkdirAll(path, perm); err != nil {
			return ErrCreateDirectoryFail
		}
	}
	// 文件夹不存在
	// 创建一个可写的文件 开始键索引
	if err := createActiveFile(path); err != nil {
		return ErrCreateActiveFileFail
	}
	// 从索引文件里面恢复数据文件，恢复内存索引
	return nil
}

// Save values to the storage engine by key
func Save(key, value []byte, as ...func(*Action) *Action) (err error) {
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
	return currentActiveFileFile.Close()
}
