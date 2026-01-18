# 布隆过滤器 - 解决 Redis 缓存穿透问题

## 什么是缓存穿透？

缓存穿透是指查询一个**一定不存在**的数据，由于缓存中没有命中（Redis 没有该数据），请求会直接穿透到数据库。如果有大量这样的请求，数据库会承受巨大压力，甚至崩溃。

## 什么是布隆过滤器？

布隆过滤器（Bloom Filter）是一种空间效率极高的概率型数据结构，用于判断一个元素是否在一个集合中：

- **可能存在**：元素可能存在（有一定的误判率）
- **一定不存在**：元素绝对不存在

## 为什么布隆过滤器能解决缓存穿透？

布隆过滤器可以**快速判断**一个 key 是否可能存在于数据库中：
- 如果布隆过滤器说 key 不存在，那么可以直接返回，不需要查询数据库
- 如果布隆过滤器说 key 可能存在，再去查询 Redis 和数据库

这样可以拦截大部分无效请求，避免数据库承受巨大压力。

## 实现结构

```
bloom-filter/
├── bloom_filter.go           # 布隆过滤器核心实现
├── bloom_filter_test.go      # 单元测试
├── cache_penetration.go      # 缓存穿透解决方案示例
└── cache_penetration_test.go # 缓存穿透测试
```

## 使用示例

```go
package main

import (
    "fmt"
    "bloomfilter"
)

func main() {
    // 初始化 Redis 和数据库
    redis := bloomfilter.NewMockRedis()
    db := bloomfilter.NewMockDatabase()
    
    // 创建带布隆过滤器的缓存
    // 参数：Redis, 数据库, 预期元素数量
    cache := bloomfilter.NewCacheWithBloomFilter(redis, db, 100)
    
    // 查询数据
    data, err := cache.GetData("user:1")
    if err != nil {
        fmt.Printf("查询失败: %v\n", err)
        return
    }
    fmt.Printf("数据: %s\n", data)
}
```

## 核心方法

### BloomFilter

| 方法 | 说明 |
|------|------|
| `NewBloomFilter(n, p)` | 创建布隆过滤器，n=预期元素数，p=误判率 |
| `Add(data)` | 添加元素 |
| `Contains(data)` | 检查元素是否存在 |
| `Clear()` | 清空过滤器 |

### CacheWithBloomFilter

| 方法 | 说明 |
|------|------|
| `NewCacheWithBloomFilter(redis, db, n)` | 创建带布隆过滤器的缓存 |
| `GetData(key)` | 获取数据（自动应用布隆过滤器） |

## 性能对比

### 查询流程对比

**不使用布隆过滤器：**
```
请求 → Redis(未命中) → 数据库(未命中) → 返回空
```

**使用布隆过滤器：**
```
请求 → 布隆过滤器(不存在) → 直接返回
请求 → 布隆过滤器(可能存在) → Redis(未命中) → 数据库
```

### 性能优势

- **时间复杂度**：O(k)，k 为哈希函数数量（通常 3-8 个）
- **空间复杂度**：仅需要几个比特位存储每个元素
- **查询速度**：极快，避免数据库查询
- **内存占用**：相比存储所有 key，节省大量内存

## 参数选择

### 预期元素数量 (n)

根据业务场景预估，比如用户表有 100 万用户，n 设为 100 万

### 误判率 (p)

- **p = 0.01** (1%)：推荐值，平衡精度和空间
- **p = 0.001** (0.1%)：高精度，需要更多内存
- **p = 0.1** (10%)：低精度，节省内存

### 计算公式

```
位图大小 m = -n * ln(p) / (ln(2)²)
哈希函数数量 k = m / n * ln(2)
```

## 注意事项

### 1. 误判率

布隆过滤器有一定的误判率，可能会：
- 误判不存在的 key 为存在（概率很低）
- 但绝不会误判存在的 key 为不存在

### 2. 不支持删除

标准布隆过滤器**不支持删除操作**，因为多个元素可能共享同一个位。

### 3. 数据预热

在系统启动时，需要将数据库中所有已存在的 key 添加到布隆过滤器中。

### 4. 数据更新

当数据库新增数据时，需要同步更新布隆过滤器。

## 测试运行

```bash
# 运行所有测试
go test ./bloom-filter/...

# 运行并显示覆盖率
go test -cover ./bloom-filter/...

# 运行基准测试
go test -bench=. ./bloom-filter/...

# 运行示例
go run ./bloom-filter/cache_penetration.go
```

## 测试结果示例

```
=== 布隆过滤器防止缓存穿透示例 ===

1. 查询存在的数据 user:1:
数据库查询并缓存: user:1
结果: user_data_1

2. 再次查询 user:1（应该从 Redis 获取）:
Redis命中: user:1
结果: user_data_1

3. 查询不存在的数据 invalid:999:
错误: key not found in bloom filter: invalid:999 (被布隆过滤器拦截，不查询数据库)

4. 模拟缓存穿透攻击（1000次不存在的查询）:
总请求数: 1000
被布隆过滤器拦截: 998
耗时: 8.234ms
平均每次请求: 8.234µs
```

## 适用场景

✅ 适合使用布隆过滤器的场景：
- 缓存穿透防护
- 邮箱黑名单/白名单
- 网页爬虫去重
- 数据库查询优化

❌ 不适合使用布隆过滤器的场景：
- 需要精确判断的场景
- 需要频繁删除元素的场景
- 元素数量不确定的场景

## 扩展阅读

- [布隆过滤器原论文](https://doi.org/10.1145/362686.362692)
- [Redis 布隆过滤器模块](https://redis.io/docs/data-types/probabilistic/bloom-filter/)
- [Guava BloomFilter](https://github.com/google/guava/blob/master/guava/src/com/google/common/hash/BloomFilter.java)
