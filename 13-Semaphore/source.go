// source: src/cmd/vendor/golang.org/x/sync/semaphore/semaphore.go

// 信号量 P V 伪代码表示如下，中括号代表原子操作
function V(semaphore S, integer I):
    [S ← S + I]

function P(semaphore S, integer I):
    repeat:
        [if S ≥ I:
        S ← S − I
        break]

// Go 内部使用信号量来控制 goroutine 的阻塞和唤醒。
// 我们在学习基本并发原语的实现时也看到了，比如互斥锁的第二个字段：
type Mutex struct {
	state int32
	sema  uint32
}

type waiter struct {
    n     int64
    ready chan<- struct{} // Closed when semaphore acquired.
}

func NewWeighted(n int64) *Weighted {
    w := &Weighted{size: n}
    return w
}

/*
Go 扩展库中的信号量是使用互斥锁 +List 实现的。互斥锁实现其它字段的保护，
而 List 实现了一个等待队列，等待者的通知是通过 Channel 的通知机制实现的。
*/
type Weighted struct {
    size    int64         // 最大资源数
    cur     int64         // 当前已被使用的资源
    mu      sync.Mutex    // 互斥锁，对字段的保护
    waiters list.List     // 等待队列
}


// 在信号量的几个实现方法里，Acquire 是代码最复杂的一个方法，
// 它不仅仅要监控资源是否可用，而且还要检测 Context 的 Done 是否已关闭。
func (s *Weighted) Acquire(ctx context.Context, n int64) error {
    s.mu.Lock()
        // fast path, 如果有足够的资源，都不考虑ctx.Done的状态，将cur加上n就返回
    if s.size-s.cur >= n && s.waiters.Len() == 0 {
      s.cur += n
      s.mu.Unlock()
      return nil
    }
        // 如果是不可能完成的任务，请求的资源数大于能提供的最大的资源数
    if n > s.size {
      s.mu.Unlock()
            // 依赖ctx的状态返回，否则一直等待
      <-ctx.Done()
      return ctx.Err()
    }
        // 否则就需要把调用者加入到等待队列中
        // 创建了一个ready chan,以便被通知唤醒
    ready := make(chan struct{})
    w := waiter{n: n, ready: ready}
    elem := s.waiters.PushBack(w)
    s.mu.Unlock()

        // 等待
    select {
    case <-ctx.Done(): // context的Done被关闭
      err := ctx.Err()
      s.mu.Lock()
      select {
      case <-ready: // 如果被唤醒了，忽略ctx的状态
        err = nil
      default: 通知waiter
        isFront := s.waiters.Front() == elem
        s.waiters.Remove(elem)
        // 通知其它的waiters,检查是否有足够的资源
        if isFront && s.size > s.cur {
          s.notifyWaiters()
        }
      }
      s.mu.Unlock()
      return err
    case <-ready: // 被唤醒了
      return nil
    }
  }


// Release 方法将当前计数值减去释放的资源数 n，并唤醒等待队列中的调用者，看是否有足够的资源被获取。
func (s *Weighted) Release(n int64) {
    s.mu.Lock()
    s.cur -= n
    if s.cur < 0 {
      s.mu.Unlock()
      panic("semaphore: released more than held")
    }
    s.notifyWaiters()
    s.mu.Unlock()
}

// 一个可用资源数量的判断，数量够用表示成功返回 true ，否则 false，此方法并不会进行阻塞，而是直接返回。
func (s *Weighted) TryAcquire(n int64) bool {
    s.mu.Lock()
    success := s.size-s.cur >= n && s.waiters.Len() == 0
    if success {
        s.cur += n
    }
    s.mu.Unlock()
    return success
}

// 通知机制
// 通过 for 循环从链表头部开始头部依次遍历出链表中的所有waiter，
// 并更新计数器 Weighted.cur，同时将其从链表中删除，直到遇到 空闲资源数量 < watier.n 为止。
func (s *Weighted) notifyWaiters() {
    for {
      next := s.waiters.Front()
      if next == nil {
        break // No more waiters blocked.
      }

      w := next.Value.(waiter)
      if s.size-s.cur < w.n {
        //避免饥饿，这里还是按照先入先出的方式处理
        break
      }

      s.cur += w.n
      s.waiters.Remove(next)
      close(w.ready)
    }
  }