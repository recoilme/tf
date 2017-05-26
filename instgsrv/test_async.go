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
2017/05/25 20:09:10 Start
2017/05/25 20:09:10 jobs send
worker 3 processing job http://code.jquery.com/jquery-1.9.1.min.js
worker 1 processing job asd
worker 2 processing job http://ajax.aspnetcdn.com/ajax/jQuery/jquery-1.9.1.min.js
worker 1 processing job 111
finished with res:
worker 2 processing job http://cdnjs.cloudflare.com/ajax/libs/jquery/1.9.1/jquery.min.js
finished with res: /*! jQuery
worker 3 processing job http://ajax.googleapis.com/ajax/libs/jquery/1.9.1/jquery.min.js
finished with res: /*! jQuery
worker 1 processing job [htym
finished with res:
worker 2 processing job http://example.ru
finished with res: /*! jQuery
worker 3 processing job http://www.example34.org/
finished with res: /*! jQuery
finished with res:
finished with res: <!DOCTYPE
finished with res:
2017/05/25 20:09:23 Done
*/
