package singleflight

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// expensiveOperation 模拟一个耗时的操作,比如从数据库查询数据
func expensiveOperation(key string) (interface{}, error) {
	fmt.Printf("[%s] 开始执行耗时操作...\n", key)
	time.Sleep(500 * time.Millisecond)
	return fmt.Sprintf("结果-%s", key), nil
}

// ExampleUsage 基本使用示例
func ExampleUsage() {
	var g Group
	var wg sync.WaitGroup
	var counter int32

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			val, err := g.Do("same-key", func() (interface{}, error) {
				atomic.AddInt32(&counter, 1)
				return expensiveOperation("same-key")
			})

			if err != nil {
				fmt.Printf("错误: %v\n", err)
				return
			}
			fmt.Printf("获得结果: %v\n", val)
		}()
	}

	wg.Wait()
	fmt.Printf("\n实际执行的函数次数: %d (应该是 1)\n", counter)
}

// ExampleMultipleKeys 演示不同 key 的请求
func ExampleMultipleKeys() {
	var g Group
	var wg sync.WaitGroup
	var counter int32

	keys := []string{"key1", "key2", "key3", "key1", "key2", "key1"}

	for _, key := range keys {
		wg.Add(1)
		go func(k string) {
			defer wg.Done()
			val, err := g.Do(k, func() (interface{}, error) {
				atomic.AddInt32(&counter, 1)
				return expensiveOperation(k)
			})

			if err != nil {
				fmt.Printf("[%s] 错误: %v\n", k, err)
				return
			}
			fmt.Printf("[%s] 获得结果: %v\n", k, val)
		}(key)
	}

	wg.Wait()
	fmt.Printf("\n实际执行的函数次数: %d (应该是 3)\n", counter)
}

// ExampleDoChan 使用 DoChan 的示例
func ExampleDoChan() {
	var g Group
	var wg sync.WaitGroup
	var counter int32

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ch := g.DoChan("chan-key", func() (interface{}, error) {
				atomic.AddInt32(&counter, 1)
				return expensiveOperation("chan-key")
			})

			res := <-ch
			if res.Err != nil {
				fmt.Printf("错误: %v\n", res.Err)
				return
			}
			fmt.Printf("获得结果: %v, 共享: %v\n", res.Val, res.Shared)
		}()
	}

	wg.Wait()
	fmt.Printf("\n实际执行的函数次数: %d (应该是 1)\n", counter)
}

// ExampleForget 演示 Forget 的使用
func ExampleForget() {
	var g Group
	var wg sync.WaitGroup
	var counter int32

	// 第一次请求
	wg.Add(1)
	go func() {
		defer wg.Done()
		val, err := g.Do("forget-key", func() (interface{}, error) {
			atomic.AddInt32(&counter, 1)
			time.Sleep(100 * time.Millisecond)
			return "第一次结果", nil
		})
		fmt.Printf("第一次请求结果: %v, 错误: %v\n", val, err)
	}()

	// 等待一小段时间后取消
	time.Sleep(50 * time.Millisecond)
	g.Forget("forget-key")
	fmt.Println("已调用 Forget,取消缓存")

	// 第二次请求,应该重新执行
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(50 * time.Millisecond)
		val, err := g.Do("forget-key", func() (interface{}, error) {
			atomic.AddInt32(&counter, 1)
			return "第二次结果", nil
		})
		fmt.Printf("第二次请求结果: %v, 错误: %v\n", val, err)
	}()

	wg.Wait()
	fmt.Printf("\n实际执行的函数次数: %d (应该是 2)\n", counter)
}

// ExampleErrorHandling 错误处理示例
func ExampleErrorHandling() {
	var g Group
	var wg sync.WaitGroup

	failureCount := 0

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := g.Do("error-key", func() (interface{}, error) {
				time.Sleep(100 * time.Millisecond)
				return nil, fmt.Errorf("模拟错误")
			})

			if err != nil {
				failureCount++
				fmt.Printf("请求失败: %v\n", err)
			}
		}()
	}

	wg.Wait()
	fmt.Printf("\n失败的请求数: %d (所有请求共享同一个错误)\n", failureCount)
}

// BenchmarkSingleflight 性能基准测试
func BenchmarkSingleflight(b *testing.B) {
	var g Group
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		g.Do("benchmark-key", func() (interface{}, error) {
			time.Sleep(10 * time.Microsecond)
			return "result", nil
		})
	}
}

// BenchmarkConcurrent 并发性能测试
func BenchmarkConcurrent(b *testing.B) {
	var g Group
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			g.Do("concurrent-key", func() (interface{}, error) {
				time.Sleep(1 * time.Microsecond)
				return "result", nil
			})
		}
	})
}

// TestConcurrency 并发安全性测试
func TestConcurrency(t *testing.T) {
	var g Group
	var wg sync.WaitGroup
	var counter int32

	// 100 个并发请求
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = g.Do("test-key", func() (interface{}, error) {
				atomic.AddInt32(&counter, 1)
				time.Sleep(10 * time.Millisecond)
				return "result", nil
			})
		}()
	}

	wg.Wait()

	if counter != 1 {
		t.Errorf("期望执行 1 次,实际执行 %d 次", counter)
	}
}

// TestMultipleKeys 测试多个 key
func TestMultipleKeys(t *testing.T) {
	var g Group
	var wg sync.WaitGroup
	mu := sync.Mutex{}
	results := make(map[string]int)

	keys := []string{"a", "b", "c", "a", "b", "a", "c", "b", "a"}

	for _, key := range keys {
		wg.Add(1)
		go func(k string) {
			defer wg.Done()
			g.Do(k, func() (interface{}, error) {
				time.Sleep(50 * time.Millisecond)
				return k, nil
			})

			mu.Lock()
			results[k]++
			mu.Unlock()
		}(key)
	}

	wg.Wait()

	expected := map[string]int{"a": 4, "b": 3, "c": 2}
	for k, v := range expected {
		if results[k] != v {
			t.Errorf("key %s: 期望 %d 次调用,实际 %d 次", k, v, results[k])
		}
	}
}
