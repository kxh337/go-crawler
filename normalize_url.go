package main

import (
	"net/url"
	"fmt"
	"strings"
)

func normalizeURL(input string) (string, error){
	u,  err := url.Parse(input)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(fmt.Sprintf("%v%v", u.Host, u.Path), "/"), nil
}
