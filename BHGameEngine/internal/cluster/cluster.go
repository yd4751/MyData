package cluster

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/openworld-server/internal/node"
	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

// ServiceInfo 服务信息
type ServiceInfo struct {
	Name        string            // 服务名称
	Addr        string            // 服务地址
	CPUUsage    float64           // CPU使用率(%)
	MemoryUsage float64           // 内存使用率(%)
	Connections int               // 连接数
	ProcessID   int               // 进程ID
	StartTime   time.Time         // 启动时间
	Metadata    map[string]string // 元数据
}

// Cluster 集群管理器
type Cluster struct {
	etcd     *clientv3.Client          // etcd客户端
	services map[string][]*ServiceInfo // 服务列表
}

func NewCluster(etcdAddr string) (*Cluster, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(etcdAddr, ","),
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	return &Cluster{
		etcd:     client,
		services: make(map[string][]*ServiceInfo),
	}, nil
}

func (c *Cluster) RegisterService(name, addr string, metadata map[string]string) error {
	key := fmt.Sprintf("/services/%s/%s", name, addr)
	value := fmt.Sprintf("%s|%s", addr, buildMetadata(metadata))

	lease, err := c.etcd.Grant(context.Background(), 30)
	if err != nil {
		return err
	}

	_, err = c.etcd.Put(context.Background(), key, value, clientv3.WithLease(lease.ID))
	if err != nil {
		return err
	}

	go func() {
		for {
			time.Sleep(20 * time.Second)
			c.etcd.KeepAliveOnce(context.Background(), lease.ID)
		}
	}()

	logrus.Info("Service registered: ", name, " ", addr)
	return nil
}

func (c *Cluster) DiscoverServices(name string) ([]*ServiceInfo, error) {
	prefix := fmt.Sprintf("/services/%s/", name)
	resp, err := c.etcd.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	services := make([]*ServiceInfo, 0)
	for _, kv := range resp.Kvs {
		nodeInfo := node.ParseNodeInfo(name, string(kv.Value))
		if nodeInfo != nil {
			services = append(services, &ServiceInfo{
				Name:        nodeInfo.Name,
				Addr:        nodeInfo.Addr,
				CPUUsage:    nodeInfo.CPUUsage,
				MemoryUsage: nodeInfo.MemoryUsage,
				Connections: nodeInfo.Connections,
				ProcessID:   nodeInfo.ProcessID,
				StartTime:   nodeInfo.StartTime,
				Metadata:    nodeInfo.Metadata,
			})
		} else {
			parts := strings.Split(string(kv.Value), "|")
			if len(parts) >= 2 {
				services = append(services, &ServiceInfo{
					Name:     name,
					Addr:     parts[0],
					Metadata: parseMetadata(parts[1]),
				})
			}
		}
	}

	c.services[name] = services
	return services, nil
}

func (c *Cluster) GetRandomService(name string) (*ServiceInfo, error) {
	services, err := c.DiscoverServices(name)
	if err != nil {
		return nil, err
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("no service found: %s", name)
	}

	return c.getLeastLoadedService(services), nil
}

func (c *Cluster) getLeastLoadedService(services []*ServiceInfo) *ServiceInfo {
	if len(services) == 0 {
		return nil
	}

	leastLoaded := services[0]
	minLoad := calculateLoad(leastLoaded)

	for i := 1; i < len(services); i++ {
		load := calculateLoad(services[i])
		if load < minLoad {
			minLoad = load
			leastLoaded = services[i]
		}
	}

	return leastLoaded
}

func calculateLoad(service *ServiceInfo) float64 {
	cpuWeight := 0.4
	memWeight := 0.3
	connWeight := 0.3

	cpuLoad := service.CPUUsage / 100.0
	memLoad := service.MemoryUsage / 100.0

	maxConn := 1000.0
	connLoad := float64(service.Connections) / maxConn

	return cpuWeight*cpuLoad + memWeight*memLoad + connWeight*connLoad
}

func (c *Cluster) ConnectToService(addr string) (*grpc.ClientConn, error) {
	return grpc.Dial(addr, grpc.WithInsecure(), grpc.WithTimeout(5*time.Second))
}

func (c *Cluster) WatchServices(name string, callback func([]*ServiceInfo)) error {
	prefix := fmt.Sprintf("/services/%s/", name)
	watchChan := c.etcd.Watch(context.Background(), prefix, clientv3.WithPrefix())

	go func() {
		for watchResp := range watchChan {
			services := make([]*ServiceInfo, 0)
			for _, event := range watchResp.Events {
				if event.Type == clientv3.EventTypePut {
					nodeInfo := node.ParseNodeInfo(name, string(event.Kv.Value))
					if nodeInfo != nil {
						services = append(services, &ServiceInfo{
							Name:        nodeInfo.Name,
							Addr:        nodeInfo.Addr,
							CPUUsage:    nodeInfo.CPUUsage,
							MemoryUsage: nodeInfo.MemoryUsage,
							Connections: nodeInfo.Connections,
							ProcessID:   nodeInfo.ProcessID,
							StartTime:   nodeInfo.StartTime,
							Metadata:    nodeInfo.Metadata,
						})
					} else {
						parts := strings.Split(string(event.Kv.Value), "|")
						if len(parts) >= 2 {
							services = append(services, &ServiceInfo{
								Name:     name,
								Addr:     parts[0],
								Metadata: parseMetadata(parts[1]),
							})
						}
					}
				}
			}
			c.services[name] = services
			callback(services)
		}
	}()

	return nil
}

func buildMetadata(md map[string]string) string {
	var parts []string
	for k, v := range md {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, ",")
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
