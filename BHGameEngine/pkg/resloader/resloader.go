package resloader

import (
	"sync"
	"time"
)

type ResourceType int32

const (
	ResourceTypeModel     ResourceType = 1
	ResourceTypeTexture   ResourceType = 2
	ResourceTypeAudio     ResourceType = 3
	ResourceTypeCollision ResourceType = 4
	ResourceTypeNavMesh   ResourceType = 5
)

type ResourcePriority int32

const (
	PriorityHigh   ResourcePriority = 1
	PriorityMedium ResourcePriority = 2
	PriorityLow    ResourcePriority = 3
)

type ResourceStatus int32

const (
	StatusPending ResourceStatus = 1
	StatusLoading ResourceStatus = 2
	StatusLoaded  ResourceStatus = 3
	StatusFailed  ResourceStatus = 4
)

type ResourceInfo struct {
	ID         string
	Type       ResourceType
	Path       string
	Priority   ResourcePriority
	Status     ResourceStatus
	Size       int64
	LoadedSize int64
	LastUsed   time.Time
	RefCount   int32
	mu         sync.RWMutex
}

type ResourceLoader struct {
	resources      map[string]*ResourceInfo
	loadingQueue   []*ResourceInfo
	completedQueue []*ResourceInfo
	maxWorkers     int
	running        bool
	mu             sync.Mutex
}

func NewResourceLoader(maxWorkers int) *ResourceLoader {
	return &ResourceLoader{
		resources:  make(map[string]*ResourceInfo),
		maxWorkers: maxWorkers,
	}
}

func (r *ResourceLoader) AddResource(id string, resourceType ResourceType, path string, priority ResourcePriority) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.resources[id]; ok {
		return
	}

	resource := &ResourceInfo{
		ID:       id,
		Type:     resourceType,
		Path:     path,
		Priority: priority,
		Status:   StatusPending,
	}

	r.resources[id] = resource
	r.loadingQueue = append(r.loadingQueue, resource)
	r.sortQueue()
}

func (r *ResourceLoader) sortQueue() {
	for i := 0; i < len(r.loadingQueue)-1; i++ {
		for j := i + 1; j < len(r.loadingQueue); j++ {
			if r.loadingQueue[j].Priority < r.loadingQueue[i].Priority {
				r.loadingQueue[i], r.loadingQueue[j] = r.loadingQueue[j], r.loadingQueue[i]
			}
		}
	}
}

func (r *ResourceLoader) Start() {
	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return
	}
	r.running = true
	r.mu.Unlock()

	for i := 0; i < r.maxWorkers; i++ {
		go r.worker()
	}
}

func (r *ResourceLoader) worker() {
	for {
		r.mu.Lock()
		if !r.running || len(r.loadingQueue) == 0 {
			r.mu.Unlock()
			time.Sleep(10 * time.Millisecond)
			continue
		}

		resource := r.loadingQueue[0]
		r.loadingQueue = r.loadingQueue[1:]
		r.mu.Unlock()

		resource.mu.Lock()
		resource.Status = StatusLoading
		resource.mu.Unlock()

		time.Sleep(time.Millisecond * 100)

		resource.mu.Lock()
		resource.Status = StatusLoaded
		resource.LastUsed = time.Now()
		resource.mu.Unlock()

		r.mu.Lock()
		r.completedQueue = append(r.completedQueue, resource)
		r.mu.Unlock()
	}
}

func (r *ResourceLoader) Stop() {
	r.mu.Lock()
	r.running = false
	r.mu.Unlock()
}

func (r *ResourceLoader) GetResourceStatus(id string) ResourceStatus {
	r.mu.Lock()
	resource, ok := r.resources[id]
	r.mu.Unlock()

	if !ok {
		return StatusFailed
	}

	resource.mu.RLock()
	defer resource.mu.RUnlock()
	return resource.Status
}

func (r *ResourceLoader) GetLoadingProgress(id string) float64 {
	r.mu.Lock()
	resource, ok := r.resources[id]
	r.mu.Unlock()

	if !ok {
		return 0
	}

	resource.mu.RLock()
	defer resource.mu.RUnlock()

	if resource.Size == 0 {
		return 0
	}
	return float64(resource.LoadedSize) / float64(resource.Size)
}

func (r *ResourceLoader) ReleaseResource(id string) {
	r.mu.Lock()
	resource, ok := r.resources[id]
	r.mu.Unlock()

	if !ok {
		return
	}

	resource.mu.Lock()
	resource.RefCount--
	if resource.RefCount <= 0 {
		r.mu.Lock()
		delete(r.resources, id)
		r.mu.Unlock()
	}
	resource.mu.Unlock()
}

func (r *ResourceLoader) AcquireResource(id string) (*ResourceInfo, bool) {
	r.mu.Lock()
	resource, ok := r.resources[id]
	r.mu.Unlock()

	if !ok {
		return nil, false
	}

	resource.mu.Lock()
	resource.RefCount++
	resource.LastUsed = time.Now()
	resource.mu.Unlock()

	return resource, true
}

func (r *ResourceLoader) GetPendingCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.loadingQueue)
}

func (r *ResourceLoader) GetCompletedCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.completedQueue)
}

func (r *ResourceLoader) PreloadResources(resources []ResourceInfo) {
	for _, res := range resources {
		r.AddResource(res.ID, res.Type, res.Path, res.Priority)
	}
}

func (r *ResourceLoader) ClearCache() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, resource := range r.resources {
		resource.mu.Lock()
		if resource.RefCount == 0 {
			delete(r.resources, id)
		}
		resource.mu.Unlock()
	}
}
