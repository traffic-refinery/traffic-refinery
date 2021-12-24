package cache

import (
	"sync"
	"time"
)

type TimeItem struct {
	Object     interface{}
	Expiration int64
	LastUsed   int64
}

type SimpleTimeCache struct {
	items       map[string]TimeItem
	cleanupTime time.Duration
	evictTime   time.Duration
	stop        chan bool
	sync.RWMutex
}

func NewSimpleTimeCache(cleanupTime, evictTime time.Duration) *SimpleTimeCache {
	sc := &SimpleTimeCache{}
	sc.items = make(map[string]TimeItem)
	sc.cleanupTime = cleanupTime
	sc.evictTime = evictTime

	if sc.cleanupTime > 0 {
		sc.runCacheTimer()
	}
	return sc
}

// Insert new entry in the IPCache
func (sc *SimpleTimeCache) Insert(key string, value interface{}, ttl int64) error {
	var expireTime int64
	if ttl == 0 {
		expireTime = 0
	} else {
		expireTime = time.Now().Unix() + ttl
	}
	item := TimeItem{
		Object:     value,
		Expiration: expireTime,
		LastUsed:   time.Now().Unix(),
	}
	sc.Lock()
	sc.items[key] = item
	sc.Unlock()
	return nil
}

// Lookup allows to lookup entries in the cache map
func (sc *SimpleTimeCache) Lookup(key string) (value interface{}, found bool) {
	now := time.Now().Unix()
	sc.Lock()
	defer sc.Unlock()
	if entry, ok := sc.items[key]; ok {
		if entry.Expiration > 0 && entry.Expiration < now {
			delete(sc.items, key)
			return nil, false
		} else {
			entry.LastUsed = now
			return entry.Object, true
		}
	}
	return nil, false
}

// Removes unused DNS mappings form the local cache. It uses a default 600s (10m) expiry time
func (sc *SimpleTimeCache) ClearCache() {
	now := time.Now().Unix()
	sc.Lock()
	for i, d := range sc.items {
		if d.Expiration < now && d.LastUsed+int64(sc.evictTime/time.Second) < now { // delete only if expired AND the IP hasn't been seen in 10 min
			delete(sc.items, i)
		}
	}
	sc.Unlock()
}

func (sc *SimpleTimeCache) runCacheTimer() {
	go func() {
		ticker := time.NewTicker(time.Duration(sc.cleanupTime))
		for {
			select {
			case <-ticker.C:
				sc.ClearCache()
			case <-sc.stop:
				ticker.Stop()
				return
			}
		}
	}()
}

func (sc *SimpleTimeCache) StopCacheTimer() {
	sc.stop <- true
}
