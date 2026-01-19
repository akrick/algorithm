package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"example.com/go-examples/algorithm/singleflight"
)

// cache 模拟缓存层
type cache struct {
	data map[string]string
	mu   sync.RWMutex
}

func newCache() *cache {
	return &cache{data: make(map[string]string)}
}

func (c *cache) get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.data[key]
	return val, ok
}

func (c *cache) set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

// db 模拟数据库查询
type db struct {
	queryCount int32
}

func (d *db) query(key string) (string, error) {
	atomic.AddInt32(&d.queryCount, 1)
	fmt.Printf("📊 查询数据库: %s\n", key)
	time.Sleep(300 * time.Millisecond)
	return fmt.Sprintf("db-value-%s", key), nil
}

// cacheDecorator 使用 SingleFlight 的缓存装饰器
type cacheDecorator struct {
	cache     *cache
	db        *db
	single    *singleflight.Group
	hitCount  int32
	missCount int32
}

func newCacheDecorator(cache *cache, db *db) *cacheDecorator {
	return &cacheDecorator{
		cache:  cache,
		db:     db,
		single: &singleflight.Group{},
	}
}

func (cd *cacheDecorator) get(key string) (string, error) {
	if val, ok := cd.cache.get(key); ok {
		atomic.AddInt32(&cd.hitCount, 1)
		fmt.Printf("✅ 缓存命中: %s\n", key)
		return val, nil
	}

	atomic.AddInt32(&cd.missCount, 1)
	fmt.Printf("❌ 缓存未命中: %s\n", key)

	val, err := cd.single.Do(key, func() (interface{}, error) {
		return cd.db.query(key)
	})

	if err != nil {
		return "", err
	}

	cd.cache.set(key, val.(string))
	fmt.Printf("💾 写入缓存: %s\n", key)

	return val.(string), nil
}

func main() {
	fmt.Println("=== SingleFlight 缓存防击穿示例 ===\n")

	cache := newCache()
	db := &db{}
	decorator := newCacheDecorator(cache, db)

	fmt.Println("🔥 模拟缓存过期后的并发请求")
	fmt.Println()

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			time.Sleep(time.Duration(10+id%20) * time.Millisecond)

			val, err := decorator.get("user:12345")
			if err != nil {
				fmt.Printf("❌ 请求 %d 失败: %v\n", id, err)
				return
			}
			fmt.Printf("请求 %d 获取到值: %s\n", id, val)
		}(i)
	}

	wg.Wait()

	fmt.Println()
	fmt.Println("=== 统计信息 ===")
	fmt.Printf("数据库查询次数: %d\n", atomic.LoadInt32(&db.queryCount))
	fmt.Printf("缓存命中次数: %d\n", atomic.LoadInt32(&decorator.hitCount))
	fmt.Printf("缓存未命中次数: %d\n", atomic.LoadInt32(&decorator.missCount))

	if atomic.LoadInt32(&db.queryCount) == 1 {
		fmt.Println("\n✅ 成功! SingleFlight 防止了缓存击穿,只查询了一次数据库")
	} else {
		fmt.Println("\n❌ 失败! 查询了多次数据库")
	}
}
