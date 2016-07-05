package main

import (
	"net/http"
	"os"
	"fmt"
	"strings"

	"github.com/julienschmidt/httprouter"
	"golang.org/x/exp/mmap"
)

func setHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "Setting %s!", ps.ByName("key"))
}

func getHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "Getting %s!", ps.ByName("key"))
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
	router.POST("/set/:key", setHandler)
	router.GET("/get/:key", getHandler)

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "3001"
	}
	http.ListenAndServe(":"+port, router)
}
