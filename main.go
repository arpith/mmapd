package main

import (
	"github.com/arpith/mmapd/db"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"os"
	"strings"
)

func main() {
	dbFilename := "../db.json"
	logFilename := "../log.json"
	DB := db.Init(dbFilename, logFilename)
	handler := db.NewHandler(DB)

	router := httprouter.New()
	router.GET("/get/:key", handler)
	router.POST("/set/:key", handler)

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "3001"
	}

	http.ListenAndServe(":"+port, router)
}
