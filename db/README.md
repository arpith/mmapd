# Data
Primarily, data is stored in-memory and persisted to disk by the OS. This is done by memory mapping the log and database files into Go byte slices. The byte slices are marshalled/unmarshalled into structures using json. When a request is made a lookup is done. The database is a map of strings - key to values; the log is a slice of structs with
where the lookup is typically based on index.

Since these structures are not concurrency safe, a single listener goroutine takes ownership of the writes/reads and all other goroutines communicate via channels to send their requests to the listener. The messages contain a response channel over which the listener will send the response. This lets the goroutine that makes the read/write request wait on a single channel for the response.

## Tell me more!
### mmap
Mmap is a system call that reads the content of a file into a byte slice in Go. Changes made to the byte slice are synced to disk by the OS. In future, this will be forced by the database. Stay tuned!

### ftruncate
Ftruncate is a system call that resizes a file - this allows the database to grow the file used for persistence. Currently, the strategy is to resize and remap the file everytime it needs to grow, but this can be optimized by doubling each time instead.

### Data structures
#### Database
The database is primarily a map of strings, keys to values. There is also data stored like the file descriptor, etc, for resizing, and the actual byte slice that contains the data.

#### Log
The log is a slice of entries where each entry has a command (for now, that is just "SET"), a key, value and term (used by Raft). Again, the file descriptor etc is also stored, and the byte slice.

### JSON
When the database starts up, the db/log files are read into byte slices which are then unmarshalled into the appropriate structs. When writes are made to the structs they are marshalled into the memory-mapped byte slices. This is then eventually synced to disk.

## Concurrency
The strategy is to have a single goroutine that is responsible for the data structures which then listens on channels for read/write requests. These requests come from multiple goroutines that the Raft server sets up, so they will be coming in concurrently. Each request (which is a channel message) also contains a return channel on which the requester goroutine will be waiting on. The listener processes the request and sends the response over the return channel. The requester goroutine can then close the channel.
