package benchmark

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"nhooyr.io/websocket"
)

// https://github.com/nhooyr/websocket/tree/master/internal/examples/echo

type Server struct{}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	acceptOpts := websocket.AcceptOptions{
		InsecureSkipVerify: true,
	}
	conn, err := websocket.Accept(w, r, &acceptOpts)
	if err != nil {
		log.Printf("failed to accept websocket: %v\n", err)
		return
	}
	log.Printf("received connection")

	defer conn.CloseNow()

	for {
		err = handleConnection(r.Context(), conn)
		switch {
		case err == io.EOF:
			fmt.Println("Reached EOF")
			continue
		case err == http.ErrServerClosed:
			fmt.Println("Connection terminated")
			return
		case websocket.CloseStatus(err) == websocket.StatusNormalClosure:
			fmt.Println("Connection terminated")
			return
		case websocket.CloseStatus(err) == websocket.StatusGoingAway:
			fmt.Println("Connection terminated")
			return
		case err != nil:
			fmt.Printf("Failed to process: %v", err)
			return
		}
	}
}

func handleConnection(ctx context.Context, conn *websocket.Conn) error {
	_, reader, err := conn.Reader(ctx)
	if err != nil {
		return err
	}

	br := bufio.NewReader(reader)

	msg, err := br.ReadString('\n')
	if err != nil && err != io.EOF {
		return err
	}

	fmt.Printf("RECEIVED: %s\n", msg)

	msgType, msgVal := extractMsg(msg)

	// Then handle the message according to its type
	switch msgType {
	case "benchmark":
		err = handleBenchmark(ctx, conn, msgVal)
	case "message":
		log.Println("Received message: " + msgVal)
	case "echo":
		err = write(ctx, conn, msgVal)
	default:
		err = write(ctx, conn, "error;valid message types are 'benchmark', 'message'\n")
	}

	return err
}

func write(ctx context.Context, conn *websocket.Conn, msg string) error {
	w, err := conn.Writer(ctx, websocket.MessageText)
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(msg))
	if err != nil {
		return err
	}

	return w.Close()
}

func handleBenchmark(ctx context.Context, conn *websocket.Conn, rootUrl string) error {
	var err error

	startmsg := "message;Benchmarking " + rootUrl + "\n"
	log.Print(startmsg)
	err = write(ctx, conn, startmsg)
	if err != nil {
		return err
	}

	urls := make(chan string)
	benchmarks := make(chan Benchmark)

	go benchmarkWebsite(rootUrl, urls, benchmarks)

	for url := range urls {
		write(ctx, conn, "url;"+url)
	}

	for benchmark := range benchmarks {
		asJson, err := json.Marshal(benchmark)
		if err != nil {
			fmt.Printf("Error marshalling benchmark %v\n", err)
		}
		topipe := "benchmark;" + string(asJson) + "\n"
		write(ctx, conn, topipe)
	}

	endmsg := "message;benchmarking_complete\n"
	log.Print(endmsg)
	err = write(ctx, conn, endmsg)
	return err
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
