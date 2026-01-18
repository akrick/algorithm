package bloomfilter

import (
	"fmt"
	"testing"
)

// TestCacheWithBloomFilter 测试带布隆过滤器的缓存
func TestCacheWithBloomFilter(t *testing.T) {
	redis := NewMockRedis()
	db := NewMockDatabase()
	cache := NewCacheWithBloomFilter(redis, db, 100)
	
	// 测试 1: 查询存在的数据
	data, err := cache.GetData("user:1")
	if err != nil {
		t.Errorf("查询存在的数据失败: %v", err)
	}
	if data != "user_data_1" {
		t.Errorf("数据不匹配，期望: user_data_1, 实际: %s", data)
	}
	
	// 测试 2: 再次查询（应该从缓存获取）
	data, err = cache.GetData("user:1")
	if err != nil {
		t.Errorf("第二次查询失败: %v", err)
	}
	if data != "user_data_1" {
		t.Errorf("数据不匹配")
	}
	
	// 测试 3: 查询不存在的数据（应该被布隆过滤器拦截）
	_, err = cache.GetData("invalid:999")
	if err == nil {
		t.Error("查询不存在的数据应该返回错误")
	}
}

// TestCachePenetrationPrevention 测试缓存穿透防护
func TestCachePenetrationPrevention(t *testing.T) {
	redis := NewMockRedis()
	db := NewMockDatabase()
	cache := NewCacheWithBloomFilter(redis, db, 100)
	
	// 模拟大量不存在的查询（缓存穿透攻击）
	attackCount := 1000
	blockedCount := 0
	
	for i := 0; i < attackCount; i++ {
		key := fmt.Sprintf("attack:%d", i)
		_, err := cache.GetData(key)
		if err != nil {
			blockedCount++
		}
	}
	
	t.Logf("总攻击数: %d, 被拦截: %d", attackCount, blockedCount)
	
	// 大部分请求应该被拦截
	if blockedCount < attackCount*95/100 {
		t.Errorf("布隆过滤器拦截率过低: %d/%d", blockedCount, attackCount)
	}
}

// BenchmarkGetDataWithBloomFilter 带布隆过滤器的基准测试
func BenchmarkGetDataWithBloomFilter(b *testing.B) {
	redis := NewMockRedis()
	db := NewMockDatabase()
	cache := NewCacheWithBloomFilter(redis, db, 100)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 混合查询存在的和不存在的数据
		if i%2 == 0 {
			cache.GetData(fmt.Sprintf("user:%d", i%100))
		} else {
			cache.GetData(fmt.Sprintf("attack:%d", i))
		}
	}
}

// BenchmarkGetDataWithoutBloomFilter 不带布隆过滤器的基准测试
func BenchmarkGetDataWithoutBloomFilter(b *testing.B) {
	redis := NewMockRedis()
	db := NewMockDatabase()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 查询 Redis
		if value, ok := redis.Get(fmt.Sprintf("user:%d", i%100)); ok {
			_ = value
			continue
		}
		
		// 查询数据库
		if value, ok := db.Query(fmt.Sprintf("user:%d", i%100)); ok {
			redis.Set(fmt.Sprintf("user:%d", i%100), value)
		}
	}
}
