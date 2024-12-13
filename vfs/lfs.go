package vfs

import (
	"bytes"
	"errors"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/auula/vasedb/conf"
	"github.com/auula/vasedb/utils"
)

var (
	once sync.Once

	indexShard = 5

	instance *LogStructuredFS

	// Data file name extension
	dataFileExtension = ".vsdb"

	dataFileMetadata = []byte{0xDB, 0x0, 0x0, 0x1}
)

// setupFS build vasedb file system
func setupFS(path string) error {
	if !utils.IsExist(path) {
		err := os.MkdirAll(path, conf.FsPerm)
		if err != nil {
			return err
		}
	}

	// 使用 os.ReadDir 代替 filepath.Walk 只遍历当前目录，不递归子目录
	files, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed to read directory: %v", err)
	}

	// 遍历目录下的文件
	for _, file := range files {
		// 检查是否是文件且文件名匹配 000x.vsdb
		if !file.IsDir() && strings.HasSuffix(file.Name(), dataFileExtension) {
			// 限制文件名格式，长度为 8，且以 "000" 开头
			if len(file.Name()) == 8 && strings.HasPrefix(file.Name(), "000") {
				// 打开文件并验证数据格式，获取完整路径
				file, err := os.Open(filepath.Join(path, file.Name()))
				if err != nil {
					return err
				}
				defer file.Close()

				// 验证单个文件的签名
				err = validateFileHeader(file)
				if err != nil {
					return err
				}
			}
		}
	}

	newLogStructuredFS()

	return nil
}

func validateFileHeader(file *os.File) error {
	var fileHeader [4]byte
	n, err := file.Read(fileHeader[:])
	if err != nil {
		return fmt.Errorf("failed to read file signature: %v", err)
	}

	if n != 4 {
		return errors.New("file is too short to contain valid signature")
	}

	if !bytes.Equal(fileHeader[:], dataFileMetadata[:]) {
		return fmt.Errorf("unsupported data file version: %v", file.Name())
	}

	return nil
}

// INode represents a file system node with metadata.
type INode struct {
	RegionID    uint16    // Unique identifier for the INode
	Offset      uint32    // Offset within the file
	CreatedTime time.Time // Creation time of the INode
	EexpireTime time.Time // Expiration time of the INode
}

type indexMap struct {
	mux   sync.RWMutex      // 每个分片使用独立的锁
	index map[uint64]*INode // 存储映射
}

// LogStructuredFS represents the virtual file storage system.
type LogStructuredFS struct {
	indexs       []*indexMap         // Index mapping for INode references
	regions      map[uint16]*os.File // Archived files keyed by unique file ID
	activeRegion *os.File            // Currently active file for writing
}

// 根据某种哈希函数（如简单的模运算）来选择分片
func (lfs *LogStructuredFS) getShardIndex(key uint64) *indexMap {
	return lfs.indexs[key%uint64(indexShard)]
}

// 使用 `getShardIndex` 获取分片，并加锁进行操作
func (lfs *LogStructuredFS) AddINode(key uint64, inode *INode) {
	shard := lfs.getShardIndex(key)
	shard.mux.Lock()
	defer shard.mux.Unlock()
	shard.index[key] = inode
}

func (lfs *LogStructuredFS) GetINode(key uint64) (*INode, bool) {
	shard := lfs.getShardIndex(key)
	shard.mux.RLock()
	defer shard.mux.RUnlock()
	inode, exists := shard.index[key]
	return inode, exists
}

func (lfs *LogStructuredFS) BatchINodes(inodes ...*INode) {

}

// ComputeHashForKey 计算节点的唯一哈希值
func ComputeHashForKey(key string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(key))
	return h.Sum64()
}

func newLogStructuredFS() {
	once.Do(func() {
		instance = &LogStructuredFS{
			indexs:       make([]*indexMap, indexShard),
			regions:      make(map[uint16]*os.File),
			activeRegion: new(os.File),
		}

		for i := 0; i < indexShard; i++ {
			instance.indexs[i] = &indexMap{
				mux:   sync.RWMutex{},
				index: make(map[uint64]*INode),
			}
		}
	})
}

func OpenFS(path string) (*LogStructuredFS, error) {
	setupFS(path)
	// 单例子模式，但是挡不住其他包通过 new(LogStructuredFS) 也能创建一个实例，那这样根本不起作用了
	return instance, nil
}

func (lfs *LogStructuredFS) CloseFS() error {

	// 关闭所有 regions
	for _, file := range lfs.regions {
		if err := utils.CloseFile(file); err != nil {
			return fmt.Errorf("failed to close region file: %w", err)
		}
	}

	// 关闭 activeRegion
	if err := utils.CloseFile(lfs.activeRegion); err != nil {
		return fmt.Errorf("failed to close active region: %w", err)
	}

	return nil
}
