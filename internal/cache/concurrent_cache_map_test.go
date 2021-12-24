package cache

import (
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"
)

type TestStruct struct {
	Num      int
	Children []*TestStruct
}

func TestCache(t *testing.T) {
	tc := NewConcurrentCacheMap(DEFAULT_SHARD_COUNT, DEFAULT_EXPIRATION, nil, 0)

	a, found := tc.Get("a")
	if found || a != nil {
		t.Error("Getting A found value that shouldn't exist:", a)
	}

	b, found := tc.Get("b")
	if found || b != nil {
		t.Error("Getting B found value that shouldn't exist:", b)
	}

	c, found := tc.Get("c")
	if found || c != nil {
		t.Error("Getting C found value that shouldn't exist:", c)
	}

	tc.Set("a", 1)
	tc.Set("b", "b")
	tc.Set("c", 3.5)

	x, found := tc.Get("a")
	if !found {
		t.Error("a was not found while getting a2")
	}
	if x == nil {
		t.Error("x for a is nil")
	} else if a2 := x.(int); a2+2 != 3 {
		t.Error("a2 (which should be 1) plus 2 does not equal 3; value:", a2)
	}

	x, found = tc.Get("b")
	if !found {
		t.Error("b was not found while getting b2")
	}
	if x == nil {
		t.Error("x for b is nil")
	} else if b2 := x.(string); b2+"B" != "bB" {
		t.Error("b2 (which should be b) plus B does not equal bB; value:", b2)
	}

	x, found = tc.Get("c")
	if !found {
		t.Error("c was not found while getting c2")
	}
	if x == nil {
		t.Error("x for c is nil")
	} else if c2 := x.(float64); c2+1.2 != 4.7 {
		t.Error("c2 (which should be 3.5) plus 1.2 does not equal 4.7; value:", c2)
	}
}

func TestCacheTimes(t *testing.T) {
	var found bool

	tc := NewConcurrentCacheMap(DEFAULT_SHARD_COUNT, 1*time.Millisecond, nil, 2*time.Second)
	tc.Set("a", 1)
	tc.Set("b", 2)
	tc.Set("c", 3)
	tc.Set("d", 4)

	<-time.After(5 * time.Second)
	_, found = tc.Get("c")
	if found {
		t.Error("Found c when it should have been automatically deleted")
	}

	_, found = tc.Get("a")
	if found {
		t.Error("Found a when it should have been automatically deleted")
	}

	_, found = tc.Get("b")
	if found {
		t.Error("Found a when it should have been automatically deleted")
	}

	_, found = tc.Get("d")
	if found {
		t.Error("Found a when it should have been automatically deleted")
	}
}

func TestStorePointerToStruct(t *testing.T) {
	tc := NewConcurrentCacheMap(DEFAULT_SHARD_COUNT, DEFAULT_EXPIRATION, nil, 0)
	tc.Set("foo", &TestStruct{Num: 1})
	x, found := tc.Get("foo")
	if !found {
		t.Fatal("*TestStruct was not found for foo")
	}
	foo := x.(*TestStruct)
	foo.Num++

	y, found := tc.Get("foo")
	if !found {
		t.Fatal("*TestStruct was not found for foo (second time)")
	}
	bar := y.(*TestStruct)
	if bar.Num != 2 {
		t.Fatal("TestStruct.Num is not 2")
	}
}

func BenchmarkCacheGetNotExpiring(b *testing.B) {
	benchmarkCacheGet(b, 0)
}

func benchmarkCacheGet(b *testing.B, exp time.Duration) {
	b.StopTimer()
	tc := NewConcurrentCacheMap(DEFAULT_SHARD_COUNT, exp, nil, 0)
	tc.Set("foo", "bar")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tc.Get("foo")
	}
}

func BenchmarkCacheGetConcurrentNotExpiring(b *testing.B) {
	benchmarkCacheGetConcurrent(b, 0)
}

func benchmarkCacheGetConcurrent(b *testing.B, exp time.Duration) {
	b.StopTimer()
	tc := NewConcurrentCacheMap(DEFAULT_SHARD_COUNT, exp, nil, 0)
	tc.Set("foo", "bar")
	wg := new(sync.WaitGroup)
	workers := runtime.NumCPU()
	each := b.N / workers
	wg.Add(workers)
	b.StartTimer()
	for i := 0; i < workers; i++ {
		go func() {
			for j := 0; j < each; j++ {
				tc.Get("foo")
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func BenchmarkCacheGetManyConcurrentNotExpiring(b *testing.B) {
	benchmarkCacheGetManyConcurrent(b, 0)
}

func benchmarkCacheGetManyConcurrent(b *testing.B, exp time.Duration) {
	// This is the same as BenchmarkCacheGetConcurrent, but its result
	// can be compared against BenchmarkShardedCacheGetManyConcurrent
	// in sharded_test.go.
	b.StopTimer()
	workers := runtime.NumCPU()
	tc := NewConcurrentCacheMap(DEFAULT_SHARD_COUNT, exp, nil, 0)
	// Benchmark using 1000 keys
	n := 1000
	keys := make([]string, n)
	for i := 0; i < n; i++ {
		k := "foo" + strconv.Itoa(n)
		keys[i] = k
		tc.Set(k, "bar")
	}
	each := b.N / workers
	queries := make([][]int, workers)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < workers; i++ {
		queries[i] = make([]int, each)
		for j := 0; j < each; j++ {
			queries[i][j] = rand.Intn(n)
		}
	}
	wg := new(sync.WaitGroup)
	wg.Add(workers)
	b.StartTimer()
	for i := 0; i < workers; i++ {
		go func(worker int) {
			for j := 0; j < each; j++ {
				tc.Get(keys[queries[worker][j]])
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func BenchmarkCacheSetNotExpiring(b *testing.B) {
	benchmarkCacheSet(b, 0)
}

func benchmarkCacheSet(b *testing.B, exp time.Duration) {
	b.StopTimer()
	tc := NewConcurrentCacheMap(DEFAULT_SHARD_COUNT, exp, nil, 0)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tc.Set("foo", "bar")
	}
}

func BenchmarkCacheSetConcurrentNotExpiring(b *testing.B) {
	benchmarkCacheSetConcurrent(b, 0)
}

func benchmarkCacheSetConcurrent(b *testing.B, exp time.Duration) {
	b.StopTimer()
	tc := NewConcurrentCacheMap(DEFAULT_SHARD_COUNT, exp, nil, 0)
	wg := new(sync.WaitGroup)
	workers := runtime.NumCPU()
	each := b.N / workers
	wg.Add(workers)
	b.StartTimer()
	for i := 0; i < workers; i++ {
		go func() {
			for j := 0; j < each; j++ {
				tc.Set("foo", "bar")
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func BenchmarkCacheSetManyConcurrentNotExpiring(b *testing.B) {
	benchmarkCacheSetManyConcurrent(b, 0)
}

func benchmarkCacheSetManyConcurrent(b *testing.B, exp time.Duration) {
	// This is the same as BenchmarkCacheGetConcurrent, but its result
	// can be compared against BenchmarkShardedCacheGetManyConcurrent
	// in sharded_test.go.
	b.StopTimer()
	workers := runtime.NumCPU()
	tc := NewConcurrentCacheMap(DEFAULT_SHARD_COUNT, exp, nil, 0)
	// Benchmark using 1000 keys
	n := 1000
	keys := make([]string, n)
	for i := 0; i < n; i++ {
		k := "foo" + strconv.Itoa(n)
		keys[i] = k
	}
	each := b.N / workers
	wg := new(sync.WaitGroup)
	wg.Add(workers)
	queries := make([][]int, workers)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < workers; i++ {
		queries[i] = make([]int, each)
		for j := 0; j < each; j++ {
			queries[i][j] = rand.Intn(n)
		}
	}
	b.StartTimer()
	for i := 0; i < workers; i++ {
		go func(worker int) {
			for j := 0; j < each; j++ {
				tc.Set(keys[queries[worker][j]], "bar")
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func BenchmarkCache1090ManyConcurrentNotExpiring(b *testing.B) {
	benchmarkCache1090ManyConcurrent(b, 0)
}

func benchmarkCache1090ManyConcurrent(b *testing.B, exp time.Duration) {
	// This is the same as BenchmarkCacheGetConcurrent, but its result
	// can be compared against BenchmarkShardedCacheGetManyConcurrent
	// in sharded_test.go.
	b.StopTimer()
	workers := runtime.NumCPU()
	tc := NewConcurrentCacheMap(DEFAULT_SHARD_COUNT, exp, nil, 0)
	// Benchmark using 1000 keys
	n := 1000
	keys := make([]string, n)
	for i := 0; i < n; i++ {
		k := "foo" + strconv.Itoa(n)
		keys[i] = k
	}
	each := b.N / workers
	wg := new(sync.WaitGroup)
	wg.Add(workers)
	queries := make([][]int, workers)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < workers; i++ {
		queries[i] = make([]int, each)
		for j := 0; j < each; j++ {
			queries[i][j] = rand.Intn(n)
		}
	}
	b.StartTimer()
	for i := 0; i < workers; i++ {
		go func(worker int) {
			for j := 0; j < each; j++ {
				if rand.Intn(10) > 0 {
					tc.Get(keys[queries[worker][j]])
				} else {
					tc.Set(keys[queries[worker][j]], "bar")
				}
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}
