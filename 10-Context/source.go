package main

import (
	"errors"
	"fmt"
	"reflect"
	"time"
)

func main() {
	// 源码地址：src/go/types/context.go
}

// Context 的具体实现包括 4 个方法，分别是 Deadline、Done、Err 和 Value，如下所示：
type Context interface {
	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{}
	Err() error
	Value(key interface{}) interface{}
}

// =================================================================

// Context 中实现了 2 个常用的生成顶层 Context 的方法。
// context.Background() --- 可以直接使用
// context.TODO() --- 不知道传什么的时候可以传TODO
// 事实上，它们两个底层的实现是一模一样的：
var (
	background = new(emptyCtx)
	todo       = new(emptyCtx)
)

func Background() Context {
	return background
}

func TODO() Context {
	return todo
}

// =================================================================
// WithValue
func WithValue(parent Context, key, val interface{}) Context {
	if key == nil {
		panic("nil key")
	}
	if !reflect.TypeOf(key).Comparable() {
		panic("key is not comparable")
	}
	return &valueCtx{parent, key, val}
}

// WithValue 方法其实是创建了一个类型为 valueCtx 的 Context
type valueCtx struct {
	Context
	key, val interface{}
}

// 它实现了两个方法：
func (c *valueCtx) String() string {
	return fmt.Sprintf("%v.WithValue(%#v, %#v)", c.Context, c.key, c.val)
}
func (c *valueCtx) Value(key interface{}) interface{} {
	if c.key == key {
		return c.val
	}
	return c.Context.Value(key)
}

/* Analysis:
对 key 的要求是可比较，因为之后需要通过 key 取出 context 中的值，可比较是必须的。
通过层层传递 context，最终形成这样一棵树：

BackgroundContext <- Context[Key1, Val1] <- Context[Key2, Val2] <- Context[Key3, Val3]
							↑
							⏐
							⎣__________ Context[Key4, Val4] <- Context[Key5, Val5]

和链表有点像，只是它的方向相反：Context 指向它的父节点，链表则指向下一个节点。
通过 WithValue 函数，可以创建层层的 valueCtx，存储 goroutine 间可以共享的变量。

取值的过程，实际上是一个递归查找的过程
它会顺着链路一直往上找，比较当前节点的 key 是否是要找的 key，如果是，
则直接返回 value。否则，一直顺着 context 往前，最终找到根节点（一般是 emptyCtx），
直接返回一个 nil。所以用 Value 方法的时候要判断结果是否为 nil。

因为查找方向是往上走的，所以，父节点没法获取子节点存储的值，子节点却可以获取父节点的值。
*/

// =================================================================
// WithCancel

func WithCancel(parent Context) (ctx Context, cancel CancelFunc) {
	c := newCancelCtx(parent)
	propagateCancel(parent, &c) // 把c朝上传播
	return &c, func() { c.cancel(true, Canceled) }
}

// newCancelCtx returns an initialized cancelCtx.
func newCancelCtx(parent Context) cancelCtx {
	return cancelCtx{Context: parent}
}

/* Analysis：
这是一个暴露给用户的方法，传入一个父 Context（这通常是一个 background，作为根节点），
返回新建的 context，新 context 的 done channel 是新建的（前文讲过）。
当 WithCancel 函数返回的 CancelFunc 被调用或者是父节点的 done channel 被关闭
（父节点的 CancelFunc 被调用），此 context（子节点） 的 done channel 也会被关闭。
注意传给 WithCancel 方法的参数，前者是 true，也就是说取消的时候，需要将自己从父节点里删除。

*/
// cancelCtx 被取消时，它的 Err 字段就是下面这个 Canceled 错误：
var Canceled = errors.New("context canceled")

// =================================================================
// WithTimeout
// WithTimeout 其实是和 WithDeadline 一样，只不过一个参数是超时时间，一个参数是截止时间。
// 超时时间加上当前时间，其实就是截止时间，因此，WithTimeout 的实现是：
func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc) {
	// 当前时间+timeout就是deadline
	return WithDeadline(parent, time.Now().Add(timeout))
}

// =================================================================
// WithDeadline
func WithDeadline(parent Context, d time.Time) (Context, CancelFunc) {
	// 如果parent的截止时间更早，直接返回一个cancelCtx即可
	if cur, ok := parent.Deadline(); ok && cur.Before(d) {
		return WithCancel(parent)
	}
	c := &timerCtx{
		cancelCtx: newCancelCtx(parent),
		deadline:  d,
	}
	propagateCancel(parent, c) // 同cancelCtx的处理逻辑
	dur := time.Until(d)
	if dur <= 0 { //当前时间已经超过了截止时间，直接cancel
		c.cancel(true, DeadlineExceeded)
		return c, func() { c.cancel(false, Canceled) }
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.err == nil {
		// 设置一个定时器，到截止时间后取消
		c.timer = time.AfterFunc(dur, func() {
			c.cancel(true, DeadlineExceeded)
		})
	}
	return c, func() { c.cancel(true, Canceled) }
}

/*
WithDeadline 会返回一个 parent 的副本，并且设置了一个不晚于参数 d 的截止时间，
类型为 timerCtx（或者是 cancelCtx）。如果它的截止时间晚于 parent 的截止时间，
那么就以 parent 的截止时间为准，并返回一个类型为 cancelCtx 的 Context，
因为 parent 的截止时间到了，就会取消这个 cancelCtx。
如果当前时间已经超过了截止时间，就直接返回一个已经被 cancel 的 timerCtx。
否则就会启动一个定时器，到截止时间取消这个 timerCtx。

综合起来，timerCtx 的 Done 被 Close 掉，主要是由下面的某个事件触发的：
 截止时间到了；
 cancel 函数被调用；
 parent 的 Done 被 close。

和 cancelCtx 一样，WithDeadline（WithTimeout）返回的 cancel 一定要调用，
并且要尽可能早地被调用，这样才能尽早释放资源，不要单纯地依赖截止时间被动取消
*/

// =================================================================

func (c *cancelCtx) cancel(removeFromParent bool, err error) {
	// 必须要传 err
	if err == nil {
		panic("context: internal error: missing cancel error")
	}
	c.mu.Lock()
	if c.err != nil {
		c.mu.Unlock()
		return // 已经被其他协程取消
	}
	// 给 err 字段赋值
	c.err = err
	// 关闭 channel，通知其他协程
	if c.done == nil {
		c.done = closedchan
	} else {
		close(c.done)
	}

	// 遍历它的所有子节点
	for child := range c.children {
		// 递归地取消所有子节点
		child.cancel(false, err)
	}
	// 将子节点置空
	c.children = nil
	c.mu.Unlock()

	if removeFromParent {
		// 从父节点中移除自己
		removeChild(c.Context, c)
	}
}

/*
cancel() 方法的功能就是关闭 channel：c.done；递归地取消它的所有子节点；
从父节点从删除自己。达到的效果是通过关闭 channel，将取消信号传递给了它的所有子节点。
goroutine 接收到取消信号的方式就是 select 语句中的读 c.done 被选中。
*/

// =================================================================
func propagateCancel(parent Context, child canceler) {
	// 父节点是个空节点
	if parent.Done() == nil {
		return // parent is never canceled
	}
	// 找到可以取消的父 context
	if p, ok := parentCancelCtx(parent); ok {
		p.mu.Lock()
		if p.err != nil {
			// 父节点已经被取消了，本节点（子节点）也要取消
			child.cancel(false, p.err)
		} else {
			// 父节点未取消
			if p.children == nil {
				p.children = make(map[canceler]struct{})
			}
			// "挂到"父节点上
			p.children[child] = struct{}{}
		}
		p.mu.Unlock()
	} else {
		// 如果没有找到可取消的父 context。新启动一个协程监控父节点或子节点取消信号
		go func() {
			select {
			case <-parent.Done():
				child.cancel(false, parent.Err())
			case <-child.Done():
			}
		}()
	}
}

/*
	这个方法的作用就是向上寻找可以“挂靠”的“可取消”的 context，并且“挂靠”上去。这样，
	调用上层 cancel 方法的时候，就可以层层传递，将那些挂靠的子 context 同时“取消”。
*/
