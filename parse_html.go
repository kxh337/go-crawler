package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type PageData struct {
	URL string
	H1 string
	FirstParagraph string
	OutgoingLinks []string
	ImageURLs []string
}

func crawlPage(rawBaseURL, rawCurrentURL string, pages map[string]int) {
	if !strings.HasPrefix(rawCurrentURL, rawBaseURL) {
		return
	}
	normUrl, err := normalizeURL(rawCurrentURL)
	if err != nil {
		fmt.Println("Failed to normalize the URL")
		fmt.Printf("Error: %v\n", err)
		return
	}
	_, ok := pages[normUrl]
	if ok {
		pages[normUrl]++
		return
	} else {
		pages[normUrl] = 1
	}
	html, err := getHTML(rawCurrentURL)
	if err != nil {
		fmt.Printf("Failed to parse html in page: %v\n", rawCurrentURL)
		fmt.Printf("Error: %v\n", err)
		return
	}

	pageData := extractPageData(html, rawBaseURL)
	fmt.Println("Found the following links:")
	for _, l := range pageData.OutgoingLinks {
		fmt.Printf("%v\n",l)
	}

	for _, outgoingUrl := range pageData.OutgoingLinks {
		nl, err:= normalizeURL(outgoingUrl)
		if err != nil{
			fmt.Printf("Failed to normalize url: %v\n", outgoingUrl)
			return
		}
		_, ok := pages[nl]
		if !ok  {
			time.Sleep(5 * time.Second)
			fmt.Printf("Crawling next link: %v\n", outgoingUrl)
			crawlPage(rawBaseURL, outgoingUrl, pages)
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

func extractPageData(html, pageURL string) PageData {
	var res PageData
	res.URL = pageURL
	res.H1 = getH1FromHTML(html)
	res.FirstParagraph = getFirstParagraphFromHTML(html)
	baseURL, err := url.Parse(pageURL)
	if err != nil {
		log.Fatal("Failed to parse URL")
	}

	res.OutgoingLinks, err = getURLsFromHTML(html, baseURL)
	if err != nil {
		log.Fatal("Failed to get URLs from HTML")
	}

	res.ImageURLs, err = getImagesFromHTML(html, baseURL)
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
