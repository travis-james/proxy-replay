package main

import (
	"log"
	"net/http"

	"github.com/travis-james/proxy-replay/internal/server"
	"github.com/travis-james/proxy-replay/internal/storage"
)

func main() {

	store := storage.FileStorage{
		Dir: "./recordings",
	}

	srv := server.New(store)

	log.Println("proxy-replay listening on :8080")

	log.Fatal(http.ListenAndServe(":8080", srv))
}
