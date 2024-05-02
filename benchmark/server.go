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
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Printf("failed to accept websocket: %v\n", err)
		return
	}

	defer conn.CloseNow()

	for {
		err = handleConnection(r.Context(), conn)
		switch {
		case err == io.EOF:
			fmt.Println("Reached EOF")
			continue
		case websocket.CloseStatus(err) == websocket.StatusNormalClosure:
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

func handleBenchmark(ctx context.Context, conn *websocket.Conn, url string) error {
	var err error

	startmsg := "message;Benchmarking " + url + "\n"
	log.Print(startmsg)
	err = write(ctx, conn, startmsg)
	if err != nil {
		return err
	}

	benchmarkWebsite(url, func(p Performance) {
		asJson, _ := json.Marshal(p)
		topipe := "url_performance;" + string(asJson) + "\n"
		write(ctx, conn, topipe)
	})

	endmsg := "Benchmarking complete\n"
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
