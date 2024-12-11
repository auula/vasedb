package gossip

import (
	"encoding/json"
	"errors"
	"hash/fnv"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/auula/vasedb/clog"
)

// 最小节点数，奇数
const minNeighbors = 3

type (
	// gossip 协议集群中的节点
	Node struct {
		ID        string // 节点唯一标识
		Address   string // 节点地址
		Heartbeat int    // 心跳计数
		Timestamp int64  // 最新心跳时间
		Alive     bool   // 节点状态
		Hash      uint32 // 节点哈希值，用于一致性哈希
	}
	innerCluster struct {
		Neighbors map[string]*Node // 所有已知节点的信息
		Self      *Node            // 本节点信息
		Mutex     sync.Mutex       // 保护 Nodes 的并发访问
		Interval  time.Duration    // 心跳间隔
		Timeout   time.Duration    // 失效阈值
		HashRing  []*Node          // 一致性哈希存储范围的键
	}

	// Cluster 协议的集群集合
	Cluster struct {
		innerCluster
	}
)

func NewNode(id, addr string) *Node {
	node := &Node{
		ID:        id,
		Address:   addr,
		Heartbeat: 0,
		Timestamp: time.Now().Unix(),
		Alive:     true,
	}
	// 通过节点 ID 算出在哈希环中的值
	node.Hash = NodeHash(id)
	return node
}

func NewCluster(id, addr string, interval, timeout time.Duration) *Cluster {
	self := NewNode(id, addr)
	return &Cluster{
		innerCluster{
			Self:      self,
			Neighbors: map[string]*Node{id: self},
			Interval:  interval,
			Timeout:   timeout,
			// 将自身添加到哈希环，多个节点形成一个哈希环
			HashRing: []*Node{self},
		},
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
	if len(nodes) < minNeighbors {
		return errors.New("insufficient number of gossip cluster nodes")
	}

	// 设置节点哈希值并将节点加入到邻居中
	for _, node := range nodes {
		_, existing := c.Neighbors[node.ID]
		// 不存在节点就放入到集群中
		if !existing {
			c.Neighbors[node.ID] = &node
			// 节点对应哈希值也要放到一致性节点环中
			c.HashRing = append(c.HashRing, &node)
		}
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

	// 二分查找哈希环中第一个哈希值大于等于 keyHash 的节点
	idx := sort.Search(len(c.HashRing), func(i int) bool {
		// 计算出 key 所对应的哈希，并且找出对应的 Node
		return c.HashRing[i].Hash >= NodeHash(key)
	})

	// 如果找到的索引等于哈希环长度，则回到第一个节点
	if idx == len(c.HashRing) {
		return c.HashRing[0]
	}

	return c.HashRing[idx]
}

func (c *Cluster) Broadcast() {

	for {
		// 根据配置周期发送 gossip 数据包
		time.Sleep(c.Interval)
		c.Mutex.Lock()

		// 更新自身心跳计数和时间戳
		c.Self.Heartbeat++
		c.Self.Timestamp = time.Now().Unix()

		// 将节点信息广播给随机选定的其他节点
		var aliveNodes []*Node
		for _, node := range c.Neighbors {
			// 要求不是自己并且节点是存活
			if node.ID != c.Self.ID && node.Alive {
				aliveNodes = append(aliveNodes, node)
			}
		}

		for _, node := range aliveNodes {
			// 发送 gossip 协议数据包到附近节点上
			go c.SendPing(node.Address)
		}

		c.Mutex.Unlock()
	}

}

// SendPing 将节点信息编码为 JSON 并发送给指定节点
func (c *Cluster) SendPing(addr string) {
	conn, err := net.Dial("udp", addr)
	if err != nil {
		clog.Warn(err)
		return
	}
	defer conn.Close()

	neighbors, err := json.Marshal(c.Neighbors)
	if err != nil {
		clog.Warn(err)
		return
	}

	_, err = conn.Write(neighbors)
	if err != nil {
		clog.Warn(err)
		return
	}
}

// EchoPong 接收其他集群节点发送过来的 Ping 数据包
func (c *Cluster) EchoPong() error {
	// 打开一个 udp 服务器，接收其他节点 ping 数据包
	addr, err := net.ResolveUDPAddr("udp", c.Self.Address)
	if err != nil {
		clog.Failed(err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		clog.Failed(err)
	}
	defer conn.Close()

	// 创建一个缓冲期接收 ping 数据包
	buffer := make([]byte, 1024)
	for {
		n, _, _ := conn.ReadFromUDP(buffer)
		nodes := make(map[string]*Node)
		json.Unmarshal(buffer[:n], &nodes)
		// 更新 neighbors 状态
		updateStatus(c, nodes)
	}
}

func updateStatus(cluster *Cluster, nodes map[string]*Node) {

}
