package main

import (
	"net/http"
	"os"
	"fmt"
	"strings"

	"github.com/julienschmidt/httprouter"
	"golang.org/x/exp/mmap"
)

func setHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Setting %s!", r.URL.Path[1:])
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Getting %s!", r.URL.Path[1:])
}

func main() {
	if _, err := os.Stat("db"); os.IsNotExist(err) {
		os.Create("db")
	}

	db, err := mmap.Open("db")
	if err != nil {
		fmt.Print("Error memory mapping db file: ", err)
	}

	len := db.Len

	fmt.Printf("Length of memory mapped db file is: %d", len)

	router := httprouter.New()
	router.HandlerFunc("POST", "/set/:key", setHandler)
	router.HandlerFunc("GET", "/get/:key", getHandler)

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "3001"
	}
	http.ListenAndServe(":"+port, router)
}
