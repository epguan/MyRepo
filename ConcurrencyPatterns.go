package main

import (
	"fmt"
	"net/http"
	"sync"
)

type Result struct {
	Error    error
	Response *http.Response
}

var wg sync.WaitGroup

func main() {
	checkStatus := func(done <-chan interface{}, urls ...string) <-chan Result {
		results := make(chan Result)
		go func() {           // 此处协程的作用是，迅速返回 result，防止程序阻塞，便于后续处理结果
			defer close(results)
			for _, url := range urls {
				wg.Add(1)
				go func(url string) {
					defer wg.Done()
					var result Result
					resp, err := http.Get(url)
					result = Result{Error: err, Response: resp}
					select {
					case <-done:
						return
					case results <- result:
					}
				}(url)
			}
			wg.Wait()
		}()
		return results
	}

	done := make(chan interface{})
	defer close(done)

	urls := []string{"https://www.baidu.com", "https://badhost"}
	for result := range checkStatus(done, urls...) {
		if result.Error != nil {
			fmt.Printf("error: %v \n", result.Error)
			continue
		}
		fmt.Printf("Response: %v\n", result.Response.Status)
	}
}
