package main

var done = make(chan bool)
var msg string

func aGoroutine() {
	// msg = "hello, world"
	done <- true
	msg = "hello, world"

}

func main() {
	go aGoroutine()
	<-done
	println(msg)
}

