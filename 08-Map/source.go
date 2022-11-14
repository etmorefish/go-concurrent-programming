package main

import (
	"sync/atomic"
	"unsafe"
)

func main() {
	// 源码地址：src/sync/map.go
}

type Map struct {
	mu Mutex
	// 基本上你可以把它看成一个安全的只读的map
	// 它包含的元素其实也是通过原子操作更新的，但是已删除的entry就需要加锁操作了
	read atomic.Value // readOnly

	// 包含需要加锁才能访问的元素
	// 包括所有在read字段中但未被expunged（删除）的元素以及新加的元素
	dirty map[interface{}]*entry

	// 记录从read中读取miss的次数，一旦miss数和dirty长度一样了，就会把dirty提升为read，并把dirty置空
	misses int
}

type readOnly struct {
	m       map[interface{}]*entry
	amended bool // 当dirty中包含read没有的数据时为true，比如新增一条数据
}

// expunged是用来标识此项已经删掉的指针
// 当map中的一个项目被删除了，只是把它的值标记为expunged，以后才有机会真正删除此项
var expunged = unsafe.Pointer(new(interface{}))

// entry代表一个值
type entry struct {
	p unsafe.Pointer // *interface{}
}

/*Analysis:
如果 dirty 字段非 nil 的话，map 的 read 字段和 dirty 字段会
包含相同的非 expunged 的项，所以如果通过 read 字段更改了这个项
的值，从 dirty 字段中也会读取到这个项的新值，因为本来它们指向的
就是同一个地址。
dirty 包含重复项目的好处就是，一旦 miss 数达到阈值需要将 dirty
提升为 read 的话，只需简单地把 dirty 设置为 read 对象即可。不
好的一点就是，当创建新的 dirty 对象的时候，需要逐条遍历 read，
把非 expunged 的项复制到 dirty 对象中。
*/

// =================================================================
func (m *Map) Store(key, value interface{}) {
	read, _ := m.read.Load().(readOnly)
	// 如果read字段包含这个项，说明是更新，cas更新项目的值即可
	if e, ok := read.m[key]; ok && e.tryStore(&value) {
		return
	}

	// read中不存在，或者cas更新失败，就需要加锁访问dirty了
	m.mu.Lock()
	read, _ = m.read.Load().(readOnly)
	if e, ok := read.m[key]; ok { // 双检查，看看read是否已经存在了
		if e.unexpungeLocked() {
			// 此项目先前已经被删除了，通过将它的值设置为nil，标记为unexpunged
			m.dirty[key] = e
		}
		e.storeLocked(&value) // 更新
	} else if e, ok := m.dirty[key]; ok { // 如果dirty中有此项
		e.storeLocked(&value) // 直接更新
	} else { // 否则就是一个新的key
		if !read.amended { //如果dirty为nil
			// 需要创建dirty对象，并且标记read的amended为true,
			// 说明有元素它不包含而dirty包含
			m.dirtyLocked()
			m.read.Store(readOnly{m: read.m, amended: true})
		}
		m.dirty[key] = newEntry(value) //将新值增加到dirty对象中
	}
	m.mu.Unlock()
}

func (m *Map) dirtyLocked() {
	if m.dirty != nil { // 如果dirty字段已经存在，不需要创建了
		return
	}

	read, _ := m.read.Load().(readOnly) // 获取read字段
	m.dirty = make(map[interface{}]*entry, len(read.m))
	for k, e := range read.m { // 遍历read字段
		if !e.tryExpungeLocked() { // 把非punged的键值对复制到dirty中
			m.dirty[k] = e
		}
	}
}

/*Analysis:
Store方法有多条路径，
第一条，如果read中存在，直接更新read和dirty（他们的key共享的value都是同一个entry，所以更新read会把dirty也更新）
第二条，如果这个key在read中存在并且之前这个key已经被删除了（expunged），那么就将他设置为nil表示未删除然后把这个nil替换成要保存的值 （read和dirty同时修改）
第三条，如果在dirty中存在 ，直接修改（这条路径其实就是 dirty有 但是read没有）
第四条，是新增key

所以从这么来看，sync.Map 适合那些只会增长的缓存系统，可以进行更新，但是不要删除，并且不要频繁地增加新元素。
新加的元素需要放入到 dirty 中，如果 dirty 为 nil，那么需要从 read 字段中复制出来一个 dirty 对象
*/

// =================================================================

func (m *Map) Load(key interface{}) (value interface{}, ok bool) {
	// 首先从read处理
	read, _ := m.read.Load().(readOnly)
	e, ok := read.m[key]
	if !ok && read.amended { // 如果不存在并且dirty不为nil(有新的元素)
		m.mu.Lock()
		// 双检查，看看read中现在是否存在此key
		read, _ = m.read.Load().(readOnly)
		e, ok = read.m[key]
		if !ok && read.amended { //依然不存在，并且dirty不为nil
			e, ok = m.dirty[key] // 从dirty中读取
			// 不管dirty中存不存在，miss数都加1
			m.missLocked()
		}
		m.mu.Unlock()
	}
	if !ok {
		return nil, false
	}
	return e.load() //返回读取的对象，e既可能是从read中获得的，也可能是从dirty中获得的
}

func (m *Map) missLocked() {
	m.misses++                   // misses计数加一
	if m.misses < len(m.dirty) { // 如果没达到阈值(dirty字段的长度),返回
		return
	}
	m.read.Store(readOnly{m: m.dirty}) //把dirty字段的内存提升为read字段
	m.dirty = nil                      // 清空dirty
	m.misses = 0                       // misses数重置为0
}

/*Analysis：
Load 方法用来读取一个 key 对应的值。它也是从 read 开始处理，一开始并不需要锁。
如果幸运的话，我们从 read 中读取到了这个 key 对应的值，那么就不需要加锁了，性能会非常好。
但是，如果请求的 key 不存在或者是新加的，就需要加锁从 dirty 中读取。
所以，读取不存在的 key 会因为加锁而导致性能下降，读取还没有提升的新值的情况下也会因为加锁性能下降。

missLocked 增加 miss 的时候，如果 miss 数等于 dirty 长度，会将 dirty 提升为 read，并将 dirty 置空。
*/

// =================================================================

func (m *Map) LoadAndDelete(key interface{}) (value interface{}, loaded bool) {
	read, _ := m.read.Load().(readOnly)
	e, ok := read.m[key]
	if !ok && read.amended {
		m.mu.Lock()
		// 双检查
		read, _ = m.read.Load().(readOnly)
		e, ok = read.m[key]
		if !ok && read.amended {
			e, ok = m.dirty[key]
			// 这一行长坤在1.15中实现的时候忘记加上了，导致在特殊的场景下有些key总是没有被回收
			delete(m.dirty, key)
			// miss数加1
			m.missLocked()
		}
		m.mu.Unlock()
	}
	if ok {
		return e.delete()
	}
	return nil, false
}

func (m *Map) Delete(key interface{}) {
	m.LoadAndDelete(key)
}
func (e *entry) delete() (value interface{}, ok bool) {
	for {
		p := atomic.LoadPointer(&e.p)
		if p == nil || p == expunged {
			return nil, false
		}
		if atomic.CompareAndSwapPointer(&e.p, p, nil) { //如果read中存在当前key，那么获取e之后调用delete的行为是。将read中的这个e设置为nil
			return *(*interface{})(p), true
		}
	}
}

/* Analysis:
Delete 方法也是先从 read 操作开始，原因我们已经知道了，因为不需要锁。

如果 read 中不存在，那么就需要从 dirty 中寻找这个项目。
最终，如果项目存在就删除（将它的值标记为 nil）。
如果项目不为 nil 或者没有被标记为 expunged，那么还可以把它的值返回。
*/
