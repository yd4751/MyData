package log

import (
	"time"

	"github.com/openworld-server/internal/cluster"
	"github.com/openworld-server/pkg/logger"
)

type Logger struct {
	serviceName      string           // 服务名称
	cluster          *cluster.Cluster // 集群管理器
	logServerRunning bool             // 日志服务器发现是否运行中
}

func NewLogger(serviceName string) *Logger {
	return &Logger{
		serviceName: serviceName,
	}
}

func (l *Logger) Init(logPath string, level string) {
	logger.Init(logPath, level, l.serviceName)
}

func (l *Logger) SetCluster(cluster *cluster.Cluster) {
	l.cluster = cluster
}

func (l *Logger) StartLogServerDiscovery() {
	if l.cluster == nil {
		logger.Warn("Cluster not set, skipping log server discovery")
		return
	}

	if l.logServerRunning {
		logger.Warn("Log server discovery already running")
		return
	}

	l.logServerRunning = true
	go l.logServerDiscoveryLoop()
}

func (l *Logger) logServerDiscoveryLoop() {
	for {
		webService, err := l.cluster.GetRandomService("webserver")
		if err == nil && webService != nil {
			currentAddr := "http://" + webService.Addr
			logger.SetLogServer(currentAddr)
			logger.Info("Log server set to: ", webService.Addr)

			for {
				time.Sleep(5 * time.Second)
				webService, err = l.cluster.GetRandomService("webserver")
				if err != nil || webService == nil {
					logger.Warn("Log server disconnected, restarting discovery...")
					break
				}
				if "http://"+webService.Addr != currentAddr {
					currentAddr = "http://" + webService.Addr
					logger.SetLogServer(currentAddr)
					logger.Info("Log server changed to: ", webService.Addr)
				}
			}
		} else {
			logger.Warn("Failed to discover webserver for log forwarding, retrying in 5s...")
			time.Sleep(5 * time.Second)
		}
	}
}

func (l *Logger) Stop() {
	l.logServerRunning = false
}

func Trace(args ...interface{}) {
	logger.Trace(args...)
}

func Debug(args ...interface{}) {
	logger.Debug(args...)
}

func Info(args ...interface{}) {
	logger.Info(args...)
}

func Warn(args ...interface{}) {
	logger.Warn(args...)
}

func Error(args ...interface{}) {
	logger.Error(args...)
}

func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

func Tracef(format string, args ...interface{}) {
	logger.Tracef(format, args...)
}

func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
}

func SetLogServer(addr string) {
	logger.SetLogServer(addr)
}
