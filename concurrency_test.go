package main

import (
	"crypto/rand"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"testing"
)

func TestSet(t *testing.T) {
	c := 9
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		t.Error("Couldn't generate random string: ", err)
	}
	s := base64.URLEncoding.EncodeToString(b)
	resp, err := http.PostForm("http://localhost:3001/set/"+s, url.Values{"value": {s}})
	if err != nil {
		t.Error("Expected ", s, " got ", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error("Expected ", s, " got ", err)
	} else if string(body) != s {
		t.Error("Expected ", s, " got ", string(body))
	}
}

func TestGet(t *testing.T) {
	c := 9
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		t.Error("Couldn't generate random string: ", err)
	}
	s := base64.URLEncoding.EncodeToString(b)
	resp, err := http.PostForm("http://localhost:3001/set/"+s, url.Values{"value": {s}})
	if err != nil {
		t.Error("Couldn't set ", s, ": ", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error("Expected ", s, " got ", err)
	} else if string(body) != s {
		t.Error("Expected ", s, " got ", err)
	}
	resp, err = http.Get("http://localhost:3001/get/" + s)
	if err != nil {
		t.Error("Couldn't get ", s, ": ", err)
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error("Expected ", s, " got ", err)
	} else if string(body) != s {
		t.Error("Expected ", s, " got ", err)
	}
}

func TestConcurrency(t *testing.T) {
	maxRequests := os.Getenv("MAX_REQUESTS")
	max, err := strconv.Atoi(maxRequests)
	if err != nil {
		max = 250
	}
	for i := 0; i < max; i++ {
		s := strconv.Itoa(i)
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
