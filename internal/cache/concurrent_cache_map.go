package cache

import (
	"encoding/json"
	"errors"
	"sync"
	"time"
)

// ConcurrentCacheMap has been heavily inspired by
// https://github.com/orcaman/concurrent-map and
// https://github.com/patrickmn/go-cache.
// ConcurrentCacheMap implements a concurrent map with variable number of shards
// and the possibility of adding a periodic garbage collection function to remove
// expired entries.

const (
	DEFAULT_SHARD_COUNT   uint32        = 32
	DEFAULT_EXPIRATION    time.Duration = 10 * time.Minute
	DEFAULT_CACHE_CLEANUP time.Duration = 5 * time.Minute
)

type Item struct {
	Object     interface{}
	Expiration time.Duration
}

// Returns true if the item has expired.
func (item Item) Expired() bool {
	if item.Expiration == 0 {
		return false
	}
	return time.Duration(time.Now().UnixNano()) > item.Expiration
}

// A "thread" safe map of type string:Anything.
// To avoid lock bottlenecks this map is dived to several (SHARD_COUNT) map shards.
type ConcurrentCacheMap struct {
	shards         []*Shard
	shardCount     uint32
	expirationTime time.Duration
	onEvicted      func(interface{})
	interval       time.Duration
	stop           chan bool
}

// A "thread" safe string to anything map.
type Shard struct {
	items map[string]Item
	// Read Write mutex, guards access to internal map.
	sync.RWMutex
}

// Creates a new concurrent map.
func NewConcurrentCacheMap(shardCount uint32, expiration time.Duration, onEvicted func(interface{}), interval time.Duration) *ConcurrentCacheMap {
	m := &ConcurrentCacheMap{}
	m.shardCount = shardCount
	m.expirationTime = expiration
	m.shards = make([]*Shard, shardCount)
	m.onEvicted = onEvicted
	m.interval = interval
	for i := uint32(0); i < shardCount; i++ {
		m.shards[i] = &Shard{items: make(map[string]Item)}
	}
	if m.interval > 0 {
		m.runCacheTimer()
	}
	return m
}

// Returns shard under given key
func (m *ConcurrentCacheMap) getShard(key string) *Shard {
	return m.shards[uint(fnv32(key))%uint(m.shardCount)]
}

// Sets the given value under the specified key.
func (m *ConcurrentCacheMap) Set(key string, value interface{}) error {
	// Get map shard.
	shard := m.getShard(key)
	shard.Lock()
	shard.items[key] = Item{Object: value, Expiration: time.Duration(time.Now().UnixNano()) + m.expirationTime}
	shard.Unlock()
	return nil
}

// Retrieves an element from map under given key.
func (m *ConcurrentCacheMap) Get(key string) (interface{}, bool) {
	// Get shard
	shard := m.getShard(key)
	shard.RLock()
	// Get item from shard.
	val, ok := shard.items[key]
	shard.RUnlock()
	if ok {
		return val.Object, ok
	} else {
		return nil, ok
	}

}

// Returns the number of elements within the map.
func (m *ConcurrentCacheMap) Count() int {
	count := 0
	for i := uint32(0); i < m.shardCount; i++ {
		shard := m.shards[i]
		shard.RLock()
		count += len(shard.items)
		shard.RUnlock()
	}
	return count
}

// Looks up an item under specified key
func (m *ConcurrentCacheMap) Has(key string) bool {
	// Get shard
	shard := m.getShard(key)
	shard.RLock()
	// See if element is within shard.
	_, ok := shard.items[key]
	shard.RUnlock()
	return ok
}

// Removes an element from the map.
func (m *ConcurrentCacheMap) Remove(key string) {
	// Try to get shard.
	shard := m.getShard(key)
	shard.Lock()
	delete(shard.items, key)
	shard.Unlock()
}

// Returns an item, but keeps its shard blocked. Useful when stored items are
// pointers.
// Risk of deadlock warning: if SetAndUnlock or Unlock are not called after this function
// the shard containing the item will never be unlocked
func (m *ConcurrentCacheMap) GetAndLock(key string) (interface{}, bool) {
	// Get shard
	shard := m.getShard(key)
	shard.Lock()
	// Get item from shard.
	val, ok := shard.items[key]

	if ok {
		return val.Object, ok
	} else {
		// In case the object is not found it unlocks
		shard.Unlock()
		return nil, ok
	}
}

// Inserts the element into the cache.
// If an element with the same key exists it replaces it.
// Finally it unlocks the mutex
// Risk of deadlock warning: if SetAndUnlock or Unlock are not called after this function
// the shard containing the item will never be unlocked
func (m *ConcurrentCacheMap) SetAndUnlock(key string, value interface{}) error {
	shard := m.getShard(key)
	shard.items[key] = Item{Object: value, Expiration: time.Duration(time.Now().UnixNano()) + m.expirationTime}
	shard.Unlock()
	return nil
}

// Unlocks the mutex associated with a given key
// Risk of deadlock warning: if SetAndUnlock or Unlock are not called after this function
// the shard containing the item will never be unlocked
func (m *ConcurrentCacheMap) Unlock(key string) error {
	shard := m.getShard(key)
	shard.Unlock()
	return nil
}

func (m *ConcurrentCacheMap) Clear() error {
	//First locks the entire cache (all shards)
	for _, shard := range m.shards {
		shard.Lock()
	}
	//Then reomve all elements
	for _, shard := range m.shards {
		for k := range shard.items {
			delete(shard.items, k)
		}
	}
	// Finally it unlocks the entire cache
	for _, shard := range m.shards {
		shard.Unlock()
	}
	return nil
}

func (m *ConcurrentCacheMap) PurgeExpired() error {
	m.DeleteExpired()
	return nil
}

// Removes an element from the map and returns it
func (m *ConcurrentCacheMap) Pop(key string) (interface{}, bool) {
	// Try to get shard.
	shard := m.getShard(key)
	shard.Lock()
	v, exists := shard.items[key]
	if exists {
		delete(shard.items, key)
		shard.Unlock()
		return v.Object, exists
	} else {
		shard.Unlock()
		return nil, exists
	}

}

// Checks if map is empty.
func (m *ConcurrentCacheMap) IsEmpty() bool {
	return m.Count() == 0
}

// Delete all expired items from the cache.
func (m *ConcurrentCacheMap) DeleteExpired() {
	var evictedItems []interface{}
	now := time.Duration(time.Now().UnixNano())
	for _, shard := range m.shards {
		shard.Lock()
		for k, v := range shard.items {
			// "Inlining" of expired
			if v.Expiration > 0 && now > v.Expiration {
				evicted := v.Object
				delete(shard.items, k)
				evictedItems = append(evictedItems, evicted)
			}
		}
		shard.Unlock()
	}
	if m.onEvicted != nil {
		for evicted := range evictedItems {
			m.onEvicted(evicted)
		}
	}
}

// Used by the Iter & IterBuffered functions to wrap two variables together over a channel,
type Tuple struct {
	Key string
	Val interface{}
}

// Returns an iterator which could be used in a for range loop.
//
// Deprecated: using IterBuffered() will get a better performence
func (m *ConcurrentCacheMap) Iter() <-chan Tuple {
	chans := snapshot(m)
	ch := make(chan Tuple)
	go fanIn(chans, ch)
	return ch
}

// Returns a buffered iterator which could be used in a for range loop.
func (m *ConcurrentCacheMap) IterBuffered() <-chan Tuple {
	chans := snapshot(m)
	total := 0
	for _, c := range chans {
		total += cap(c)
	}
	ch := make(chan Tuple, total)
	go fanIn(chans, ch)
	return ch
}

// Returns a array of channels that contains elements in each shard,
// which likely takes a snapshot of `m`.
// It returns once the size of each buffered channel is determined,
// before all the channels are populated using goroutines.
func snapshot(m *ConcurrentCacheMap) (chans []chan Tuple) {
	chans = make([]chan Tuple, m.shardCount)
	wg := sync.WaitGroup{}
	wg.Add(int(m.shardCount))
	// Foreach shard.
	for index, shard := range m.shards {
		go func(index int, shard *Shard) {
			// Foreach key, value pair.
			shard.RLock()
			chans[index] = make(chan Tuple, len(shard.items))
			wg.Done()
			for key, val := range shard.items {
				chans[index] <- Tuple{key, val}
			}
			shard.RUnlock()
			close(chans[index])
		}(index, shard)
	}
	wg.Wait()
	return chans
}

// fanIn reads elements from channels `chans` into channel `out`
func fanIn(chans []chan Tuple, out chan Tuple) {
	wg := sync.WaitGroup{}
	wg.Add(len(chans))
	for _, ch := range chans {
		go func(ch chan Tuple) {
			for t := range ch {
				out <- t
			}
			wg.Done()
		}(ch)
	}
	wg.Wait()
	close(out)
}

// Returns all items as map[string]interface{}
func (m *ConcurrentCacheMap) Items() map[string]interface{} {
	tmp := make(map[string]interface{})

	// Insert items to temporary map.
	for item := range m.IterBuffered() {
		tmp[item.Key] = item.Val
	}

	return tmp
}

func (m *ConcurrentCacheMap) Dump() map[string]interface{} {
	return m.Items()
}

// Return all keys as []string
func (m *ConcurrentCacheMap) Keys() []string {
	count := m.Count()
	ch := make(chan string, count)
	go func() {
		// Foreach shard.
		wg := sync.WaitGroup{}
		wg.Add(int(m.shardCount))
		for _, shard := range m.shards {
			go func(shard *Shard) {
				// Foreach key, value pair.
				shard.RLock()
				for key := range shard.items {
					ch <- key
				}
				shard.RUnlock()
				wg.Done()
			}(shard)
		}
		wg.Wait()
		close(ch)
	}()

	// Generate keys
	keys := make([]string, 0, count)
	for k := range ch {
		keys = append(keys, k)
	}
	return keys
}

//Reviles ConcurrentCacheMap "private" variables to json marshal.
func (m *ConcurrentCacheMap) MarshalJSON() ([]byte, error) {
	// Create a temporary map, which will hold all item spread across shards.
	tmp := make(map[string]interface{})

	// Insert items to temporary map.
	for item := range m.IterBuffered() {
		tmp[item.Key] = item.Val
	}
	return json.Marshal(tmp)
}

//
type ConcurrentCacheMapIterator struct {
	//
	init bool
	//
	done bool
	//
	m *ConcurrentCacheMap
	//
	currentShard int
	//
	keys []string
	//
	currentKey int
}

func (i *ConcurrentCacheMapIterator) getNextElement() (interface{}, bool) {
	// If the iteration was completed previously return
	if i.done {
		return nil, false
	}
	// Initialize if needed
	if !i.init {
		// Find first available shard with items
		i.currentShard = -1
		for j := 0; j < len(i.m.shards); j++ {
			i.m.shards[j].Lock()
			if len(i.m.shards[j].items) > 0 {
				i.currentShard = j
				break
			}
			i.m.shards[j].Unlock()
		}

		// No items in the map, it's done
		if i.currentShard == -1 {
			i.done = true
			return nil, false
		}

		// Create list of keys
		i.keys = make([]string, 0, len(i.m.shards[i.currentShard].items))
		for k := range i.m.shards[i.currentShard].items {
			i.keys = append(i.keys, k)
		}
		i.currentKey = 0
		i.init = true
	} else if i.currentKey >= len(i.keys) {
		// Find next shard with items otherwise return done
		i.m.shards[i.currentShard].Unlock()
		ns := i.currentShard + 1
		i.currentShard = -1
		for j := ns; j < len(i.m.shards); j++ {
			i.m.shards[j].Lock()
			if len(i.m.shards[j].items) > 0 {
				i.currentShard = j
				i.keys = make([]string, 0, len(i.m.shards[i.currentShard].items))
				for k := range i.m.shards[i.currentShard].items {
					i.keys = append(i.keys, k)
				}
				i.currentKey = 0
				break
			}
			i.m.shards[j].Unlock()
		}

		// No more items in the map
		if i.currentShard == -1 {
			i.done = true
			return nil, false
		}
	}

	// Return the key
	i.currentKey += 1
	return i.m.shards[i.currentShard].items[i.keys[i.currentKey-1]].Object, true

}

func (i *ConcurrentCacheMapIterator) clear() {
	if i.init && !i.done {
		if i.currentShard < len(i.m.shards) {
			i.m.shards[i.currentShard].Unlock()
		}
	}
	i.done = true
}

func (m *ConcurrentCacheMap) IterativeDump() (CacheIterator, error) {
	return &ConcurrentCacheMapIterator{
		init: false,
		done: false,
		m:    m,
	}, nil
}

func (m *ConcurrentCacheMap) NextElement(i CacheIterator) (interface{}, error) {
	if iter, ok := i.(*ConcurrentCacheMapIterator); !ok {
		return nil, errors.New("wrong iterator type")
	} else if v, ok := iter.getNextElement(); ok {
		return v, nil
	} else {
		return nil, nil
	}
}

func (m *ConcurrentCacheMap) StopIteration(i CacheIterator) error {
	if iter, ok := i.(*ConcurrentCacheMapIterator); !ok {
		return errors.New("wrong iterator type")
	} else {
		iter.clear()
		return nil
	}
}

func (m *ConcurrentCacheMap) runCacheTimer() {
	go func() {
		ticker := time.NewTicker(time.Duration(m.interval))
		for {
			select {
			case <-ticker.C:
				m.DeleteExpired()
			case <-m.stop:
				ticker.Stop()
				return
			}
		}
	}()
}

func (m *ConcurrentCacheMap) stopCacheTimer() {
	m.stop <- true
}
