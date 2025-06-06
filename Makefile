cc = go

server=server
client=client

server:
	$(cc) build -o $(server)/main $(server)/main.go

client:
	$(cc) build -o $(client)/main $(client)/main.go

clean:
	echo "Deleting server binary"
	rm $(server)/main
	echo "Deleting client binary"
	rm $(client)/main
