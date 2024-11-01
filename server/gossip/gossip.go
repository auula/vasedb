package gossip

import (
	"errors"
	"hash/fnv"
	"sort"
	"sync"
	"time"
)

const MinNeighbors = 3 // 最小节点数，奇数

type Node struct {
	ID        string // 节点唯一标识
	Address   string // 节点地址
	Heartbeat int    // 心跳计数
	Timestamp int64  // 最新心跳时间
	Alive     bool   // 节点状态
	Hash      uint32 // 节点哈希值，用于一致性哈希
}

type Cluster struct {
	Neighbors map[string]*Node // 所有已知节点的信息
	Self      *Node            // 本节点信息
	Mutex     sync.Mutex       // 保护 Nodes 的并发访问
	Interval  time.Duration    // 心跳间隔
	Timeout   time.Duration    // 失效阈值
	HashRing  []*Node          // 一致性哈希存储范围的键
}

func NewNode(id, addr string) *Node {
	node := &Node{
		ID:        id,
		Address:   addr,
		Heartbeat: 0,
		Timestamp: time.Now().Unix(),
		Alive:     true,
	}
	// 计算节点的哈希值
	node.Hash = NodeHash(id)
	return node
}

func NewCluster(id, addr string, interval, timeout time.Duration) *Cluster {
	self := NewNode(id, addr)
	return &Cluster{
		Neighbors: map[string]*Node{id: self},
		Self:      self,
		Interval:  interval,
		Timeout:   timeout,
		HashRing:  []*Node{self}, // 将自身添加到哈希环，多个节点形成一个哈希环
	}
}

// NodeHash 计算节点的唯一哈希值
func NodeHash(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

// AddNodes 批量添加数据节点
func (c *Cluster) AddNodes(nodes ...Node) error {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	// 必须是奇数数量的节点集群
	if len(nodes) < MinNeighbors {
		return errors.New("")
	}

	// 设置节点哈希值并将节点加入到邻居中
	for _, node := range nodes {
		// 通过节点 ID 算出在哈希环中的值
		node.Hash = NodeHash(node.ID)
		c.Neighbors[node.ID] = &node
		// 节点对应哈希值也要放到一致性节点环中
		c.HashRing = append(c.HashRing, &node)
	}

	// 对一致性哈希环中的节点按哈希值排序
	sort.Slice(c.HashRing, func(i, j int) bool {
		return c.HashRing[i].Hash < c.HashRing[j].Hash
	})

	return nil
}

// GetNode 根据键查找哈希环上的最近节点
func (c *Cluster) GetNode(key string) *Node {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	if len(c.HashRing) == 0 {
		return nil
	}

	// 计算出 key 所对应的哈希
	keyHash := NodeHash(key)
	// 二分查找哈希环中第一个哈希值大于等于 keyHash 的节点
	idx := sort.Search(len(c.HashRing), func(i int) bool {
		return c.HashRing[i].Hash >= keyHash
	})

	// 如果找到的索引等于哈希环长度，则回到第一个节点
	// 因为是环形结构首尾相连
	if idx == len(c.HashRing) {
		return c.HashRing[0]
	}

	return c.HashRing[idx]
}
