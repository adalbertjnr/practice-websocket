SERVER_BINARY := server
CLIENT_BINARY := client

server-build:
	@go build -o bin/$(SERVER_BINARY) 

client-build:
	@go build -o bin/$(CLIENT_BINARY) ./client

server: server-build
	@./bin/$(SERVER_BINARY)

client: client-build
	@./bin/$(CLIENT_BINARY)