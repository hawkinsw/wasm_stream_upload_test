package main

import (
	"crypto/tls"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

type uploader struct {
	n uint32
}

// In the situation where the wasm runtime detects that streaming uploads are possible,
// it will ask for the body of the POST `WriteBufferSize` bytes at a time. When we configure
// the client that will do the POST (see below), we configure that value to be 4 bytes.
//
// In the situation where the wasm runtime detects that streaming uploads are *not* possible,
// it will ask for "all" of the body and the runtime will immediately ask for a swath of 512
// bytes (see io/io.go:695 in the source).
//
// We will rely on this difference in behavior to post data to the server that will help it
// determine whether this client is delivering its body to the runtime (and subsequently the web
// server) all at once or whether it is being sent it pieces.
//
// Ultimately, the server will get
// \x01\x00\x00\x00\x02\x00\x00\x00\x03\x00\x00\x00\x04\x00\x00\x00
// as the body of the request from a client that is streaming and
// \x01\x00\x00\x00
// as the body of the request from a client that is not streaming.
func (cr *uploader) Read(p []byte) (n int, err error) {

	// Once we have sent 4 chunks, we will stop! (streaming case)
	if cr.n > 3 {
		n = 0
		err = io.EOF
		return
	}

	// In all cases, write the value of cr.n into the response (in little-endian format).
	err = nil
	n = len(p)
	binary.LittleEndian.PutUint32(p, cr.n+1)
	cr.n += 1

	// If the length of the buffer is greater than 4, we know that we are bulk sending
	// and we bail out immediately (see above).
	if n > 4 {
		err = io.EOF
	}
	return
}

func (cr *uploader) Close() error {
	return nil
}

// Note: When we build this with the wasm target, the knobs to set these values are not exposed to the end
// user -- be careful.
var (
	serverHostname = flag.String("hostname", "localhost", "hostname of the server with which to connect")
	serverPort     = flag.Uint("port", 5002, "port of the server with which to connect")
)

func main() {
	client := http.Client{}
	// Set the WriteBufferSize (see above).
	client.Transport = &http.Transport{WriteBufferSize: 4, TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}

	request, err := client.Post(fmt.Sprintf("https://%s:%d/upload", *serverHostname, *serverPort), "application/octet-stream", &uploader{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "There was an error creating the request: %v (used %s:%d)\n", err, *serverHostname, *serverPort)
		return
	}
	if request.StatusCode == 200 {
		fmt.Printf("Request sent successfully, check the server for diagnostic results.\n")
	} else {
		fmt.Fprintf(os.Stderr, "There was an error on the server (%s:%d): it returned a status code of %d\n", *serverHostname, *serverPort, request.StatusCode)
	}
}
