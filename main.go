package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

func main() {
	var (
		wordlistPath string
		speed        int
		targetURL    string
	)

	flag.StringVar(&wordlistPath, "txt", "", "Wordlist dosya yolu")
	flag.IntVar(&speed, "speed", 10, "Fuzzing hızı")
	flag.StringVar(&targetURL, "url", "", "Fuzzing yapılacak hedef URL")
	flag.Parse()

	wordlist, err := readWordlist(wordlistPath)
	if err != nil {
		fmt.Println("Wordlist okuma hatası:", err)
		return
	}

	results := make(chan string)
	var wg sync.WaitGroup
	visitedURLs := make(map[string]bool)
	var mu sync.Mutex

	for i := 0; i < speed; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for _, word := range wordlist {
				url := fmt.Sprintf("%s/%s", targetURL, word)
				mu.Lock()
				if !visitedURLs[url] {
					visitedURLs[url] = true
					mu.Unlock()
					if checkURL(url) {
						results <- url
					}
				} else {
					mu.Unlock()
				}
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	fmt.Println("Sonuçlar:")
	count := 1
	for result := range results {
		fmt.Printf("%d - %s\n", count, result)
		count++
	}
}

func readWordlist(filePath string) ([]string, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	lines := make([]string, 0)
	for _, line := range strings.Split(string(content), "\n") {
		lines = append(lines, line)
	}

	return lines, nil
}

func checkURL(url string) bool {
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
