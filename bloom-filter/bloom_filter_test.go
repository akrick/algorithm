package bloomfilter

import (
	"fmt"
	"math/rand"
	"testing"
)

// TestBasic 测试基本的添加和查找功能
func TestBasic(t *testing.T) {
	bf := NewBloomFilter(1000, 0.01)
	
	// 测试不存在的元素
	if bf.Contains([]byte("hello")) {
		t.Error("空布隆过滤器不应该包含任何元素")
	}
	
	// 添加元素
	bf.Add([]byte("hello"))
	
	// 测试存在的元素
	if !bf.Contains([]byte("hello")) {
		t.Error("布隆过滤器应该包含刚添加的元素")
	}
	
	// 测试不存在的元素（可能误判，但大概率返回false）
	if bf.Contains([]byte("world")) {
		t.Log("注意：可能出现误判，这是布隆过滤器的特性")
	}
}

// TestMultipleElements 测试多个元素
func TestMultipleElements(t *testing.T) {
	bf := NewBloomFilter(1000, 0.01)
	
	elements := []string{"apple", "banana", "cherry", "date", "elderberry"}
	
	// 添加多个元素
	for _, elem := range elements {
		bf.Add([]byte(elem))
	}
	
	// 验证所有添加的元素都能被找到
	for _, elem := range elements {
		if !bf.Contains([]byte(elem)) {
			t.Errorf("布隆过滤器应该包含 %s", elem)
		}
	}
}

// TestFalsePositive 测试误判率
func TestFalsePositive(t *testing.T) {
	n := 1000
	p := 0.01
	bf := NewBloomFilter(n, p)
	
	// 添加 n 个元素
	for i := 0; i < n; i++ {
		bf.Add([]byte(fmt.Sprintf("item%d", i)))
	}
	
	// 测试 1000 个不存在的元素
	falsePositive := 0
	tests := 1000
	for i := 0; i < tests; i++ {
		// 生成不存在的元素
		key := fmt.Sprintf("nonexistent%d", i)
		if bf.Contains([]byte(key)) {
			falsePositive++
		}
	}
	
	actualRate := float64(falsePositive) / float64(tests)
	t.Logf("期望误判率: %.4f, 实际误判率: %.4f, 误判数: %d/%d", p, actualRate, falsePositive, tests)
	
	// 允许实际误判率略高于期望值（测试允许一定的误差）
	if actualRate > p*3 {
		t.Errorf("实际误判率 %.4f 远高于期望值 %.4f", actualRate, p)
	}
}

// BenchmarkAdd 测试添加性能
func BenchmarkAdd(b *testing.B) {
	bf := NewBloomFilter(100000, 0.01)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data := []byte(fmt.Sprintf("item%d", i))
		bf.Add(data)
	}
}

// BenchmarkContains 测试查找性能
func BenchmarkContains(b *testing.B) {
	bf := NewBloomFilter(100000, 0.01)
	
	// 预先添加一些数据
	for i := 0; i < 10000; i++ {
		bf.Add([]byte(fmt.Sprintf("item%d", i)))
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data := []byte(fmt.Sprintf("item%d", rand.Intn(20000)))
		bf.Contains(data)
	}
}

// TestClear 测试清空功能
func TestClear(t *testing.T) {
	bf := NewBloomFilter(100, 0.01)
	
	// 添加元素
	bf.Add([]byte("test"))
	if !bf.Contains([]byte("test")) {
		t.Error("布隆过滤器应该包含 test")
	}
	
	// 清空
	bf.Clear()
	
	// 验证元素不存在
	if bf.Contains([]byte("test")) {
		t.Error("清空后布隆过滤器不应该包含任何元素")
	}
}
