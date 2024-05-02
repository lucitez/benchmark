package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/lucitez/benchmark/benchmark"
)

func main() {
	listener, err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	fmt.Println("Listening on :8000")

	s := http.Server{
		Handler: benchmark.Server{},
	}

	s.Serve(listener)
}
