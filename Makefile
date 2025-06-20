cc = go

server=server
server_bin = Server
client=client
client_bin = Client

.PHONY: server client

server:
	$(cc) build -o /$(server_bin) $(server)/main.go

client:
	$(cc) build -o $(client_bin) $(client)/main.go

clean:
	@echo "Deleting server binary"
	@echo "Deleting client binary"
	rm $(client_bin) $(server_bin)
