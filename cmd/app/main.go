package main

import "github.com/travis-james/proxy-replay/internal/recorder"

func main() {
	url := "https://jsonplaceholder.typicode.com/todos/1"
	recorder.Record(url)
}
