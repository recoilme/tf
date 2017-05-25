package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func worker(id int, jobs <-chan string, results chan<- string) {

	for j := range jobs {

		fmt.Println("worker", id, "processing job", j)

		var res = doJob(j)

		//rate limits wor workers
		time.Sleep(time.Duration(1) * time.Second)

		if len(res) > 10 {
			res = res[:10]
		}
		results <- res
	}

}

func doJob(url string) string {

	// ... выполняем что-нибудь
	var b []byte
	if resp, err := http.Get(url); err == nil {
		b, err = ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		//_ = b
		//log.Println(string(b))
	}
	return string(b)
}

func main() {
	log.Println("Start")

	urls := []string{
		"http://code.jquery.com/jquery-1.9.1.min.js",
		"asd",
		"http://ajax.aspnetcdn.com/ajax/jQuery/jquery-1.9.1.min.js",
		"111",
		"http://cdnjs.cloudflare.com/ajax/libs/jquery/1.9.1/jquery.min.js",
		"http://ajax.googleapis.com/ajax/libs/jquery/1.9.1/jquery.min.js",
		"[htym",
		"http://example.ru",
		"http://www.example34.org/",
	}

	jobs := make(chan string, 10)
	results := make(chan string, 10)

	for w := 1; w <= 3; w++ {
		go worker(w, jobs, results)
	}

	for _, url := range urls {
		jobs <- url
	}
	close(jobs)
	log.Println("jobs send")
	for r := 0; r < len(urls); r++ {
		res := <-results
		fmt.Println("finished with res:", res)
	}
	close(results)
	time.Sleep(time.Duration(10) * time.Second)
	log.Println("Done")
}

/*
Output
2017/05/25 19:24:25 Start
2017/05/25 19:24:25 jobs send
worker 3 processing job http://code.jquery.com/jquery-1.9.1.min.js
worker 1 processing job asd
worker 2 processing job http://ajax.aspnetcdn.com/ajax/jQuery/jquery-1.9.1.min.js
worker 1 processing job 111
finished with res: asd
finished with res: http://ajax.aspnetcdn.com/ajax/jQuery/jquery-1.9.1.min.js
worker 2 processing job http://cdnjs.cloudflare.com/ajax/libs/jquery/1.9.1/jquery.min.js
worker 3 processing job http://ajax.googleapis.com/ajax/libs/jquery/1.9.1/jquery.min.js
finished with res: http://code.jquery.com/jquery-1.9.1.min.js
worker 1 processing job [htym
finished with res: 111
worker 2 processing job weqwe
finished with res: http://cdnjs.cloudflare.com/ajax/libs/jquery/1.9.1/jquery.min.js
finished with res: http://ajax.googleapis.com/ajax/libs/jquery/1.9.1/jquery.min.js
finished with res: [htym
finished with res: weqwe
2017/05/25 19:24:42 Done
*/
