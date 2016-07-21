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
	dbFilenamePtr := flag.String("db", "db.json", "database filename")
	logFilenamePtr := flag.String("log", "log.json", "log filename")
	configFilenamePtr := flag.String("config", "config.json", "config file name")
	portPtr := flag.String("port", "3001", "port to listen on")
	ipPtr := flag.String("ip", "localhost", "ip that the server is running on")
	flag.Parse()

	port := *portPtr
	if port == "3001" {
		envPort := strings.TrimSpace(os.Getenv("PORT"))
		if envPort != "" {
			port = envPort
		}
	}

	DB := db.Init(*dbFilenamePtr, *logFilenamePtr)
	server := raft.Init(*ipPtr+port, *configFilenamePtr, DB)
	appendEntryHandler := raft.NewHandler(server, "Append Entry")
	requestForVoteHandler := raft.NewHandler(server, "Request For Vote")
	clientRequestHandler := raft.NewHandler(server, "Client Request")

	router := httprouter.New()
	router.POST("/append", appendEntryHandler)
	router.POST("/votes", requestForVoteHandler)
	router.GET("/get/:key", clientRequestHandler)
	router.POST("/set/:key", clientRequestHandler)

	http.ListenAndServe(":"+port, router)
}
