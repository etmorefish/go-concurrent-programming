// source: src/sync/atomic/doc.go

// Add
// atomic 的 Add 是针对 int 和 uint 进行原子加值的：
func AddInt32(addr *int32, delta int32) (new int32)
func AddUint32(addr *uint32, delta uint32) (new uint32)
func AddInt64(addr *int64, delta int64) (new int64)
func AddUint64(addr *uint64, delta uint64) (new uint64)
func AddUintptr(addr *uintptr, delta uintptr) (new uintptr)


// CompareAndSwap
// 比较并交换方法实现了类似乐观锁的功能，只有原来的值和传入的 old 值一样，才会去修改：
// CompareAndSwap 有可能产生 ABA 现象发生。也就是原来的值是 A，后面被修改 B，
// 再后面修改为 A。在这种情况下也符合了 CompareAndSwap 规则，即使中途有被改动过。
func CompareAndSwapInt32(addr *int32, old, new int32) (swapped bool)
func CompareAndSwapInt64(addr *int64, old, new int64) (swapped bool)
func CompareAndSwapUint32(addr *uint32, old, new uint32) (swapped bool)
func CompareAndSwapUint64(addr *uint64, old, new uint64) (swapped bool)
func CompareAndSwapUintptr(addr *uintptr, old, new uintptr) (swapped bool)
func CompareAndSwapPointer(addr *unsafe.Pointer, old, new unsafe.Pointer) (swapped bool)

// Load
// Load 方法是为了防止在读取过程中，有其他协程发起修改动作，影响了读取结果，常用于配置项的整个读取。
func LoadInt32(addr *int32) (val int32)
func LoadInt64(addr *int64) (val int64)
func LoadUint32(addr *uint32) (val uint32)
func LoadUint64(addr *uint64) (val uint64)
func LoadUintptr(addr *uintptr) (val uintptr)
func LoadPointer(addr *unsafe.Pointer) (val unsafe.Pointer)

// Store
// 有原子读取，就有原子修改值，前面提到过的 Add 只适用于 int、uint 类型的增减，
// 并没有其他类型的修改，而 Sotre 方法通过 unsafe.Pointer 指针原子修改，来达到了对其他类型的修改。
func StoreInt32(addr *int32, val int32)
func StoreInt64(addr *int64, val int64)
func StoreUint32(addr *uint32, val uint32)
func StoreUint64(addr *uint64, val uint64)
func StoreUintptr(addr *uintptr, val uintptr)
func StorePointer(addr *unsafe.Pointer, val unsafe.Pointer)


// Swap
// Swap 方法实现了对值的原子交换，不仅 int，uint 可以交换，指针也可以。

func SwapInt32(addr *int32, new int32) (old int32)
func SwapInt64(addr *int64, new int64) (old int64)
func SwapUint32(addr *uint32, new uint32) (old uint32)
func SwapUint64(addr *uint64, new uint64) (old uint64)
func SwapUintptr(addr *uintptr, new uintptr) (old uintptr)
func SwapPointer(addr *unsafe.Pointer, new unsafe.Pointer) (old unsafe.Pointer)
