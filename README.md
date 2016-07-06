# mmapd
In memory datastore written in Go that uses mmap for persistence.

## Usage
### Set a value
Make a `POST` request to `/set/key` with the value as a parameter.

### Get a value
Make a `GET` request to `/get/key`
