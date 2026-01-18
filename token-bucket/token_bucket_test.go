package main

import (
	"fmt"
	"time"
)

// TokenBucket 令牌桶结构
type TokenBucket struct {
	capacity     int       // 桶的容量
	tokens       int       // 当前令牌数
	rate         int       // 令牌生成速率（每秒）
	lastRefill   time.Time // 上次填充时间
}

// NewTokenBucket 创建一个新的令牌桶
func NewTokenBucket(capacity, rate int) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		rate:       rate,
		lastRefill: time.Now(),
	}
}

// TryConsume 尝试消费令牌
func (tb *TokenBucket) TryConsume(count int) bool {
	tb.refill()

	if tb.tokens >= count {
		tb.tokens -= count
		return true
	}
	return false
}

// refill 补充令牌
func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()

	newTokens := int(elapsed * float64(tb.rate))

	if newTokens > 0 {
		tb.tokens += newTokens
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefill = now
	}
}

// GetTokens 获取当前令牌数
func (tb *TokenBucket) GetTokens() int {
	tb.refill()
	return tb.tokens
}

func main() {
	// 创建一个容量为10，速率为5的令牌桶
	tb := NewTokenBucket(10, 5)

	fmt.Println("=== 令牌桶算法演示 ===")
	fmt.Println(tb.Info())

	// 模拟请求
	fmt.Println("\n=== 模拟请求 ===")
	requests := []struct {
		name   string
		count  int
		delay  time.Duration
	}{
		{"请求1", 3, 0},
		{"请求2", 5, 100 * time.Millisecond},
		{"请求3", 4, 0},
		{"请求4", 2, 200 * time.Millisecond},
		{"请求5", 6, 0},
	}

	for _, req := range requests {
		if req.delay > 0 {
			time.Sleep(req.delay)
		}
		success := tb.TryConsume(req.count)
		status := "❌ 被限流"
		if success {
			status = "✓ 通过"
		}
		fmt.Printf("%s: 消费%d个令牌 - %s (剩余: %d)\n",
			req.name, req.count, status, tb.GetTokens())
	}

	// 等待令牌补充
	fmt.Println("\n=== 等待令牌补充 ===")
	time.Sleep(1 * time.Second)
	fmt.Printf("等待1秒后: %s\n", tb.Info())

	// 并发测试
	fmt.Println("\n=== 并发测试 ===")
	tb2 := NewTokenBucket(5, 2)
	successCount := 0
	for i := 0; i < 10; i++ {
		if tb2.TryConsume(1) {
			successCount++
		}
	}
	fmt.Printf("10个并发请求，通过: %d，被限流: %d\n", successCount, 10-successCount)
}

// Info 获取令牌桶信息
func (tb *TokenBucket) Info() string {
	tb.refill()
	return fmt.Sprintf("容量: %d, 速率: %d/s, 当前令牌: %d", tb.capacity, tb.rate, tb.tokens)
}
