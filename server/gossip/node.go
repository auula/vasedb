package gossip

type Node struct {
	ID        string // 节点唯一标识
	Address   string // 节点地址
	Heartbeat int    // 心跳计数
	Timestamp int64  // 最新心跳时间
	Alive     bool   // 节点状态
}
