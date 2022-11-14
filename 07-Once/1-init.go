// 1. 比如定义 package 级别的变量，这样程序在启动的时候就可以初始化：

package abc

import time

var startTime = time.Now()

// ---------------------------------

// 2. 或者在 init 函数中进行初始化：
package abc

var startTime time.Time

func init() {
  startTime = time.Now()
}

// ---------------------------------
// 3. 或者在 main 函数开始执行的时候，执行一个初始化的函数
package abc

var startTime time.Tim

func initApp() {
    startTime = time.Now()
}
func main() {
  initApp()
}

// 这三种方法都是线程安全的，并且后两种方法还可以根据传入的参数实现定制化的初始化操作。