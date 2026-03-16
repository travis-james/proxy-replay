package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/travis-james/proxy-replay/internal/server"
	"github.com/travis-james/proxy-replay/internal/storage"
)

func main() {
	dir := flag.String("dir", "./recordings", "directory for storing recordings")
	port := flag.String("port", ":8080", "server listen address")
	flag.Parse()

	store := storage.FileStorage{
		Dir: *dir,
	}

	srv := server.New(store)

	log.Printf("proxy-replay listening on %s (directory=%s)\n", *port, *dir)
	log.Fatal(http.ListenAndServe(*port, srv))
}
