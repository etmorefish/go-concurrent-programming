package main

import "context"

// 在使用 ErrGroup 时，我们要用到三个方法，分别是 WithContext、Go 和 Wait。

// 1.WithContext
// 在创建一个 Group 对象时，需要使用 WithContext 方法：
func WithContext(ctx context.Context) (*Group, context.Context)

// 这个方法返回一个 Group 实例，同时还会返回一个使用 context.WithCancel(ctx) 生成的新 Context。
// 一旦有一个子任务返回错误，或者是 Wait 调用返回，这个新 Context 就会被 cancel。

// 2.Go
// 执行子任务的 Go 方法：
func (g *Group) Go(f func() error)

// 传入的子任务函数 f 是类型为 func() error 的函数，
// 如果任务执行成功，就返回 nil，否则就返回 error，并且会 cancel 那个新的 Context。

// 3.Wait
// 类似 WaitGroup，Group 也有 Wait 方法，等所有的子任务都完成后，它才会返回，否则只会阻塞等待。
// 如果有多个子任务返回错误，它只会返回第一个出现的错误，如果所有的子任务都执行成功，就返回 nil：

func (g *Group) Wait() error

// ----------------------------------------------------------------
// gollback

// 1.All
// 方法All 方法的签名如下：
func All(ctx context.Context, fns ...AsyncFunc) ([]interface{}, []error)

// 它会等待所有的异步函数（AsyncFunc）都执行完才返回，而且返回结果的顺序和传入的函数的顺序保持一致。
// 第一个返回参数是子任务的执行结果，第二个参数是子任务执行时的错误信息。其中，异步函数的定义如下：
type AsyncFunc func(ctx context.Context) (interface{}, error)

// 可以看到，ctx 会被传递给子任务。如果你 cancel 这个 ctx，可以取消子任务。
// 例：example_all.go

// 2.Race
// 方法Race 方法跟 All 方法类似，只不过，在使用 Race 方法的时候，只要一个异步函数执行没有错误，就立马返回，
// 而不会返回所有的子任务信息。如果所有的子任务都没有成功，就会返回最后一个 error 信息。
// Race 方法签名如下：

func Race(ctx context.Context, fns ...AsyncFunc) (interface{}, error)

// 如果有一个正常的子任务的结果返回，Race 会把传入到其它子任务的 Context cancel 掉，这样子任务就可以中断自己的执行。
// Race 的使用方法也跟 All 方法类似，可以把 All 方法的例子中的 All 替换成 Race 方式。

// 3.Retry
// 方法Retry 不是执行一组子任务，而是执行一个子任务。如果子任务执行失败，它会尝试一定的次数，
// 如果一直不成功 ，就会返回失败错误 ，如果执行成功，它会立即返回。如果 retires 等于 0，它会永远尝试，直到成功。

func Retry(ctx context.Context, retires int, fn AsyncFunc) (interface{}, error)

// 例：example_retry.go

// ---------------------------------------------------------------
// Hunch
// Hunch提供的功能和 gollback 类似，不过它提供的方法更多，而且它提供的和 gollback 相应的方法，也有一些不同。
// 它定义了执行子任务的函数，这和 gollback 的 AyncFunc 是一样的，
// 它的定义如下：

type Executable func(context.Context) (interface{}, error)

// 1.All
// 方法All 方法的签名如下：

func All(parentCtx context.Context, execs ...Executable) ([]interface{}, error)

// 它会传入一组可执行的函数（子任务），返回子任务的执行结果。
// 和 gollback 的 All 方法不一样的是，一旦一个子任务出现错误，它就会返回错误信息，执行结果（第一个返回参数）为 nil。

// 2.Take 方法Take 方法的签名如下：

func Take(parentCtx context.Context, num int, execs ...Executable) ([]interface{}, error)

// 你可以指定 num 参数，只要有 num 个子任务正常执行完没有错误，这个方法就会返回这几个子任务的结果。
// 一旦一个子任务出现错误，它就会返回错误信息，执行结果（第一个返回参数）为 nil。

// 3.Last
// 方法Last 方法的签名如下：

func Last(parentCtx context.Context, num int, execs ...Executable) ([]interface{}, error)

// 它只返回最后 num 个正常执行的、没有错误的子任务的结果。
// 一旦一个子任务出现错误，它就会返回错误信息，执行结果（第一个返回参数）为 nil。
// 比如 num 等于 1，那么，它只会返回最后一个无错的子任务的结果。

//	4.Retry
// 方法Retry 方法的签名如下：

func Retry(parentCtx context.Context, retries int, fn Executable) (interface{}, error)

// 它的功能和 gollback 的 Retry 方法的功能一样，如果子任务执行出错，就会不断尝试，直到成功或者是达到重试上限。
// 如果达到重试上限，就会返回错误。如果 retries 等于 0，它会不断尝试。

// 5.Waterfall
// 方法Waterfall 方法签名如下：

func Waterfall(parentCtx context.Context, execs ...ExecutableInSequence) (interface{}, error)

// 它其实是一个 pipeline 的处理方式，所有的子任务都是串行执行的，
// 前一个子任务的执行结果会被当作参数传给下一个子任务，
// 直到所有的任务都完成，返回最后的执行结果。
// 一旦一个子任务出现错误，它就会返回错误信息，执行结果（第一个返回参数）为 nil。
// gollback 和 Hunch 是属于同一类的并发原语，对一组子任务的执行结果，
// 可以选择一个结果或者多个结果，这也是现在热门的微服务常用的服务治理的方法。

// ----------------------------------------------------------------
// schedgroup
// 一个和时间相关的处理一组 goroutine 的并发原语 schedgroup。
// 这个并发原语包含的方法如下：

type Group
  func New(ctx context.Context) *Group
  func (g *Group) Delay(delay time.Duration, fn func())
  func (g *Group) Schedule(when time.Time, fn func())
  func (g *Group) Wait() error

// 1 2.Delay 和 Schedule 方法。
// 它们的功能其实是一样的，都是用来指定在某个时间或者之后执行一个函数。
// 只不过，Delay 传入的是一个 time.Duration 参数，它会在 time.Now()+delay 之后执行函数，
// 而 Schedule 可以指定明确的某个时间执行。

// 3. Wait 方法。
// 这个方法调用会阻塞调用者，直到之前安排的所有子任务都执行完才返回。
// 如果 Context 被取消，那么，Wait 方法会返回这个 cancel error。
// 需要两点注意：
	// 第一点是，如果调用了 Wait 方法，你就不能再调用它的 Delay 和 Schedule 方法，否则会 panic。
	// 第二点是，Wait 方法只能调用一次，如果多次调用的话，就会 panic。