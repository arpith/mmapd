# mmapd
Distributed key value datastore written in Go. Uses the Raft distributed consensus algorithm to replicate log, and uses mmap for persistence.

## Usage
### Set a value
Make a `POST` request to `/set/key` with the value as a parameter.

### Get a value
Make a `GET` request to `/get/key`

### Get node status
Make a `GET` request to `/status`
