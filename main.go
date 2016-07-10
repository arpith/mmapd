package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"os"
	"strings"
)

func main() {
	filename := "db"
	db := openDB(filename)
	handler := NewHandler(db)

	router := httprouter.New()
	router.GET("/get/:key", handler)
	router.POST("/set/:key", handler)

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "3001"
	}

	http.ListenAndServe(":"+port, router)
}
