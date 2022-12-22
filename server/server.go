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
	"sync"
)

var (
	// Giving valeus to  the port and addr command-line options will require you to
	// rebuild the client with matching values (even though the client allows the user
	// to set those values on the command line as well).
	listenPort = flag.Int("port", 5002, "The port to listen on for measurement accesses")
	listenAddr = flag.String("addr", "localhost", "address to bind to")
	// `make` will generate appropriately named cert and key files that match these defaults.
	certFilename = flag.String("cert", "cert.pem", "address to bind to")
	keyFilename  = flag.String("key", "key.pem", "address to bind to")
)

// See client.go for documentation on why we use these particular values.
const streamingGroundTruth = "\x01\x00\x00\x00\x02\x00\x00\x00\x03\x00\x00\x00\x04\x00\x00\x00"
const bulkGroundTruth = "\x01\x00\x00\x00"

func main() {

	// See what configuration options the user gave.
	flag.Parse()

	mux := http.NewServeMux()
	// /upload is the endpoint to which the wasm client will POST.
	mux.HandleFunc("/upload", uploadHandler)
	// We use a http.FileServer to serve the wasm and HTML (statically) from the web_client
	// directory (on disk) to the /test/ directory (on the web).
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

	fmt.Printf("Server is now ready to accept testing connections (https://%s:%d/test/).\n", *listenAddr, *listenPort)

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
	// Make sure that we set up the headers ...
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Proxy-Cache-Control", "max-age=604800, public")
	w.Header().Set("Cache-Control", "no-store, must-revalidate, private, max-age=0")

	// Take in all the data from the client.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if bytes.Compare([]byte(streamingGroundTruth), body) == 0 {
		fmt.Printf("The client is streaming their uploads.\n")
	} else if bytes.Compare([]byte(bulkGroundTruth), body[:4]) == 0 {
		fmt.Printf("The client is *not* streaming their uploads.\n")
	} else {
		fmt.Printf("There was a general problem with the client's uploads.\n")
	}
}
