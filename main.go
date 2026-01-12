package main

import (
	"fmt"
	"os"
	"sync"
	"net/url"
	"strconv"
)

func main() {
	args := os.Args[1:]
	baseUrl := ""
	maxThreadCount := 5
	maxPageCount := 10

	switch len(args) {
	case 1:
		baseUrl = string(args[0])
	case 2:
		baseUrl = string(args[0])
		arg1, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Printf("Error encounterd: %v\n", err)
			os.Exit(1)
		}
		maxThreadCount = arg1

	case 3:
		baseUrl = args[0]
		arg1, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Printf("Error encounterd: %v\n", err)
			os.Exit(1)
		}
		maxThreadCount = arg1
		arg2, err := strconv.Atoi(args[2])
		if err != nil {
			fmt.Printf("Error encounterd: %v\n", err)
			os.Exit(1)
		}
		maxPageCount = arg2

	default:
		fmt.Println("Pass atleast one arg")
		os.Exit(1)
	}

	u, err := url.Parse(args[0])
	if err != nil {
		fmt.Printf("Bad URL was inputted")
		os.Exit(1)
	}
	fmt.Printf("starting crawl of: %v\n", baseUrl)
	cfg := config{
		pages : make(map[string]PageData),
		baseURL:  u,
		mu : &sync.Mutex{},
		concurrencyControl : make(chan struct{}, maxThreadCount),
		wg : &sync.WaitGroup{},
		maxPages: maxPageCount,
	}
	cfg.crawlPage(baseUrl)
	cfg.wg.Wait()
	err = writeCSVReport(cfg.pages, "./report.csv")
	if err != nil {
		fmt.Printf("Error encounterd: %v\n", err)
		os.Exit(1)
	}
}
