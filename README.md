# mmapd
Distributed key value datastore written in Go. Uses the Raft distributed consensus algorithm to replicate log, and uses mmap for persistence.

## Should I use it?
Not yet! This project is under development, and you WILL lose data. 

### What works?
Currently, you can get three nodes running (tested on a single machine) and set values on the leader - they will be replicated (and persisted) in the followers' logs and dbs. You can also crash the leader and continue using the followers.

### What doesn't work?
Lots of stuff! But primarily, if you re-start the former leader, it WILL try to regain leadership, and won't accept the entries it missed!

### Why is this interesting?
My primary goal is to explore the Raft consensus algorithm and mmap. They are both quite cool! Follow my explorations on [Medium](medium.com/@arpith)!

## Under the hood
### Database
`/db` has the files that handle persistence. The basic idea is that the database and log files are memory mapped into a byte slice, which is then parsed as JSON into a map of keys -> values. When the files need to grow, `ftruncate` is used to extend the underlying file, which is then re-mapped into memory. Finally the new byte slice is copied into the mmaped byte slice, which the OS then copies onto the file eventually. Going forward, this is going to be forced using `msync`.

### Consensus
`/raft` has my implementation of the Raft consensus algorithm. When a leader receives a request from the client, it sends an append entry request to the followers. The followers then add this to their log and reply if successful. If a majority of the followers successfully replicate this entry onto their log, the leader commits the entry (to the database) and responds to the client. On the next heartbeat, the leader includes the new commit index and the followers commit the latest entry.

If a follower doesn't receive a heartbeat before its election timeout goes off, it sends vote requests to all the other nodes. Nodes respond with success to the first vote request they receive (and reject the others). If a candidate receives successful responses from a majority of the nodes, it promotes itself to leader. This is also the protocol for the initial election.

## Usage
### Set a value
Make a `POST` request to `/set/key` with the value as a parameter.

### Get a value
Make a `GET` request to `/get/key`

### Get node status
Make a `GET` request to `/status`

## Install
`go get github.com/arpith/mmapd`

## Run
`mmapd -port PORT -db DATABASE_FILENAME -log LOG_FILENAME`

### -port
Set the port you want this node to be listening on. The default value is `3001`

### -db
The filename used to store the database (for persistence). The default value is `db.json`

### -log
The filename used to store the log (for persistence). The default value is `log.json`

## Example
```
$ mmapd -port 3001 -db db3001.json -log log3001.json \
   && mmapd -port 3002 -db db3002.json -log log3002.json \
   && mmapd -port 3003 -db db3003.json -log log3003.json
```
