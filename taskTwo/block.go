package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

//题目 ：编写一个程序，使用 sync.Mutex 来保护一个共享的计数器。启动10个协程，每个协程对计数器进行1000次递增操作，最后输出计数器的值。
//考察点 ： sync.Mutex 的使用、并发数据安全。

func main9() {
	count := 0
	var mutex sync.Mutex
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 1000; j++ {
				mutex.Lock()
				count++
				mutex.Unlock()
			}
		}()

	}
	// 该等待时间，可以在去掉锁之后，发现值并非10000
	time.Sleep(time.Second)
	fmt.Println("计数器值：", count)
}

//题目 ：使用原子操作（ sync/atomic 包）实现一个无锁的计数器。启动10个协程，每个协程对计数器进行1000次递增操作，最后输出计数器的值。
//考察点 ：原子操作、并发数据安全。

func main() {
	var count int32
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 1000; j++ {
				atomic.AddInt32(&count, 1)
			}
		}()
	}

	time.Sleep(time.Second)
	fmt.Println("计数器值：", count)
}
