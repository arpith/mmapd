package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"testing"
)

func TestConcurrency(t *testing.T) {
	for i := 0; i < 146; i++ {
		s := strconv.Itoa(i)
		fmt.Println("Setting", s)
		_, err := http.PostForm("http://localhost:3001/set/"+s, url.Values{"value": {s}})
		if err != nil {
			fmt.Println("Error setting ", s, ": ", err)
		}
	}
}
