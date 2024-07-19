package benchmark

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
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
		// continue listening to connected client
		case err == io.EOF:
			log.Println("Reached EOF")
			continue
		case err == http.ErrServerClosed:
			log.Printf("Closing server: %v\n", err)
			return
		case websocket.CloseStatus(err) == websocket.StatusNormalClosure:
			log.Printf("Connection terminated: %v\n", err)
			return
		case websocket.CloseStatus(err) == websocket.StatusGoingAway:
			log.Printf("Connection terminated: %v\n", err)
			return
		case err != nil:
			log.Printf("Unhandled error while handling websocket connection: %v\n", err)
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

	// read until the reader receives a termination
	msg, err := br.ReadString('\n')
	if err != nil && err != io.EOF {
		return err
	}

	log.Printf("Received message: %s\n", msg)

	msgType, msgVal, err := extractMsg(msg)
	if err != nil {
		return err
	}

	// Then handle the message according to its type
	switch msgType {
	case "benchmark":
		err = handleBenchmark(ctx, conn, msgVal)
		if err != nil {
			_ = sendStatus(ctx, conn, "error")
		}
	case "echo":
		err = write(ctx, conn, msgVal)
	default:
		err = write(ctx, conn, "error;valid message types are 'benchmark' and 'echo'\n")
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

func sendStatus(ctx context.Context, conn *websocket.Conn, status string) error {
	message := "status;" + status + "\n"
	log.Print(message)
	if err := write(ctx, conn, message); err != nil {
		return err
	}
	return nil
}

func extractMsg(msg string) (string, string, error) {
	splitMsg := strings.Split(msg, ";")

	if len(splitMsg) != 2 {
		return "", "", errors.New("invalid message format")
	}

	msgType := splitMsg[0]
	msgVal := splitMsg[1]

	return strings.TrimSpace(msgType), strings.TrimSpace(msgVal), nil
}

func handleBenchmark(ctx context.Context, conn *websocket.Conn, rootUrl string) error {
	validatedUrl, ok := newFunction(ctx, rootUrl, conn)
	if !ok {
		return errors.New("invalid URL")
	}

	if err := sendStatus(ctx, conn, "crawling"); err != nil {
		return err
	}

	urlsIn := make(chan string)
	benchmarksIn := make(chan Benchmark)

	go benchmarkWebsite(validatedUrl, urlsIn, benchmarksIn)

	// send client crawled urls as we receive them
	for url := range urlsIn {
		write(ctx, conn, "url;"+url)
	}

	if err := sendStatus(ctx, conn, "benchmarking"); err != nil {
		return err
	}

	// then send client benchmarks as we receive them
	for benchmark := range benchmarksIn {
		asJson, err := json.Marshal(benchmark)
		if err != nil {
			log.Printf("Error marshalling benchmark %v\n", err)
			continue
		}
		topipe := "benchmark;" + string(asJson) + "\n"
		write(ctx, conn, topipe)
	}

	return sendStatus(ctx, conn, "complete")
}

func newFunction(ctx context.Context, rootUrl string, conn *websocket.Conn) (validatedUrl string, isValid bool) {
	// validate url, TODO add more here
	if rootUrl == "" {
		return "", false
	}

	// if url does not include protocol, add it
	if !strings.HasPrefix(rootUrl, "http://") || !strings.HasPrefix(rootUrl, "https://") {
		rootUrl = "https://" + rootUrl
	}
	return rootUrl, true
}
