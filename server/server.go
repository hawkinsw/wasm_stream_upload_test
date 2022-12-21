package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"

	// Do *not* remove this import. Per https://pkg.go.dev/net/http/pprof:
	// The package is typically only imported for the side effect of registering
	// its HTTP handlers. The handled paths all begin with /debug/pprof/.
	_ "net/http/pprof"
	// See -debug for how we use it.
)

var (
	listenPort   = flag.Int("port", 5002, "The port to listen on for measurement accesses")
	listenAddr   = flag.String("addr", "localhost", "address to bind to")
	certFilename = flag.String("cert", "cert.pem", "address to bind to")
	keyFilename  = flag.String("key", "key.pem", "address to bind to")
)

func main() {
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/upload", uploadHandler)
	mux.Handle("/test/", http.StripPrefix("/test/", http.FileServer(http.Dir("web_client/"))))

	var wg sync.WaitGroup
	wg.Add(1)

	server := &http.Server{}
	server.Addr = fmt.Sprintf("%s:%d", *listenAddr, *listenPort)
	server.Handler = mux

	go func() {
		if err := server.ListenAndServeTLS(*certFilename, *keyFilename); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
		wg.Done()
	}()

	// The user can stop the server with SIGINT
	signalChannel := make(chan os.Signal, 1)   // make the channel buffered, per documentation.
	signal.Notify(signalChannel, os.Interrupt) // only Interrupt is guaranteed to exist on all platforms.

SignalLoop:
	for {
		select {
		case <-signalChannel:
			log.Printf("Shutting down the server ...\n")
			server.Shutdown(context.Background())
			break SignalLoop
		}
	}

	wg.Wait()
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Proxy-Cache-Control", "max-age=604800, public")
	w.Header().Set("Cache-Control", "no-store, must-revalidate, private, max-age=0")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// The client is going to send 4 bytes every time that read is called. And it is going to put
	// the number of that call in those bytes.
	streamingGroundTruth := "\x01\x00\x00\x00\x02\x00\x00\x00\x03\x00\x00\x00\x04\x00\x00\x00"
	bulkGroundTruth := "\x01\x00\x00\x00"

	fmt.Printf("body I read: %v\n", strings.ToValidUTF8(string(body), ""))
	if bytes.Compare([]byte(streamingGroundTruth), body) == 0 {
		fmt.Printf("The client is streaming their uploads.\n")
	} else if bytes.Compare([]byte(bulkGroundTruth), body[:4]) == 0 {
		fmt.Printf("The client is *not* streaming their uploads.\n")
	} else {
		fmt.Printf("There was a general problem with the client's uploads.\n")
	}
}

func testHandler(w http.ResponseWriter, r *http.Request) {

}
