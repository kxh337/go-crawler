package main

import (
	"log"
	"net/url"
	"strings"
	"github.com/PuerkitoBio/goquery"
)

type PageData struct {
	URL string
	H1 string
	FirstParagraph string
	OutgoingLinks []string
	ImageURLs []string
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
