package tokenbucket

import (
	"fmt"
	"sync"
	"time"
)

// TokenBucket 令牌桶结构
type TokenBucket struct {
	capacity     int       // 桶的容量
	tokens       int       // 当前令牌数
	rate         int       // 令牌生成速率（每秒）
	lastRefill   time.Time // 上次填充时间
	mu           sync.Mutex
}

// NewTokenBucket 创建一个新的令牌桶
// capacity: 桶的容量
// rate: 令牌生成速率（每秒）
func NewTokenBucket(capacity, rate int) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity, // 初始时桶满
		rate:       rate,
		lastRefill: time.Now(),
	}
}

// TryConsume 尝试消费令牌
// count: 需要消费的令牌数
// 返回: 是否成功消费
func (tb *TokenBucket) TryConsume(count int) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// 重新计算令牌数
	tb.refill()

	if tb.tokens >= count {
		tb.tokens -= count
		return true
	}
	return false
}

// refill 根据时间间隔补充令牌
func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()

	// 计算应该补充的令牌数
	newTokens := int(elapsed * float64(tb.rate))

	if newTokens > 0 {
		tb.tokens += newTokens
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefill = now
	}
}

// GetTokens 获取当前令牌数（仅用于测试）
func (tb *TokenBucket) GetTokens() int {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.refill()
	return tb.tokens
}

// Info 获取令牌桶信息
func (tb *TokenBucket) Info() string {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.refill()
	return fmt.Sprintf("Capacity: %d, Rate: %d/s, Tokens: %d", tb.capacity, tb.rate, tb.tokens)
}
