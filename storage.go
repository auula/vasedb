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

package bigmap

import "os"

var (
	// 全局的字典定位到记录映射
	indexMap map[uint32]*Record
	// 文件id对应的文件句柄，只读了
	files map[uint32]*os.File
	// 标记删除的key存储在这里
	rubbishList []uint32
	// 当前可写的文件
	af ActiveFile
)

type ActiveFile struct {
	*os.File
}

func OpenAF() *ActiveFile {
	return nil
}

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


func NewFile() *ActiveFile {
	return nil
}
