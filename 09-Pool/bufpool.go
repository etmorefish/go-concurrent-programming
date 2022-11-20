/* bufpool: https://github.com/gohugoio/hugo/blob/master/bufferpool/bufpool.go
著名的静态网站生成工具 Hugo 

 */

var buffers = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func GetBuffer() *bytes.Buffer {
	return buffers.Get().(*bytes.Buffer)
}

func PutBuffer(buf *bytes.Buffer) {
	buf.Reset()
	buffers.Put(buf)
}