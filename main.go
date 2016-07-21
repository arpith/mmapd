package main

import (
	"flag"
	"github.com/arpith/mmapd/db"
	"github.com/arpith/mmapd/raft"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"os"
	"strings"
)

func main() {
	dbFilename := "db.json"
	logFilename := "log.json"
	configFilename := "config.json"
	DB := db.Init(dbFilename, logFilename)
	server := raft.Init("10.0.17.176", configFilename, DB)
	appendEntryHandler := raft.NewHandler(server, "Append Entry")
	requestForVoteHandler := raft.NewHandler(server, "Request For Vote")
	clientRequestHandler := raft.NewHandler(server, "Client Request")

	router := httprouter.New()
	router.POST("/append", appendEntryHandler)
	router.POST("/votes", requestForVoteHandler)
	router.GET("/get/:key", clientRequestHandler)
	router.POST("/set/:key", clientRequestHandler)

	portPtr := flag.String("port", "3001", "port to listen on")
	flag.Parse()
	port := *portPtr
	if port == "3001" {
		envPort := strings.TrimSpace(os.Getenv("PORT"))
		if envPort != "" {
			port = envPort
		}
	}

	http.ListenAndServe(":"+port, router)
}
