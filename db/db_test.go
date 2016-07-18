package db

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"testing"
)

func TestGet(t *testing.T) {
	fmt.Println("Testing GET")
	resp, err := http.Get("http://localhost:3001/get/testKey")
	if err != nil {
		t.Error("Couldn't get testKey: ", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error("Expected testValue got ", err)
	} else if string(body) != "testValue" {
		t.Error("Expected testValue got ", string(body))
	} else {
		fmt.Println("Test Passed!")
	}
}

func TestSet(t *testing.T) {
	fmt.Println("Testing SET")
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
	} else {
		fmt.Println("Test passed!")
	}
}

func TestSetAndGet(t *testing.T) {
	fmt.Println("Testing SET & GET")
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
		t.Error("Expected ", s, " got ", string(body))
	} else {
		fmt.Println("Set ", s)
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
		t.Error("Expected ", s, " got ", string(body))
	} else {
		fmt.Println("Test passed!")
	}
}

func TestConcurrency(t *testing.T) {
	fmt.Println("Testing Concurrency")
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
