package servicemap

import (
	"time"

	"github.com/traffic-refinery/traffic-refinery/internal/cache"
)

// IPCache contains the saved entries extracted from DNS queries and IP prefix
// matches
type IPCache struct {
	IPCacheMap *cache.SimpleTimeCache
}

func NewIPCache(cleanupTime, evictTime time.Duration) (*IPCache, error) {
	dc := &IPCache{}
	dc.IPCacheMap = cache.NewSimpleTimeCache(cleanupTime, evictTime)
	return dc, nil
}

// Insert new entry in the IPCache
func (dc *IPCache) Insert(ip string, services []ServiceID, ttl int64) {
	dc.IPCacheMap.Insert(ip, services, ttl)
}

// Lookup allows to lookup entries in the cache map
func (dc *IPCache) Lookup(ip string) ([]ServiceID, bool) {
	if entry, ok := dc.IPCacheMap.Lookup(ip); ok {
		return entry.([]ServiceID), true
	}
	return nil, false
}
