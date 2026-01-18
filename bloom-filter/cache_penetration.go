package bloomfilter

import (
	"fmt"
	"time"
)

// MockRedis 模拟 Redis 客户端
type MockRedis struct {
	cache map[string]string
}

func NewMockRedis() *MockRedis {
	return &MockRedis{
		cache: make(map[string]string),
	}
}

func (r *MockRedis) Get(key string) (string, bool) {
	val, ok := r.cache[key]
	return val, ok
}

func (r *MockRedis) Set(key, value string) {
	r.cache[key] = value
}

// MockDatabase 模拟数据库
type MockDatabase struct {
	data map[string]string
}

func NewMockDatabase() *MockDatabase {
	data := make(map[string]string)
	// 初始化一些数据
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("user:%d", i)
		data[key] = fmt.Sprintf("user_data_%d", i)
	}
	return &MockDatabase{data: data}
}

func (d *MockDatabase) Query(key string) (string, bool) {
	time.Sleep(10 * time.Millisecond) // 模拟数据库查询延迟
	val, ok := d.data[key]
	return val, ok
}

// CacheWithBloomFilter 使用布隆过滤器防止缓存穿透
type CacheWithBloomFilter struct {
	bloomFilter *BloomFilter
	redis       *MockRedis
	database    *MockDatabase
}

func NewCacheWithBloomFilter(redis *MockRedis, db *MockDatabase, expectedElements int) *CacheWithBloomFilter {
	// 创建布隆过滤器，误判率设置为 0.01
	bf := NewBloomFilter(expectedElements, 0.01)
	
	// 预热布隆过滤器：将数据库中所有已存在的 key 添加到布隆过滤器
	for key := range db.data {
		bf.Add([]byte(key))
	}
	
	return &CacheWithBloomFilter{
		bloomFilter: bf,
		redis:       redis,
		database:    db,
	}
}

// GetData 获取数据，使用布隆过滤器防止缓存穿透
func (c *CacheWithBloomFilter) GetData(key string) (string, error) {
	// 第一步：检查布隆过滤器
	if !c.bloomFilter.Contains([]byte(key)) {
		// 布隆过滤器说这个 key 一定不存在，直接返回
		return "", fmt.Errorf("key not found in bloom filter: %s", key)
	}
	
	// 第二步：查询 Redis 缓存
	if value, ok := c.redis.Get(key); ok {
		fmt.Printf("Redis命中: %s\n", key)
		return value, nil
	}
	
	// 第三步：查询数据库
	if value, ok := c.database.Query(key); ok {
		// 数据库中存在，写入 Redis 缓存
		c.redis.Set(key, value)
		fmt.Printf("数据库查询并缓存: %s\n", key)
		return value, nil
	}
	
	// 数据库中也不存在
	return "", fmt.Errorf("key not found: %s", key)
}

// Example 使用示例
func ExampleUsage() {
	// 初始化 Redis 和数据库
	redis := NewMockRedis()
	db := NewMockDatabase()
	
	// 创建带布隆过滤器的缓存
	cache := NewCacheWithBloomFilter(redis, db, 100)
	
	fmt.Println("=== 布隆过滤器防止缓存穿透示例 ===\n")
	
	// 示例 1: 查询存在的数据
	fmt.Println("1. 查询存在的数据 user:1:")
	result, err := cache.GetData("user:1")
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		fmt.Printf("结果: %s\n\n", result)
	}
	
	// 示例 2: 再次查询相同数据（应该从 Redis 获取）
	fmt.Println("2. 再次查询 user:1（应该从 Redis 获取）:")
	result, err = cache.GetData("user:1")
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		fmt.Printf("结果: %s\n\n", result)
	}
	
	// 示例 3: 查询不存在的数据（被布隆过滤器拦截）
	fmt.Println("3. 查询不存在的数据 invalid:999:")
	result, err = cache.GetData("invalid:999")
	if err != nil {
		fmt.Printf("错误: %v (被布隆过滤器拦截，不查询数据库)\n\n", err)
	} else {
		fmt.Printf("结果: %s\n\n", result)
	}
	
	// 示例 4: 批量攻击测试
	fmt.Println("4. 模拟缓存穿透攻击（1000次不存在的查询）:")
	start := time.Now()
	attackCount := 1000
	blockedCount := 0
	
	for i := 0; i < attackCount; i++ {
		key := fmt.Sprintf("attack:%d", i)
		_, err := cache.GetData(key)
		if err != nil {
			blockedCount++
		}
	}
	
	elapsed := time.Since(start)
	fmt.Printf("总请求数: %d\n", attackCount)
	fmt.Printf("被布隆过滤器拦截: %d\n", blockedCount)
	fmt.Printf("耗时: %v\n", elapsed)
	fmt.Printf("平均每次请求: %v\n\n", elapsed/time.Duration(attackCount))
}
