package cache

import (
	"hash/fnv"
	"sort"
	"sync"
)

const (
	VirtualNodes  = 10
	BatchSize     = 10
	WorkerCount   = 8
	WriteQueueCap = 1000
)

type CacheManager struct {
	shards       []*Cache
	hashRing     []uint32
	nodeMap      map[uint32]*Cache
	writeThrough sync.Map
	writeQueue   chan writeTask
	wg           sync.WaitGroup
}

type writeTask struct {
	key   string
	value string
	shard *Cache
}

func NewCacheManager(numShards int) *CacheManager {

	manager := &CacheManager{
		shards:     make([]*Cache, numShards),
		nodeMap:    make(map[uint32]*Cache),
		writeQueue: make(chan writeTask, WriteQueueCap),
	}

	for i := range numShards {
		manager.shards[i] = NewCache()
		for v := range VirtualNodes {
			virtualKey := string(rune(i)) + "#" + string(rune(v))
			hash := hashKey(virtualKey)
			manager.hashRing = append(manager.hashRing, hash)
			manager.nodeMap[hash] = manager.shards[i]
		}
	}

	sort.Slice(manager.hashRing, func(i, j int) bool {
		return manager.hashRing[i] < manager.hashRing[j]
	})

	manager.startWorkers()

	return manager
}

func (cm *CacheManager) startWorkers() {
	for i := 0; i < WorkerCount; i++ {
		cm.wg.Add(1)
		go cm.worker()
	}
}

func (cm *CacheManager) worker() {
	defer cm.wg.Done()

	batch := make([]writeTask, 0, BatchSize)

	for task := range cm.writeQueue {
		batch = append(batch, task)

		if len(batch) >= BatchSize {
			cm.processBatch(batch)
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		cm.processBatch(batch)
	}
}

func (cm *CacheManager) processBatch(batch []writeTask) {
	for _, task := range batch {
		task.shard.Put(task.key, task.value)
	}
}

func (cm *CacheManager) Put(key, value string) {
	shard := cm.getShard(key)

	cm.writeQueue <- writeTask{key, value, shard}
	cm.writeThrough.Store(key, value) 
}

func (cm *CacheManager) Get(key string) (string, bool) {
	if val, ok := cm.writeThrough.Load(key); ok {
		return val.(string), true
	}

	shard := cm.getShard(key)
	return shard.Get(key)
}

func (cm *CacheManager) getShard(key string) *Cache {
	hash := hashKey(key)
	idx := sort.Search(len(cm.hashRing), func(i int) bool {
		return cm.hashRing[i] >= hash
	})

	if idx == len(cm.hashRing) {
		idx = 0
	}

	return cm.nodeMap[cm.hashRing[idx]]
}


func hashKey(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

func (cm *CacheManager) Shutdown() {
	close(cm.writeQueue)
	cm.wg.Wait()
}
