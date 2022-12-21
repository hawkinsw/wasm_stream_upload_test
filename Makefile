all: web_client/client.wasm server.bin
web_client/client.wasm: client/client.go
	GOOS=js GOARCH=wasm go build -o web_client/client.wasm client/client.go
server.bin: server/server.go
	go build -o server.bin server/server.go
keys:
	openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -sha256 -days 365 -nodes
clean:
	rm -f server.bin web_client/client.wasm
