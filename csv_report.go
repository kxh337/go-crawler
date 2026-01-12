package main

import (
	"encoding/csv"
	"os"
	"strings"
)

func writeCSVReport(pages map[string]PageData, filename string) error{
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	writer := csv.NewWriter(file)

	// Write Header row
	err = writer.Write([]string{
		"page_url",
		"h1",
		"first_paragraph",
		"outgoing_link_urls",
		"image_urls",
	})
	if err != nil {
		return err
	}

	for url, pageData := range pages {
		err = writer.Write([]string{
			url,
			pageData.H1,
			pageData.FirstParagraph,
			strings.Join(pageData.OutgoingLinks, ";"),
			strings.Join(pageData.ImageURLs, ";"),
		})
		if err != nil {
			return err
		}
	}
	return nil
}
