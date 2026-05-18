package worldmap

import (
	"fmt"
	"math"
	"sync"

	"github.com/openworld-server/internal/cluster"
)

const (
	GridMapWidth  = 10000 // 地图宽度
	GridMapHeight = 10000 // 地图高度
)

// GridMapRouter 网格地图路由
type GridMapRouter struct {
	cluster       *cluster.Cluster // 集群管理器
	gridMapCount  int              // 网格数量
	gridMapWidth  float64          // 每个网格宽度
	gridMapHeight float64          // 每个网格高度
	mu            sync.RWMutex     // 并发锁
}

func NewGridMapRouter(cluster *cluster.Cluster, gridCount int) *GridMapRouter {
	return &GridMapRouter{
		cluster:       cluster,
		gridMapCount:  gridCount,
		gridMapWidth:  GridMapWidth / float64(gridCount),
		gridMapHeight: GridMapHeight / float64(gridCount),
	}
}

func (r *GridMapRouter) GetGridMapID(pos Vec3) int {
	if pos.X < 0 || pos.X >= GridMapWidth || pos.Y < 0 || pos.Y >= GridMapHeight {
		return 0
	}

	gridX := int(math.Floor(pos.X / r.gridMapWidth))
	gridY := int(math.Floor(pos.Y / r.gridMapHeight))
	return gridY*int(math.Sqrt(float64(r.gridMapCount))) + gridX + 1
}

func (r *GridMapRouter) GetGridMapName(pos Vec3) string {
	gridID := r.GetGridMapID(pos)
	return fmt.Sprintf("gridmap-%03d", gridID)
}

func (r *GridMapRouter) GetGridMapService(pos Vec3) (*cluster.ServiceInfo, error) {
	serviceName := r.GetGridMapName(pos)
	return r.cluster.GetRandomService(serviceName)
}

func (r *GridMapRouter) GetGridMapByID(gridID int) (*cluster.ServiceInfo, error) {
	serviceName := fmt.Sprintf("gridmap-%03d", gridID)
	return r.cluster.GetRandomService(serviceName)
}

func (r *GridMapRouter) GetGridMapBounds(gridID int) (Vec3, Vec3) {
	gridCount := int(math.Sqrt(float64(r.gridMapCount)))
	gridX := (gridID - 1) % gridCount
	gridY := (gridID - 1) / gridCount

	min := Vec3{
		X: float64(gridX) * r.gridMapWidth,
		Y: float64(gridY) * r.gridMapHeight,
		Z: 0,
	}
	max := Vec3{
		X: float64(gridX+1) * r.gridMapWidth,
		Y: float64(gridY+1) * r.gridMapHeight,
		Z: 0,
	}

	return min, max
}

func (r *GridMapRouter) IsCrossingBoundary(oldPos, newPos Vec3) bool {
	return r.GetGridMapID(oldPos) != r.GetGridMapID(newPos)
}

func (r *GridMapRouter) GetAdjacentGridMaps(pos Vec3) []int {
	currentGrid := r.GetGridMapID(pos)
	if currentGrid == 0 {
		return nil
	}

	gridCount := int(math.Sqrt(float64(r.gridMapCount)))
	gridX := (currentGrid - 1) % gridCount
	gridY := (currentGrid - 1) / gridCount

	adjacent := make([]int, 0)
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			if dx == 0 && dy == 0 {
				continue
			}
			newX := gridX + dx
			newY := gridY + dy
			if newX >= 0 && newX < gridCount && newY >= 0 && newY < gridCount {
				adjacent = append(adjacent, newY*gridCount+newX+1)
			}
		}
	}

	return adjacent
}

func (r *GridMapRouter) GetAllGridMapServices() ([]*cluster.ServiceInfo, error) {
	var allServices []*cluster.ServiceInfo
	gridCount := int(math.Sqrt(float64(r.gridMapCount)))

	for i := 0; i < gridCount; i++ {
		for j := 0; j < gridCount; j++ {
			gridID := i*gridCount + j + 1
			serviceName := fmt.Sprintf("gridmap-%03d", gridID)
			services, err := r.cluster.DiscoverServices(serviceName)
			if err != nil {
				continue
			}
			allServices = append(allServices, services...)
		}
	}

	return allServices, nil
}
