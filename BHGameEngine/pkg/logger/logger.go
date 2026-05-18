package logger

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	Log           *logrus.Logger
	once          sync.Once
	logServerAddr string
)

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Service   string `json:"service"`
	Message   string `json:"message"`
}

func init() {
	once.Do(func() {
		Log = logrus.New()
		Log.SetLevel(logrus.InfoLevel)
		Log.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05.000",
			FullTimestamp:   true,
		})
		Log.SetOutput(os.Stdout)
	})
}

func SetLogServer(addr string) {
	logServerAddr = addr
}

func Init(logPath string, level string, serviceName string) {
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	Log.SetLevel(logLevel)

	customFormatter := &logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000",
		FullTimestamp:   true,
	}

	Log.SetFormatter(customFormatter)

	writer := &multiWriter{
		serviceName: serviceName,
	}

	if err := os.MkdirAll(logPath, 0755); err != nil {
		writer.stdout = os.Stdout
	} else {
		logFile := filepath.Join(logPath, serviceName+".log")
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			writer.stdout = os.Stdout
		} else {
			writer.file = file
		}
	}

	Log.SetOutput(writer)
}

type multiWriter struct {
	serviceName string
	stdout      *os.File
	file        *os.File
}

func (w *multiWriter) Write(p []byte) (n int, err error) {
	if w.stdout != nil {
		n, err = w.stdout.Write(p)
	} else if w.file != nil {
		n, err = w.file.Write(p)
	}

	if logServerAddr != "" {
		go func() {
			entry := &LogEntry{
				Timestamp: time.Now().Format("2006-01-02 15:04:05.000"),
				Level:     getLogLevel(p),
				Service:   w.serviceName,
				Message:   string(p),
			}
			sendToLogServer(entry)
		}()
	}

	return n, err
}

func getLogLevel(p []byte) string {
	if bytes.Contains(p, []byte("level=trace")) {
		return "trace"
	} else if bytes.Contains(p, []byte("level=debug")) {
		return "debug"
	} else if bytes.Contains(p, []byte("level=info")) {
		return "info"
	} else if bytes.Contains(p, []byte("level=warn")) {
		return "warn"
	} else if bytes.Contains(p, []byte("level=error")) {
		return "error"
	} else if bytes.Contains(p, []byte("level=fatal")) {
		return "fatal"
	}
	return "info"
}

func sendToLogServer(entry *LogEntry) {
	data, _ := json.Marshal(entry)
	resp, err := http.Post(logServerAddr+"/log", "application/json", bytes.NewBuffer(data))
	if err != nil {
		Log.Warn("Failed to send log to server:", err)
		return
	}
	resp.Body.Close()
}

func Trace(args ...interface{}) {
	Log.Trace(args...)
}

func Debug(args ...interface{}) {
	Log.Debug(args...)
}

func Info(args ...interface{}) {
	Log.Info(args...)
}

func Warn(args ...interface{}) {
	Log.Warn(args...)
}

func Error(args ...interface{}) {
	Log.Error(args...)
}

func Fatal(args ...interface{}) {
	Log.Fatal(args...)
}

func Tracef(format string, args ...interface{}) {
	Log.Tracef(format, args...)
}

func Debugf(format string, args ...interface{}) {
	Log.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	Log.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	Log.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	Log.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	Log.Fatalf(format, args...)
}
