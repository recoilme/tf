// We can also use this syntax to iterate over
// values received from a channel.

package main

import "fmt"
import "time"
import "log"

var (
	queue = make(chan string, 2)
)

func writer(start int) {
	for i := start; i < start+4; i++ {
		queue <- fmt.Sprintf("%d", i)
	}
}

func reader() {
	for s := range queue {
		log.Println(s)
		time.Sleep(1 * time.Second)
	}

}

func main() {
	go reader()
	go writer(1)
	go writer(100)
	go writer(10)
	go writer(200)
	time.Sleep(40 * time.Second)
}
