package connector

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/openworld-server/internal/cluster"
	"github.com/openworld-server/internal/msg"
	"github.com/openworld-server/pkg/logger"
)

type Connection interface {
	ID() string
	Addr() string
	NodeType() msg.NodeType
	NodeID() string
	RawConn() net.Conn
	Send(msgID uint32, nodeType msg.NodeType, data []byte) error
	Receive(timeout time.Duration) (uint32, msg.NodeType, []byte, error)
	Close() error
	IsValid() bool
	LastUsed() time.Time
	UpdateLastUsed()
}

type connection struct {
	id       string
	addr     string
	nodeType msg.NodeType
	nodeID   string
	conn     net.Conn
	lastUsed time.Time
	mu       sync.RWMutex
}

func newConnection(addr string, nodeType msg.NodeType, nodeID string, conn net.Conn) *connection {
	return &connection{
		id:       fmt.Sprintf("%s_%d", addr, time.Now().UnixNano()),
		addr:     addr,
		nodeType: nodeType,
		nodeID:   nodeID,
		conn:     conn,
		lastUsed: time.Now(),
	}
}

func (c *connection) ID() string {
	return c.id
}

func (c *connection) Addr() string {
	return c.addr
}

func (c *connection) NodeType() msg.NodeType {
	return c.nodeType
}

func (c *connection) NodeID() string {
	return c.nodeID
}

func (c *connection) RawConn() net.Conn {
	return c.conn
}

func (c *connection) Send(msgID uint32, nodeType msg.NodeType, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return fmt.Errorf("connection is closed")
	}

	msgLen := uint32(len(data))
	packet := make([]byte, 12+msgLen)
	binary.LittleEndian.PutUint32(packet[:4], msgLen)
	binary.LittleEndian.PutUint32(packet[4:8], uint32(nodeType))
	binary.LittleEndian.PutUint32(packet[8:12], msgID)
	copy(packet[12:], data)

	_, err := c.conn.Write(packet)
	if err != nil {
		logger.Error("Failed to send message: ", err)
		return err
	}
	return nil
}

func (c *connection) Receive(timeout time.Duration) (uint32, msg.NodeType, []byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return 0, 0, nil, fmt.Errorf("connection is closed")
	}

	if timeout > 0 {
		c.conn.SetReadDeadline(time.Now().Add(timeout))
		defer c.conn.SetReadDeadline(time.Time{})
	}

	buf := make([]byte, 4096)
	n, err := c.conn.Read(buf)
	if err != nil {
		return 0, 0, nil, err
	}

	data := buf[:n]
	if len(data) < 12 {
		return 0, 0, nil, fmt.Errorf("invalid message length")
	}

	msgLen := binary.LittleEndian.Uint32(data[:4])
	nodeType := msg.NodeType(binary.LittleEndian.Uint32(data[4:8]))
	if len(data) < int(msgLen)+12 {
		return 0, 0, nil, fmt.Errorf("incomplete message")
	}

	msgID := binary.LittleEndian.Uint32(data[8:12])
	msgData := data[12 : 12+msgLen]

	return msgID, nodeType, msgData, nil
}

func (c *connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}

func (c *connection) IsValid() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn != nil
}

func (c *connection) LastUsed() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastUsed
}

func (c *connection) UpdateLastUsed() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastUsed = time.Now()
}

type ConnPool struct {
	connections   map[string][]*connection
	mu            sync.RWMutex
	maxPerAddr    int
	cluster       *cluster.Cluster
	cleanupTicker *time.Ticker
}

func NewConnPool(cluster *cluster.Cluster, maxPerAddr int) *ConnPool {
	p := &ConnPool{
		connections: make(map[string][]*connection),
		maxPerAddr:  maxPerAddr,
		cluster:     cluster,
	}
	go p.startCleanupLoop()
	return p
}

func (p *ConnPool) startCleanupLoop() {
	p.cleanupTicker = time.NewTicker(30 * time.Second)
	defer p.cleanupTicker.Stop()

	for range p.cleanupTicker.C {
		p.cleanupStaleConnections()
	}
}

func (p *ConnPool) cleanupStaleConnections() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	staleTimeout := 5 * time.Minute

	for addr, conns := range p.connections {
		validConns := make([]*connection, 0, len(conns))
		for _, conn := range conns {
			if now.Sub(conn.LastUsed()) < staleTimeout && conn.IsValid() {
				validConns = append(validConns, conn)
			} else {
				conn.Close()
				logger.Debug("Cleaned up stale connection to ", addr)
			}
		}
		if len(validConns) == 0 {
			delete(p.connections, addr)
		} else {
			p.connections[addr] = validConns
		}
	}
}

func (p *ConnPool) getConnection(addr string, nodeType msg.NodeType, nodeID string) (Connection, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if conns, ok := p.connections[addr]; ok {
		for _, conn := range conns {
			if conn.IsValid() {
				conn.UpdateLastUsed()
				return conn, nil
			}
		}
		p.connections[addr] = nil
	}

	if len(p.connections[addr]) >= p.maxPerAddr {
		return nil, fmt.Errorf("max connections per address reached: %d", p.maxPerAddr)
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	newConn := newConnection(addr, nodeType, nodeID, conn)
	p.connections[addr] = append(p.connections[addr], newConn)

	logger.Info("Established new connection to ", addr, " (", nodeType.String(), ")")
	return newConn, nil
}

type Connector struct {
	pool    *ConnPool
	cluster *cluster.Cluster
}

func NewConnector(cluster *cluster.Cluster) *Connector {
	return &Connector{
		pool:    NewConnPool(cluster, 10),
		cluster: cluster,
	}
}

func (c *Connector) GetConnectionByNodeType(nodeType msg.NodeType) (Connection, error) {
	serviceName := nodeType.ServiceName()
	if serviceName == "" {
		return nil, fmt.Errorf("unknown node type: %s", nodeType.String())
	}

	service, err := c.cluster.GetRandomService(serviceName)
	if err != nil {
		return nil, err
	}

	if service == nil {
		return nil, fmt.Errorf("no service found for node type: %s", serviceName)
	}

	return c.pool.getConnection(service.Addr, nodeType, service.Addr)
}

func (c *Connector) GetConnectionByNodeID(nodeID string) (Connection, error) {
	parts := strings.Split(nodeID, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid node ID format: %s", nodeID)
	}

	nodeTypeStr := parts[0]
	addr := parts[1]

	nodeType, err := parseNodeType(nodeTypeStr)
	if err != nil {
		return nil, err
	}

	return c.pool.getConnection(addr, nodeType, nodeID)
}

func (c *Connector) SendToNodeType(nodeType msg.NodeType, msgID uint32, data []byte) error {
	conn, err := c.GetConnectionByNodeType(nodeType)
	if err != nil {
		return err
	}
	return conn.Send(msgID, nodeType, data)
}

func (c *Connector) RequestToNodeType(nodeType msg.NodeType, msgID uint32, data []byte, timeout time.Duration) (uint32, msg.NodeType, []byte, error) {
	conn, err := c.GetConnectionByNodeType(nodeType)
	if err != nil {
		return 0, 0, nil, err
	}

	err = conn.Send(msgID, nodeType, data)
	if err != nil {
		return 0, 0, nil, err
	}

	return conn.Receive(timeout)
}

func (c *Connector) SendToNodeID(nodeID string, msgID uint32, nodeType msg.NodeType, data []byte) error {
	conn, err := c.GetConnectionByNodeID(nodeID)
	if err != nil {
		return err
	}
	return conn.Send(msgID, nodeType, data)
}

func (c *Connector) Close() {
	c.pool.mu.Lock()
	defer c.pool.mu.Unlock()

	for _, conns := range c.pool.connections {
		for _, conn := range conns {
			conn.Close()
		}
	}
	c.pool.connections = make(map[string][]*connection)
}

func parseNodeType(typeStr string) (msg.NodeType, error) {
	switch strings.ToLower(typeStr) {
	case "gate":
		return msg.NodeTypeGate, nil
	case "login":
		return msg.NodeTypeLogin, nil
	case "logic":
		return msg.NodeTypeLogic, nil
	case "battle":
		return msg.NodeTypeBattle, nil
	case "gridmap":
		return msg.NodeTypeGridMap, nil
	case "cross":
		return msg.NodeTypeCross, nil
	case "data":
		return msg.NodeTypeData, nil
	case "gm":
		return msg.NodeTypeGM, nil
	default:
		return 0, fmt.Errorf("unknown node type: %s", typeStr)
	}
}
