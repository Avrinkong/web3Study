package main

import "fmt"

//题目 ：编写一个程序，使用通道实现两个协程之间的通信。一个协程生成从1到10的整数，并将这些整数发送到通道中，另一个协程从通道中接收这些整数并打印出来。
//考察点 ：通道的基本使用、协程间通信。

func setNumber(ch chan<- int) {
	go func() {
		for i := 1; i <= 10; i++ {
			ch <- i
			fmt.Println("生产：", i)
		}
		close(ch)
	}()
}

func getNumber(ch <-chan int) {
	for i := range ch {
		fmt.Println("消费：", i)
	}
}

func main7() {
	ch := make(chan int, 10)
	setNumber(ch)
	getNumber(ch)
}

// 题目 ：实现一个带有缓冲的通道，生产者协程向通道中发送100个整数，消费者协程从通道中接收这些整数并打印。
// 考察点 ：通道的缓冲机制。
func setNumber1(ch chan<- int) {
	go func() {
		for i := 1; i <= 100; i++ {
			ch <- i
			fmt.Println("生产：", i)
		}
		close(ch)
	}()
}

func getNumber1(ch <-chan int) {
	for i := range ch {
		fmt.Println("消费：", i)
	}
}

func main8() {
	ch := make(chan int, 10)
	setNumber1(ch)
	getNumber1(ch)
}
