package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"testing"
)

func TestConcurrency(t *testing.T) {
	maxRequests := os.Getenv("MAX_REQUESTS")
	max, err := strconv.Atoi(maxRequests)
	if err != nil {
		max = 250
	}
	for i := 0; i < max; i++ {
		s := strconv.Itoa(i)
		fmt.Println("Setting", s)
		resp, err := http.PostForm("http://localhost:3001/set/"+s, url.Values{"value": {s}})
		if err != nil {
			t.Error("Expected ", s, " got ", err)
		}
		defer resp.Body.Close()
	}
	for i := 0; i < max; i++ {
		s := strconv.Itoa(i)
		resp, err := http.Get("http://localhost:3001/get/" + s)
		if err != nil {
			t.Error("Expected ", s, " got ", err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil || string(body) != s {
			t.Error("Expected ", s, " got ", err)
		}
		defer resp.Body.Close()
	}
}
