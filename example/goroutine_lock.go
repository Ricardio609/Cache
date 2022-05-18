package main

import (
	"sync"
	"fmt"
	"time"
)

var m sync.Mutex
var set = make(map[int]bool, 0)

/* 避免重复打印 */
func printOnce(num int)  {					//使用set记录已打印的数字，如果数字已打印，则不再打印
	m.Lock()
	defer	m.Unlock()						//unlcok第二种写法
	if _, exist := set[num]; !exist {
		fmt.Println(num)
	}
	set[num] = true
	// m.Unlock()							//unlcok第一种写法
}

func main() {
	for i := 0; i < 10; i++ {				//10个并发协程
		go printOnce(100)
	}
	time.Sleep(time.Second)
}