package bloomfilter

import (
	"hash"
	"hash/fnv"
	"math"
)

// BloomFilter 布隆过滤器结构
type BloomFilter struct {
	bitSet    []bool      // 位图
	size      int         // 位图大小
	hashFuncs []hash.Hash64 // 哈希函数列表
}

// NewBloomFilter 创建一个新的布隆过滤器
// n: 预计插入的元素数量
// p: 期望的误判率 (0 < p < 1)
func NewBloomFilter(n int, p float64) *BloomFilter {
	// 计算最优的位图大小 m
	m := optimalSize(n, p)
	
	// 计算最优的哈希函数数量 k
	k := optimalHashCount(n, m)
	
	// 创建位图
	bitSet := make([]bool, m)
	
	// 创建哈希函数
	hashFuncs := make([]hash.Hash64, k)
	for i := 0; i < k; i++ {
		hashFuncs[i] = fnv.New64a()
	}
	
	return &BloomFilter{
		bitSet:    bitSet,
		size:      m,
		hashFuncs: hashFuncs,
	}
}

// optimalSize 计算最优的位图大小
func optimalSize(n int, p float64) int {
	m := -float64(n) * math.Log(p) / (math.Ln2 * math.Ln2)
	return int(math.Ceil(m))
}

// optimalHashCount 计算最优的哈希函数数量
func optimalHashCount(n, m int) int {
	k := float64(m) / float64(n) * math.Ln2
	return int(math.Ceil(k))
}

// Add 添加元素到布隆过滤器
func (bf *BloomFilter) Add(data []byte) {
	for i, h := range bf.hashFuncs {
		// 重置哈希函数
		h.Reset()
		
		// 写入数据
		h.Write(data)
		
		// 写入索引以区分不同的哈希函数
		h.Write([]byte{byte(i)})
		
		// 获取哈希值并计算位置
		hashValue := h.Sum64()
		position := int(hashValue % uint64(bf.size))
		
		// 设置位
		bf.bitSet[position] = true
	}
}

// Contains 检查元素是否可能存在
// 返回 true 表示可能存在, false 表示一定不存在
func (bf *BloomFilter) Contains(data []byte) bool {
	for i, h := range bf.hashFuncs {
		// 重置哈希函数
		h.Reset()
		
		// 写入数据
		h.Write(data)
		
		// 写入索引以区分不同的哈希函数
		h.Write([]byte{byte(i)})
		
		// 获取哈希值并计算位置
		hashValue := h.Sum64()
		position := int(hashValue % uint64(bf.size))
		
		// 如果任意一位为 false, 则元素一定不存在
		if !bf.bitSet[position] {
			return false
		}
	}
	
	// 所有位都为 true, 元素可能存在
	return true
}

// Clear 清空布隆过滤器
func (bf *BloomFilter) Clear() {
	for i := range bf.bitSet {
		bf.bitSet[i] = false
	}
}

// Size 返回布隆过滤器的大小
func (bf *BloomFilter) Size() int {
	return bf.size
}

// HashCount 返回哈希函数数量
func (bf *BloomFilter) HashCount() int {
	return len(bf.hashFuncs)
}
