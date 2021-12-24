package cache

// import (
// 	"encoding/json"
// 	"errors"
// 	"sync"
// 	"time"
//
// 	"github.com/allegro/bigcache"
// )

// //Wrapper structure around github.com/allegro/bigcache
// type BigCache struct {
// 	innerCache *bigcache.BigCache
// }
//
// //Initializes a new BigCache with default config
// func NewBigCache() (*BigCache, error) {
// 	config := bigcache.Config{
// 		// number of shards (must be a power of 2)
// 		Shards: 1024,
// 		// time after which entry can be evicted
// 		LifeWindow: 10 * time.Minute,
// 		// rps * lifeWindow, used only in initial memory allocation
// 		MaxEntriesInWindow: 1000 * 10 * 60,
// 		// max entry size in bytes, used only in initial memory allocation
// 		MaxEntrySize: 500,
// 		// prints information about additional memory allocation
// 		Verbose: false,
// 		// cache will not allocate more memory than this limit, value in MB
// 		// if value is reached then the oldest entries can be overridden for the new ones
// 		// 0 value means no size limit
// 		HardMaxCacheSize: 8192,
// 		// callback fired when the oldest entry is removed because of its expiration time or no space left
// 		// for the new entry, or because delete was called. A bitmask representing the reason will be returned.
// 		// Default value is nil which means no callback and it prevents from unwrapping the oldest entry.
// 		OnRemove: nil,
// 		// OnRemoveWithReason is a callback fired when the oldest entry is removed because of its expiration time or no space left
// 		// for the new entry, or because delete was called. A constant representing the reason will be passed through.
// 		// Default value is nil which means no callback and it prevents from unwrapping the oldest entry.
// 		// Ignored if OnRemove is specified.
// 		// OnRemoveWithReason: nil,
// 	}
// 	cache := &BigCache{}
// 	innerCache, initErr := bigcache.NewBigCache(config)
// 	cache.innerCache = innerCache
//
// 	return cache, initErr
// }
//
// // BigCache implementation of a cache's Get
// func (bc *BigCache) Get(key string) (Flow, error) {
// 	bytesBlob, err := bc.innerCache.Get(key)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	//TODO: Cast should be done dependent on the stored data...
// 	f := &GeneralFlow{}
// 	err = json.Unmarshal(bytesBlob, f)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return f, nil
// }
//
// // BigCache implementation of a cache's Set
// func (bc *BigCache) Set(key string, f Flow) error {
// 	bytesBlob, err := f.MarshalBinary()
// 	if err != nil {
// 		return err
// 	}
//
// 	bc.innerCache.Set(key, bytesBlob)
// 	if err != nil {
// 		return err
// 	}
//
// 	return nil
// }
//
// // BigCache implementation of a cache's Clear
// func (bc *BigCache) Clear() error {
// 	return bc.innerCache.Reset()
// }
//
// // BigCache implementation of a cache's PurgeExpired
// // Always returns an error as no inner function of
// // BigCache allows its implementation
// func (bc *BigCache) PurgeExpired() error {
// 	return errors.New("BigCache: not implemented")
// }
