package main

import (
	"net/url"
	"fmt"
)

func normalizeURL(input string) (string, error){
	u,  err := url.Parse(input)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v%v", u.Host, u.Path), nil
}
