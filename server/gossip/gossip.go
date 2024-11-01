package gossip

import (
	"sync"
	"time"
)

type Status struct {
	Nodes      map[string]*Node // 所有已知节点的信息
	Self       *Node            // 本节点信息
	Mutex      sync.Mutex       // 保护 Nodes 的并发访问
	Interval   time.Duration    // 心跳间隔
	LossThresh time.Duration    // 失效阈值
}
