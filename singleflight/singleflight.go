package singleflight

import (
	"sync"
)

// Result 是 Do 方法返回的结果
type Result struct {
	Val    interface{}
	Err    error
	Shared bool // Shared 表示结果是否是共享的
}

// call 表示正在进行的请求
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// Group 管理共享相同 key 的请求
type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

// Do 执行函数 fn,确保对于给定的 key,只调用一次 fn
// 多个并发调用会等待第一个调用完成,然后共享结果
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}

	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()

	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}

// DoChan 类似于 Do,但返回一个 channel
func (g *Group) DoChan(key string, fn func() (interface{}, error)) <-chan Result {
	ch := make(chan Result, 1)
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}

	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		go func() {
			c.wg.Wait()
			ch <- Result{c.val, c.err, true}
			close(ch)
		}()
		return ch
	}

	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	go func() {
		c.val, c.err = fn()
		c.wg.Done()

		g.mu.Lock()
		delete(g.m, key)
		g.mu.Unlock()

		ch <- Result{c.val, c.err, false}
		close(ch)
	}()

	return ch
}

// Forget 用于主动取消某个 key 的等待
// 使得下一次 Do 调用会重新执行 fn
func (g *Group) Forget(key string) {
	g.mu.Lock()
	if c, ok := g.m[key]; ok {
		delete(g.m, key)
		c.wg.Done()
	}
	g.mu.Unlock()
}
