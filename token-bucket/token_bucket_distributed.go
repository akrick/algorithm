package tokenbucket

import (
	"fmt"
	"sync"
	"time"
)

// TokenBucketDistributed 分布式令牌桶结构（基于Redis）
type TokenBucketDistributed struct {
	capacity     int       // 桶的容量
	rate         int       // 令牌生成速率（每秒）
	lastRefill   time.Time // 上次填充时间
	tokens       int       // 当前令牌数
	mu           sync.Mutex
}

// NewTokenBucketDistributed 创建一个新的分布式令牌桶
func NewTokenBucketDistributed(capacity, rate int) *TokenBucketDistributed {
	return &TokenBucketDistributed{
		capacity:   capacity,
		tokens:     capacity,
		rate:       rate,
		lastRefill: time.Now(),
	}
}

// TryConsumeDistributed 尝试消费令牌（分布式场景）
func (tbd *TokenBucketDistributed) TryConsumeDistributed(count int) bool {
	tbd.mu.Lock()
	defer tbd.mu.Unlock()

	tbd.refill()

	if tbd.tokens >= count {
		tbd.tokens -= count
		return true
	}
	return false
}

// refill 补充令牌
func (tbd *TokenBucketDistributed) refill() {
	now := time.Now()
	elapsed := now.Sub(tbd.lastRefill).Seconds()

	newTokens := int(elapsed * float64(tbd.rate))

	if newTokens > 0 {
		tbd.tokens += newTokens
		if tbd.tokens > tbd.capacity {
			tbd.tokens = tbd.capacity
		}
		tbd.lastRefill = now
	}
}

// GetTokensDistributed 获取当前令牌数
func (tbd *TokenBucketDistributed) GetTokensDistributed() int {
	tbd.mu.Lock()
	defer tbd.mu.Unlock()
	tbd.refill()
	return tbd.tokens
}

// InfoDistributed 获取令牌桶信息
func (tbd *TokenBucketDistributed) InfoDistributed() string {
	tbd.mu.Lock()
	defer tbd.mu.Unlock()
	tbd.refill()
	return fmt.Sprintf("Capacity: %d, Rate: %d/s, Tokens: %d", tbd.capacity, tbd.rate, tbd.tokens)
}
