// source: src/sync/pool.go

type Pool struct {
	noCopy noCopy

	local     unsafe.Pointer // 每个 P 的本地队列，实际类型为 [P]poolLocal
	localSize uintptr        // [P]poolLocal 的大小

	//  Victim Cache 本来是计算机架构里面的一个概念，是CPU硬件处理缓存的一种技术，
	// sync.Pool引入的意图在于降低GC压力的同时提高命中率。
	// 在一轮GC到来时，victim 和 victimSize 会分别接管 local 和 localSize，
	// victim机制在于减少GC冷启动导致的性能抖动让分配对象更加平滑。
	victim     unsafe.Pointer
	victimSize uintptr

	// 自定义的对象创建回调函数，当 pool 中无可用对象时会调用此函数
	New func() any  // interface{}
}

// ================================================================
// 垃圾回收时 sync.Pool 的处理逻辑：
func poolCleanup() {
	// 丢弃当前victim, STW所以不用加锁
	for _, p := range oldPools {
		p.victim = nil
		p.victimSize = 0
	}

	// 将local复制给victim, 并将原local置为nil
	for _, p := range allPools {
		p.victim = p.local
		p.victimSize = p.localSize
		p.local = nil
		p.localSize = 0
	}

	oldPools, allPools = allPools, nil
}

type poolLocalInternal struct {
	private any       // P 的私有缓存区，使用时不需要加锁.
	shared  poolChain // 公共缓存区。本地 P 可以 pushHead/popHead; 其它 P 只能 popTail.
}

/* Analysis：
你需要关注一下 local 字段，因为所有当前主要的空闲可用的元素都存放在 local 字段中，
请求元素时也是优先从 local 字段中查找可用的元素。
local 字段包含一个 poolLocalInternal 字段，并提供 CPU 缓存对齐，从而避免 false sharing。

poolLocalInternal 也包含两个字段：private 和 shared。
- private，代表一个缓存的元素，而且只能由相应的一个 P 存取。
  因为一个 P 同时只能执行一个 goroutine，所以不会有并发的问题。
- shared，可以由任意的 P 访问，但是只有本地的 P 才能 pushHead/popHead，
  其它 P 可以 popTail，相当于只有一个本地的 P 作为生产者（Producer），
  多个 P 作为消费者（Consumer），它是使用一个 local-free 的 queue 列表实现的。

*/

// =================================================================
// Get 方法
func (p *Pool) Get() interface{} {
	// 把当前goroutine固定在当前的P上
	l, pid := p.pin()
	x := l.private // 优先从local的private字段取，快速
	l.private = nil
	if x == nil {
		// 从当前的local.shared弹出一个，注意是从head读取并移除
		x, _ = l.shared.popHead()
		if x == nil { // 如果没有，则去偷一个
			x = p.getSlow(pid)
		}
	}
	runtime_procUnpin()
	// 如果没有获取到，尝试使用New函数生成一个新的
	if x == nil && p.New != nil {
		x = p.New()
	}
	return x
}

func (p *Pool) getSlow(pid int) interface{} {

	size := atomic.LoadUintptr(&p.localSize)
	locals := p.local
	// 从其它proc中尝试偷取一个元素
	for i := 0; i < int(size); i++ {
		l := indexLocal(locals, (pid+i+1)%int(size))
		if x, _ := l.shared.popTail(); x != nil {
			return x
		}
	}

	// 如果其它proc也没有可用元素，那么尝试从victim中获取
	size = atomic.LoadUintptr(&p.victimSize)
	if uintptr(pid) >= size {
		return nil
	}
	locals = p.victim
	l := indexLocal(locals, pid)
	if x := l.private; x != nil { // 同样的逻辑，先从victim中的local private获取
		l.private = nil
		return x
	}
	for i := 0; i < int(size); i++ { // 从victim其它proc尝试偷取
		l := indexLocal(locals, (pid+i)%int(size))
		if x, _ := l.shared.popTail(); x != nil {
			return x
		}
	}

	// 如果victim中都没有，则把这个victim标记为空，以后的查找可以快速跳过了
	atomic.StoreUintptr(&p.victimSize, 0)

	return nil
}

/* Analysis：
1）首先，调用 p.pin()函数将当前的 goroutine 和 P 绑定，禁止被抢占，返回当前 P对应的poolLocal 以及pid.
2）然后直接取 l.private， 赋值给 x，并置 l.private 为 nil。
3）判断x 是否为空，若为空，则尝试从 l.shared 的头部 pop一个对象出来，同时赋值给 x。
4）如果×仍然为空，则调用 getslow 尝试从其他 P 的shared 双端队列尾部“偷”一个对象出来。
5）Pool 的相关操作做完了，调用runtime_procUnpin()解除非抢占。
6) 最后如果还是没有取到缓存对象，那就直接调用预先设置好的New函数，创建一个出来。
*/

// =================================================================
// pin
// 调用方必须完成取值后调用 runtime_procUnpin() 来取消抢占
func (p *Pool) pin() (*poolLocal, int) {
	pid := runtime_procPin()

	s := runtime_LoadAcquintptr(&p.localSize) // load-acquire
	l := p.local                              // load-consume
	// 因为可能存在动态的 P （运行时调整 p 的个数）
	if uintptr(pid) < s {
		return indexLocal(l, pid), pid
	}
	return p.pinSlow()
}

// ====================================================================
// Put
func (p *Pool) Put(x interface{}) {
	if x == nil { // nil值直接丢弃
		return
	}
	l, _ := p.pin()
	if l.private == nil { // 如果本地private没有值，直接设置这个值即可
		l.private = x
		x = nil
	}
	if x != nil { // 否则加入到本地队列中
		l.shared.pushHead(x)
	}
	runtime_procUnpin()
}

/* Analysis：
Put 的逻辑相对简单，优先设置本地 private，如果 private 字段已经有值了，
那么就把此元素 push 到本地队列中。
*/