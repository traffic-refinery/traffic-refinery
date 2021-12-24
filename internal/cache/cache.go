package cache

type Cache interface {
	// Get returns a copy of the cache element
	Get(string) (interface{}, bool)
	// GetAndLock returns a copy of the cached element, but maintains the lock in the cache
	GetAndLock(key string) (interface{}, bool)
	// Set inserts the element into the cache.
	// If an element with the same key exists it replaces it.
	Set(string, interface{}) error
	// SetAndUnlock inserts the element into the cache.
	// If an element with the same key exists it replaces it.
	// Finally it unlocks the mutex
	SetAndUnlock(string, interface{}) error
	// Unlock unlocks the mutex associated with a given key
	Unlock(string) error
	// Clear empties the entire cache
	Clear() error
	// PurgeExpired removes expired elements from the cache
	PurgeExpired() error
	// Dump returns a dump of the entire cache
	Dump() map[string]interface{}
	//
	IterativeDump() (CacheIterator, error)
	//
	NextElement(CacheIterator) (interface{}, error)
	//
	StopIteration(CacheIterator) error
}

// Interface used to iterate over a cache and avoid changes on elements that are
// currently selected
type CacheIterator interface {
	getNextElement() (interface{}, bool)
	clear()
}
