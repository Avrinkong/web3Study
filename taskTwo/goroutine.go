package main

import (
	"fmt"
	"sync"
	"time"
)

//题目 ：编写一个程序，使用 go 关键字启动两个协程，一个协程打印从1到10的奇数，另一个协程打印从2到10的偶数。
//考察点 ： go 关键字的使用、协程的并发执行。

func main3() {
	var wg sync.WaitGroup

	// 添加两个任务到等待组
	wg.Add(2)
	go printOddNumbers(&wg)
	go printEvenNumbers(&wg)
	wg.Wait()
}
func printOddNumbers(wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 1; i <= 10; i += 2 {
		fmt.Println("奇数：", i)
		time.Sleep(time.Millisecond * 500)
	}
}

func printEvenNumbers(wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 2; i <= 10; i += 2 {
		fmt.Println("偶数：", i)
		time.Sleep(time.Millisecond * 500)
	}
}

//题目 ：设计一个任务调度器，接收一组任务（可以用函数表示），并使用协程并发执行这些任务，同时统计每个任务的执行时间。
//考察点 ：协程原理、并发任务调度。

func main4() {
	var wg sync.WaitGroup
	tasks := []func(){
		func() {
			fmt.Println("任务1")
			time.Sleep(time.Second * 3)
		},
		func() {
			fmt.Println("任务2")
			time.Sleep(time.Second * 1)
		},
		func() {
			fmt.Println("任务3")
			time.Sleep(time.Second * 2)
		},
	}
	// 设置等待组计数器
	wg.Add(len(tasks))
	for _, task := range tasks {
		go func(t func()) {
			// 确保在协程结束时减少等待组计数
			defer wg.Done()
			start := time.Now()
			t()
			elapsed := time.Since(start)
			fmt.Printf("任务执行完成，耗时：%s\n", elapsed)
		}(task)
	}
	// 等待所有协程完成
	wg.Wait()
}
