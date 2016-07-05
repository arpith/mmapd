package main

import (
	"net/http"
	"os"
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
	router := httprouter.New()
	router.HandlerFunc("POST", "/set/:key", setHandler())
	router.HandlerFunc("GET", "/get/:key", getHandler())

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = 3001
	}
	http.ListenAndServe(":"+port, router)
}
