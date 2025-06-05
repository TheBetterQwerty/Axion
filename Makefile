cc = go

server=server
client=client

server:
	$(cc) build $(server)/main.go -o $(server)/main

client:
	$(cc) build $(client)/main.go -o $(client)/main

clean:
	echo "Deleting server binary"
	rm $(server)/main
	echo "Deleting client binary"
	rm $(client)/main
