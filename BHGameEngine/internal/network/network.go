package network

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"io"
	"net"

	"github.com/gogo/protobuf/proto"
	"github.com/openworld-server/internal/msg"
	"github.com/openworld-server/pkg/logger"
)

type Message struct {
	ID       uint32       // 消息ID
	NodeType msg.NodeType // 目标节点类型
	Data     []byte       // 消息数据
	Session  net.Conn     // 客户端连接会话
}

type Handler interface {
	Handle(msg *Message)
}

type Server struct {
	listener net.Listener // TCP监听器
	handler  Handler      // 消息处理器
	running  bool         // 服务器运行状态
}

func NewServer(handler Handler) *Server {
	return &Server{
		handler: handler,
	}
}

func (s *Server) OnOpened(conn net.Conn) {
	logger.Info("New connection from: ", conn.RemoteAddr().String())
}

func (s *Server) OnClosed(conn net.Conn, err error) {
	if err != nil {
		logger.Info("Connection closed with error: ", conn.RemoteAddr().String(), " - ", err.Error())
	} else {
		logger.Info("Connection closed: ", conn.RemoteAddr().String())
	}
}

func (s *Server) Start(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = listener
	s.running = true

	logger.Info("Network server started on ", addr)
	go s.acceptLoop()
	return nil
}

func (s *Server) acceptLoop() {
	for s.running {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.running {
				logger.Error("Accept error: ", err)
			}
			return
		}
		s.OnOpened(conn)
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		s.OnClosed(conn, nil)
		conn.Close()
	}()

	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			s.OnClosed(conn, err)
			return
		}

		data := buf[:n]
		for len(data) >= 12 {
			msgLen := binary.LittleEndian.Uint32(data[:4])
			nodeType := msg.NodeType(binary.LittleEndian.Uint32(data[4:8]))
			if len(data) < int(msgLen)+12 {
				break
			}

			msgID := binary.LittleEndian.Uint32(data[8:12])
			msgData := data[12 : 12+msgLen]

			msgObj := &Message{
				ID:       msgID,
				NodeType: nodeType,
				Data:     msgData,
				Session:  conn,
			}

			logger.Debug("Received message from ", conn.RemoteAddr(), " - MsgID:", msgID, "(", msg.GetMsgName(msgID), "), NodeType:", nodeType, "(", nodeType.String(), "), Length:", msgLen)
			s.handler.Handle(msgObj)
			data = data[12+msgLen:]
		}
	}
}

func (s *Server) Stop() {
	s.running = false
	if s.listener != nil {
		s.listener.Close()
	}
	logger.Info("Network server stopped")
}

func SendMessage(conn net.Conn, msgID uint32, nodeType msg.NodeType, data proto.Message) error {
	buf, err := proto.Marshal(data)
	if err != nil {
		logger.Error("Failed to marshal proto message: ", err)
		return err
	}

	msgLen := uint32(len(buf))
	packet := make([]byte, 12+msgLen)
	binary.LittleEndian.PutUint32(packet[:4], msgLen)
	binary.LittleEndian.PutUint32(packet[4:8], uint32(nodeType))
	binary.LittleEndian.PutUint32(packet[8:12], msgID)
	copy(packet[12:], buf)

	logger.Debug("Sending message to ", conn.RemoteAddr(), " - MsgID:", msgID, "(", msg.GetMsgName(msgID), "), NodeType:", nodeType, "(", nodeType.String(), "), Length:", msgLen)
	_, err = conn.Write(packet)
	if err != nil {
		logger.Error("Failed to send message: ", err)
	}
	return err
}

func SendRawMessage(conn net.Conn, msgID uint32, nodeType msg.NodeType, data []byte) error {
	msgLen := uint32(len(data))
	packet := make([]byte, 12+msgLen)
	binary.LittleEndian.PutUint32(packet[:4], msgLen)
	binary.LittleEndian.PutUint32(packet[4:8], uint32(nodeType))
	binary.LittleEndian.PutUint32(packet[8:12], msgID)
	copy(packet[12:], data)

	logger.Debug("Sending raw message to ", conn.RemoteAddr(), " - MsgID:", msgID, "(", msg.GetMsgName(msgID), "), NodeType:", nodeType, "(", nodeType.String(), "), Length:", msgLen)
	_, err := conn.Write(packet)
	if err != nil {
		logger.Error("Failed to send raw message: ", err)
	}
	return err
}

func ParseMessage(data []byte) (uint32, msg.NodeType, []byte, error) {
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

func ComputeHash(s string) uint64 {
	h := fnv.New64()
	h.Write([]byte(s))
	return h.Sum64()
}

func GetIP(conn net.Conn) string {
	addr := conn.RemoteAddr().String()
	host, _, _ := net.SplitHostPort(addr)
	return host
}

func IsValidPacket(data []byte) bool {
	if len(data) < 8 {
		return false
	}
	msgLen := binary.LittleEndian.Uint32(data[:4])
	if msgLen > 1024*1024 {
		return false
	}
	return true
}

func ReadFull(conn io.Reader, buf []byte) (int, error) {
	n := 0
	for n < len(buf) {
		nn, err := conn.Read(buf[n:])
		if err != nil {
			return n + nn, err
		}
		n += nn
	}
	return n, nil
}
