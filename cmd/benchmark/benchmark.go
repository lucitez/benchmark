package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/lucitez/benchmark/benchmarker"
)

const URL = "https://go.dev"

// start tcp server
// handle connection with one message type of benchmark
// call benchmark and send each one back
func main() {
	listener, err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	fmt.Println("Listening on port 8000")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Connection closed")
			os.Exit(1)
		}

		fmt.Println("Accepted connection")

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')
		switch {
		case err == io.EOF:
			fmt.Println("Client connection terminated")
			return
		case err != nil:
			fmt.Printf("An error occurred while reading from client %v\n", err)
			return
		}

		msgType, msgVal := extractMsg(message)

		// Then handle the message according to its type
		switch msgType {
		case "benchmark":
			handleBenchmark(conn, msgVal)
		case "message":
			log.Println("Received message: " + msgVal)
		default:
			conn.Write([]byte("error;valid message types are 'benchmark', 'message'\n"))
		}
	}
}

func handleBenchmark(conn net.Conn, url string) {
	b := benchmarker.New(url)

	startmsg := "message;Benchmarking " + url + "\n"
	log.Print(startmsg)
	conn.Write([]byte(startmsg))

	b.BenchmarkWebsite(func(p benchmarker.Performance) {
		asJson, _ := json.Marshal(p)
		topipe := "url_performance;" + string(asJson) + "\n"
		conn.Write([]byte(topipe))
	})

	endmsg := "Benchmarking complete\n"
	log.Print(endmsg)
	conn.Write([]byte(endmsg))
}

func extractMsg(msg string) (string, string) {
	splitMsg := strings.Split(msg, ";")

	msgType := splitMsg[0]

	msgVal := ""

	if len(splitMsg) > 1 {
		msgVal = splitMsg[1]
	}

	return strings.TrimSpace(msgType), strings.TrimSpace(msgVal)
}
