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

func (cr *uploader) Read(p []byte) (n int, err error) {
	if cr.n > 3 {
		n = 0
		err = io.EOF
		return
	}

	err = nil
	n = len(p)
	binary.LittleEndian.PutUint32(p, cr.n+1)
	cr.n += 1

	if n > 4 {
		err = io.EOF
	}
	return
}
func (cr *uploader) Close() error {
	return nil
}

var (
	serverHostname = flag.String("hostname", "localhost", "hostname of the server with which to connect")
	serverPort     = flag.Uint("port", 5002, "port of the server with which to connect")
)

func main() {
	client := http.Client{}
	client.Transport = &http.Transport{WriteBufferSize: 4, TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}

	request, err := client.Post(fmt.Sprintf("https://%s:%d/upload", *serverHostname, *serverPort), "application/octet-stream", &uploader{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "There was an error creating the request: %v (used %s:%d)\n", err, *serverHostname, *serverPort)
		return
	}
	if request.StatusCode == 200 {
		fmt.Printf("Request sent successfully, check the server for confirmation.\n")
	} else {
		fmt.Fprintf(os.Stderr, "There was an error on the server (%s:%d): it returned a status code of %d\n", *serverHostname, *serverPort, request.StatusCode)
	}
}
