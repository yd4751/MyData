package node

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/openworld-server/pkg/logger"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type NodeInfo struct {
	Name           string
	Addr           string
	CPUUsage       float64
	MemoryUsage    float64
	Connections    int
	ProcessID      int
	StartTime      time.Time
	LastUpdateTime time.Time
	Metadata       map[string]string
}

type Node struct {
	etcd         *clientv3.Client
	nodeInfo     *NodeInfo
	leaseID      clientv3.LeaseID
	updateTicker *time.Ticker
	running      bool
	connCounter  *ConnectionCounter
}

type ConnectionCounter struct {
	count int
}

func NewConnectionCounter() *ConnectionCounter {
	return &ConnectionCounter{count: 0}
}

func (c *ConnectionCounter) Increment() {
	c.count++
}

func (c *ConnectionCounter) Decrement() {
	if c.count > 0 {
		c.count--
	}
}

func (c *ConnectionCounter) GetCount() int {
	return c.count
}

func NewNode(name, addr, etcdAddr string, metadata map[string]string) (*Node, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(etcdAddr, ","),
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	nodeInfo := &NodeInfo{
		Name:           name,
		Addr:           addr,
		ProcessID:      os.Getpid(),
		StartTime:      time.Now(),
		LastUpdateTime: time.Now(),
		Metadata:       metadata,
	}

	node := &Node{
		etcd:        client,
		nodeInfo:    nodeInfo,
		running:     true,
		connCounter: NewConnectionCounter(),
	}

	if err := node.register(); err != nil {
		return nil, err
	}

	go node.updateLoop()

	return node, nil
}

func (n *Node) register() error {
	err := n.updateLoadInfo()
	if err != nil {
		logger.Warn("Failed to update load info:", err)
	}

	key := fmt.Sprintf("/services/%s/%s", n.nodeInfo.Name, n.nodeInfo.Addr)
	value := n.serializeNodeInfo()

	lease, err := n.etcd.Grant(context.Background(), 90)
	if err != nil {
		return err
	}
	n.leaseID = lease.ID

	_, err = n.etcd.Put(context.Background(), key, value, clientv3.WithLease(lease.ID))
	if err != nil {
		return err
	}

	go func() {
		for n.running {
			time.Sleep(30 * time.Second)
			n.etcd.KeepAliveOnce(context.Background(), n.leaseID)
		}
	}()

	logger.Info("Node registered: ", n.nodeInfo.Name, " ", n.nodeInfo.Addr)
	return nil
}

func (n *Node) updateLoop() {
	n.updateTicker = time.NewTicker(60 * time.Second)
	defer n.updateTicker.Stop()

	for n.running {
		select {
		case <-n.updateTicker.C:
			n.updateNodeInfo()
		}
	}
}

func (n *Node) updateNodeInfo() {
	err := n.updateLoadInfo()
	if err != nil {
		logger.Warn("Failed to update load info:", err)
	}

	key := fmt.Sprintf("/services/%s/%s", n.nodeInfo.Name, n.nodeInfo.Addr)
	value := n.serializeNodeInfo()

	_, err = n.etcd.Put(context.Background(), key, value, clientv3.WithLease(n.leaseID))
	if err != nil {
		logger.Error("Failed to update node info to etcd:", err)
		return
	}

	n.nodeInfo.LastUpdateTime = time.Now()
	logger.Debug("Node info updated: ", n.nodeInfo.Name, " CPU:", n.nodeInfo.CPUUsage, "%, Memory:", n.nodeInfo.MemoryUsage, "%")
}

func (n *Node) updateLoadInfo() error {
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return err
	}
	if len(cpuPercent) > 0 {
		n.nodeInfo.CPUUsage = cpuPercent[0]
	}

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return err
	}
	n.nodeInfo.MemoryUsage = memInfo.UsedPercent

	n.nodeInfo.Connections = n.connCounter.GetCount()

	return nil
}

func (n *Node) serializeNodeInfo() string {
	return fmt.Sprintf("%s|%f|%f|%d|%d|%d|%s",
		n.nodeInfo.Addr,
		n.nodeInfo.CPUUsage,
		n.nodeInfo.MemoryUsage,
		n.nodeInfo.Connections,
		n.nodeInfo.ProcessID,
		n.nodeInfo.StartTime.Unix(),
		buildMetadata(n.nodeInfo.Metadata),
	)
}

func buildMetadata(md map[string]string) string {
	var parts []string
	for k, v := range md {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, ",")
}

func ParseNodeInfo(name, value string) *NodeInfo {
	parts := strings.Split(value, "|")
	if len(parts) < 7 {
		return nil
	}

	addr := parts[0]

	var cpuUsage, memUsage float64
	fmt.Sscanf(parts[1], "%f", &cpuUsage)
	fmt.Sscanf(parts[2], "%f", &memUsage)

	var connections, pid, startTime int64
	fmt.Sscanf(parts[3], "%d", &connections)
	fmt.Sscanf(parts[4], "%d", &pid)
	fmt.Sscanf(parts[5], "%d", &startTime)

	metadata := parseMetadata(parts[6])

	return &NodeInfo{
		Name:           name,
		Addr:           addr,
		CPUUsage:       cpuUsage,
		MemoryUsage:    memUsage,
		Connections:    int(connections),
		ProcessID:      int(pid),
		StartTime:      time.Unix(startTime, 0),
		LastUpdateTime: time.Now(),
		Metadata:       metadata,
	}
}

func parseMetadata(str string) map[string]string {
	md := make(map[string]string)
	parts := strings.Split(str, ",")
	for _, part := range parts {
		kv := strings.Split(part, "=")
		if len(kv) == 2 {
			md[kv[0]] = kv[1]
		}
	}
	return md
}

func (n *Node) GetNodeInfo() *NodeInfo {
	return n.nodeInfo
}

func (n *Node) GetConnectionCounter() *ConnectionCounter {
	return n.connCounter
}

func (n *Node) Close() {
	n.running = false
	if n.updateTicker != nil {
		n.updateTicker.Stop()
	}
	if n.etcd != nil {
		n.etcd.Close()
	}
	logger.Info("Node stopped: ", n.nodeInfo.Name)
}

func (n *Node) GetStats() string {
	return fmt.Sprintf("Node[%s] CPU: %.2f%%, Memory: %.2f%%, Connections: %d, PID: %d",
		n.nodeInfo.Name,
		n.nodeInfo.CPUUsage,
		n.nodeInfo.MemoryUsage,
		n.nodeInfo.Connections,
		n.nodeInfo.ProcessID,
	)
}

func GetRuntimeStats() string {
	return fmt.Sprintf("Goroutines: %d, Memory Alloc: %dKB",
		runtime.NumGoroutine(),
		runtime.MemStats{}.Alloc/1024,
	)
}
