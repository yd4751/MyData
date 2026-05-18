package pool

import (
	"sync"
)

type ObjectPool struct {
	pool sync.Pool
	new  func() interface{}
}

func NewObjectPool(newFunc func() interface{}) *ObjectPool {
	return &ObjectPool{
		pool: sync.Pool{
			New: newFunc,
		},
		new: newFunc,
	}
}

func (p *ObjectPool) Get() interface{} {
	return p.pool.Get()
}

func (p *ObjectPool) Put(obj interface{}) {
	p.pool.Put(obj)
}

type BufferPool struct {
	pools []*sync.Pool
}

func NewBufferPool(maxSize int) *BufferPool {
	p := &BufferPool{
		pools: make([]*sync.Pool, 0, 20),
	}
	for i := 0; i <= 20; i++ {
		size := 1 << i
		if size > maxSize {
			break
		}
		p.pools = append(p.pools, &sync.Pool{
			New: func() interface{} {
				return make([]byte, size)
			},
		})
	}
	return p
}

func (p *BufferPool) Get(size int) []byte {
	for i, pool := range p.pools {
		if 1<<i >= size {
			return pool.Get().([]byte)[:size]
		}
	}
	return make([]byte, size)
}

func (p *BufferPool) Put(buf []byte) {
	for i, pool := range p.pools {
		if 1<<i == cap(buf) {
			pool.Put(buf[:cap(buf)])
			return
		}
	}
}

type CoroutinePool struct {
	sem chan struct{}
	wg  sync.WaitGroup
}

func NewCoroutinePool(maxWorkers int) *CoroutinePool {
	return &CoroutinePool{
		sem: make(chan struct{}, maxWorkers),
	}
}

func (p *CoroutinePool) Submit(task func()) {
	p.wg.Add(1)
	go func() {
		p.sem <- struct{}{}
		defer func() {
			<-p.sem
			p.wg.Done()
		}()
		task()
	}()
}

func (p *CoroutinePool) Wait() {
	p.wg.Wait()
}
