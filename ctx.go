package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

func main() {

	result := 0
	ctx, cancel := context.WithCancel(context.Background())
	chanQueue := make([]chan int, 3)
	defer cancel()

	for i := 0; i < 3; i++ {
		chanQueue[i] = make(chan int)

		if i == 2 {
			go func(i int) {
				chanQueue[i] <- 1
			}(i)
		}
	}
	fmt.Println("chan--------", chanQueue)

	exitChan := make(chan bool)

	for i := 0; i < 3; i++ {
		var lastChan chan int
		var curChan chan int

		if i == 0 {
			lastChan = chanQueue[2]
		} else {
			lastChan = chanQueue[i-1]
		}

		curChan = chanQueue[i]

		go func(i int, lastChan, curChan chan int) {
			ticker := time.NewTicker(1 * time.Second)
			for {
				select {
				case <-ctx.Done():
					fmt.Println("goroutine ", i, "exit")
					break
				case <-ticker.C:
					if result > 100 {

						exitChan <- true
					} else {
						r := rand.Intn(10)
						result += r
						fmt.Println("--goroutine ", i, result, r)

					}

				}
			}
		}(i, lastChan, curChan)
	}

	<-exitChan
	fmt.Println("--------------", result)
	cancel()
	time.Sleep(2 * time.Second)
}
