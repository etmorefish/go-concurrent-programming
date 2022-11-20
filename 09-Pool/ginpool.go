/* source： https://github.com/gin-gonic/gin/blob/55e27f12465e058058180280d5f0bdc473eb3302/gin.go#L205
gin框架，对context的取用也使用了 sync.pool


*/

// 设置 New 函数
engine.pool.New = func() any {
	return engine.allocateContext(engine.maxParams)
}

func (engine *Engine) allocateContext(maxParams uint16) *Context {
	v := make(Params, 0, maxParams)
	skippedNodes := make([]skippedNode, 0, engine.maxSections)
	return &Context{engine: engine, params: &v, skippedNodes: &skippedNodes}
}

// 使用
// ServeHTTP conforms to the http.Handler interface.
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := engine.pool.Get().(*Context)
	c.writermem.reset(w)
	c.Request = req
	c.reset()

	engine.handleHTTPRequest(c)

	engine.pool.Put(c)
}

// 先调用 Get 取出来缓存的对象，然后会做一些reset操作，
// 再执行 handleHTTPRequest，最后再 Put 回 Pool。
// 另外，Echo框架也使用了 sync.Pool 来管理 context，并且几乎达到了零堆内存分配。