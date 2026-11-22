
package main

import (
	"fmt"
	"sync"
)

type Fetcher interface {
	Fetch(url string) (body string, urls []string, err error)
}

func crawlHelper(url string, depth int,f Fetcher, m map[string]bool, mu *sync.Mutex, wg *sync.WaitGroup){
	defer wg.Done()
	if depth <= 0{
		return
	}
	mu.Lock()
	alreadyVisited := m[url]
	if alreadyVisited {
		mu.Unlock()
		return 
	}
    m[url] = true
	mu.Unlock()
	body, urls, err := f.Fetch(url)
	 if err != nil {
		 fmt.Println(err)
        return
    }
	fmt.Printf("found: %s %q\n", url, body)
	
	for _, u := range urls {
    wg.Add(1)

	go crawlHelper(u, depth-1, f, m, mu, wg)
	}
}
	
func Crawl(url string, depth int, f Fetcher) {
	visited := make(map[string]bool)
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(1)
	crawlHelper(url, depth, f, visited, &mu, &wg)
	wg.Wait()
}


func main() {
	var mu sync.Mutex
	var wg sync.WaitGroup
	url := "golang.org"
	visited := make(map[string]bool)
	for i :=0; i<5; i++{
		wg.Add(1)
		go visit(url,visited, &mu, &wg)
	}
	wg.Wait()

}
