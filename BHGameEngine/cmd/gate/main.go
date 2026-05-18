package main

import (
	"flag"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/openworld-server/internal/cluster"
	"github.com/openworld-server/internal/log"
	"github.com/openworld-server/internal/msg"
	"github.com/openworld-server/internal/network"
	"github.com/openworld-server/pkg/config"
)

var configPath = flag.String("config", "./config/config.toml", "config file path")

type GateHandler struct {
	cluster *cluster.Cluster
}

func (h *GateHandler) Handle(msgObj *network.Message) {
	log.Info("Gate received message from ", msgObj.Session.RemoteAddr(), " - MsgID:", msgObj.ID, "(", msg.GetMsgName(msgObj.ID), "), NodeType:", msgObj.NodeType, "(", msgObj.NodeType.String(), ")")

	if msgObj.NodeType == msg.NodeTypeGate {
		h.handleGateMessage(msgObj)
	} else {
		h.forwardMessage(msgObj)
	}
}

func (h *GateHandler) handleGateMessage(msgObj *network.Message) {
	log.Info("Handling gate message, MsgID:", msgObj.ID, "(", msg.GetMsgName(msgObj.ID), ")")
}

func (h *GateHandler) forwardMessage(msgObj *network.Message) {
	serviceName := h.getServiceNameByNodeType(msgObj.NodeType)
	if serviceName == "" {
		log.Error("Unknown node type:", msgObj.NodeType, "(", msgObj.NodeType.String(), ")")
		return
	}

	log.Info("Forwarding message to ", serviceName, " - MsgID:", msgObj.ID, "(", msg.GetMsgName(msgObj.ID), ")")
	service, err := h.cluster.GetRandomService(serviceName)
	if err != nil {
		log.Error("Failed to get service ", serviceName, ": ", err)
		return
	}

	log.Info("Connecting to ", serviceName, " at ", service.Addr)
	conn, err := net.Dial("tcp", service.Addr)
	if err != nil {
		log.Error("Failed to connect to ", serviceName, " at ", service.Addr, ": ", err)
		return
	}
	defer conn.Close()

	err = network.SendRawMessage(conn, msgObj.ID, msgObj.NodeType, msgObj.Data)
	if err != nil {
		log.Error("Failed to send message to ", serviceName, ": ", err)
		return
	}

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		log.Error("Failed to read response from ", serviceName, ": ", err)
		return
	}

	respMsgID, respNodeType, respData, err := network.ParseMessage(buf[:n])
	if err != nil {
		log.Error("Failed to parse response: ", err)
		return
	}

	log.Info("Received response from ", serviceName, " - MsgID:", respMsgID, "(", msg.GetMsgName(respMsgID), "), NodeType:", respNodeType, "(", respNodeType.String(), ")")
	err = network.SendRawMessage(msgObj.Session, respMsgID, respNodeType, respData)
	if err != nil {
		log.Error("Failed to send response to client: ", err)
	}
}

func (h *GateHandler) getServiceNameByNodeType(nodeType msg.NodeType) string {
	switch nodeType {
	case msg.NodeTypeLogin:
		return "login"
	case msg.NodeTypeLogic:
		return "logic"
	case msg.NodeTypeBattle:
		return "battle"
	case msg.NodeTypeGridMap:
		return "gridmap"
	case msg.NodeTypeCross:
		return "cross"
	case msg.NodeTypeData:
		return "dataservice"
	case msg.NodeTypeGM:
		return "gm"
	default:
		return ""
	}
}

func main() {
	flag.Parse()

	log.Info("Gate server initializing...")

	err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	logger := log.NewLogger("gate")
	loggerConfig := config.GetLoggerConfig()
	logger.Init(loggerConfig.Path, loggerConfig.Level)
	log.Info("Gate server using log config: Path=", loggerConfig.Path, ", Level=", loggerConfig.Level)

	etcdAddr := config.GetEtcdAddr()
	log.Info("Connecting to etcd at ", etcdAddr)
	cluster, err := cluster.NewCluster(etcdAddr)
	if err != nil {
		log.Fatal("Failed to connect to etcd:", err)
	}

	logger.SetCluster(cluster)
	go logger.StartLogServerDiscovery()

	listenAddr := config.GetGateListenAddr()
	log.Info("Registering gate service at ", listenAddr)
	err = cluster.RegisterService("gate", listenAddr, map[string]string{
		"type": "gateway",
	})
	if err != nil {
		log.Fatal("Failed to register service:", err)
	}

	log.Info("Creating GateHandler and starting network server")
	handler := &GateHandler{cluster: cluster}
	server := network.NewServer(handler)

	go func() {
		if err := server.Start(listenAddr); err != nil {
			log.Fatal("Server failed to start:", err)
		}
	}()

	log.Info("Gate server started successfully on ", listenAddr)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("Shutting down gate server...")
	server.Stop()
}
