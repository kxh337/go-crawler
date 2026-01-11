package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

type PageData struct {
	URL string
	H1 string
	FirstParagraph string
	OutgoingLinks []string
	ImageURLs []string
}

type config struct {
	pages              map[string]PageData
	baseURL            *url.URL
	mu                 *sync.Mutex
	concurrencyControl chan struct{}
	wg                 *sync.WaitGroup
	maxPages 	 	   int
}

func (cfg *config) crawlPage(rawCurrentURL string) {
	cfg.concurrencyControl <- struct{}{}
	cfg.wg.Add(1)
	defer func() {
		<- cfg.concurrencyControl
		cfg.wg.Done()
	} ()

	cfg.mu.Lock()
	if len(cfg.pages) >= cfg.maxPages {
		cfg.mu.Unlock()
		return
	}
	cfg.mu.Unlock()

	if !strings.HasPrefix(rawCurrentURL, cfg.baseURL.String()) {
		return
	}

	normUrl, err := normalizeURL(rawCurrentURL)
	if err != nil {
		fmt.Println("Failed to normalize the URL")
		fmt.Printf("Error: %v\n", err)
		return
	}

	cfg.mu.Lock()
	_, ok := cfg.pages[normUrl]
	cfg.mu.Unlock()

	if ok {
		return 
	} 

	html, err := getHTML(rawCurrentURL)
	if err != nil {
		fmt.Printf("Failed to parse html in page: %v\n", rawCurrentURL)
		fmt.Printf("Error: %v\n", err)
		return
	}

	pageData := cfg.extractPageData(html, rawCurrentURL)
	cfg.mu.Lock()
	if len(cfg.pages) >= cfg.maxPages {
		cfg.mu.Unlock()
		return
	}
	fmt.Printf("Added entry for %v\n", normUrl)
	cfg.pages[normUrl] = pageData
	cfg.mu.Unlock()

	for _, outgoingUrl := range cfg.pages[normUrl].OutgoingLinks {
		nl, err := normalizeURL(outgoingUrl)
		if err != nil{
			fmt.Printf("Failed to normalize url: %v\n", outgoingUrl)
			return
		}

		cfg.mu.Lock()
		_, ok := cfg.pages[nl]
		if len(cfg.pages) >= cfg.maxPages {
			cfg.mu.Unlock()
			return
		}
		cfg.mu.Unlock()

		if !ok  {
			time.Sleep(0 * time.Second)
			fmt.Printf("Crawling: %v\n", outgoingUrl)
			go cfg.crawlPage(outgoingUrl)
		} 
	}
}

func getHTML(rawURL string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", rawURL, nil)
	req.Header.Set("User-Agent", "BootCrawler/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return "", nil
	}
	if resp.StatusCode > 400 {
		return "", fmt.Errorf("Response had error status: %v\n", resp.StatusCode)
	}
	if !strings.ContainsAny(resp.Header.Get("Content-Type"),"text/html") {
		return "", fmt.Errorf("Response was the wrong content type: %v\n", resp.Header.Get("Content-type"))
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return string(body), err
}

func (cfg *config) extractPageData(html, pageURL string) PageData {
	var res PageData
	res.URL = pageURL
	res.H1 = getH1FromHTML(html)
	res.FirstParagraph = getFirstParagraphFromHTML(html)

	var err error
	res.OutgoingLinks, err = getURLsFromHTML(html, cfg.baseURL)
	if err != nil {
		log.Fatal("Failed to get URLs from HTML")
	}

	res.ImageURLs, err = getImagesFromHTML(html, cfg.baseURL)
	if err != nil {
		log.Fatal("Failed to parse images from HTML")
	}

	return res
}

func getURLsFromHTML(htmlBody string, baseURL *url.URL) ([]string, error) {
	res := []string{}
	reader := strings.NewReader(htmlBody)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return res, err
	}
	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		val, ok := s.Attr("href")
		if ok {
			if strings.HasPrefix(val, "/") {
				val = baseURL.String() + val
			}
			res = append(res, val)
		}
	})
	return res, nil
}

func getImagesFromHTML(htmlBody string, baseURL *url.URL) ([]string, error) {
	res := []string{}
	reader := strings.NewReader(htmlBody)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return res, err
	}
	doc.Find("img").Each(func(_ int, s *goquery.Selection) {
		val, ok := s.Attr("src")
		if ok {
			if strings.HasPrefix(val, "/") {
				val = baseURL.String() + val
			}
			res = append(res, val)
		}
	})
	return res, nil
}

func getH1FromHTML(html string) string {
	reader := strings.NewReader(html)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Fatal("Failed to parse html doc")
	}
	return doc.Find("h1").First().Text()
}

func getFirstParagraphFromHTML(html string) string {
	reader := strings.NewReader(html)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Fatal("Failed to parse html doc")
	}
	res := doc.Find("main p").First().Text()
	if res == "" {
		return doc.Find("p").First().Text() 
	}
	return res
}
