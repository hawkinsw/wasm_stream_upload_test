## Test CL 458395

This code will test whether a web browser is streaming the body of a POST request or uploading it in bulk.

## Operation

### Prequisites
1. Go
2. Chrome/Firefox/Opera
3. `openssl` on your path (for generating self-signed certificates for HTTPS)
4. `make` on your path (for automated building)

### Configuration/Installation

```
$ make
```

will generate 

1. `server.bin`: A binary that runs a webserver that hosts the test (from [`server/server.go`](server/server.go))
2. `web_client/main.wasm`: The wasm code for the client (from [`client/client.go`](client/client.go))
3. `key.pem`, `cert.pem`: Self-signed key and certificate needed so that the testing server can do HTTPS

### Running

To start the server, execute
```
$ ./server.bin
```
When you are finished, send `SIGINT` to the server process (Ctrl-C on Linux/macOS).

#### Testing in a browser

To test support in a web browser for streaming the body of POST requests, click on the link displayed by the server when it starts. 

#### Testing in Deno

To test Deno support for streaming the body of POST requests, execute

```console
$ deno run --unsafely-ignore-certificate-errors web_client/deno.js
```

from the root directory of this source code.


### Outcomes

If the browser you are using to access the testing server supports streaming the body of POST requests,
the server will output

```
The client is streaming their uploads.
```

On the other hand, if the browser you use to access the test server does *not* support streaming the body
of POST requests, the server will output

```
The client is *not* streaming their uploads.
```

## How

See the comments through [`client/client.go`](client/client.go) and [`server/server.go`](server/server.go) for information on *how* the test does its work.
