cc = go

server = server
server_bin = Server

client = client
client_bin = Client

tui-dep1 = github.com/rivo/tview
tui-dep2 = github.com/gdamore/tcell/v2

.PHONY: server client

server:
	$(cc) build -o $(server_bin) $(server)/main.go

client: $(dependencies)
	$(cc) build -o $(client_bin) $(client)/main.go

dependencies:
	$(cc) get $(tui-dep1)
	$(cc) get $(tui-dep2)

clean:
	@echo "Deleting server binary"
	@echo "Deleting client binary"
	rm $(client_bin) $(server_bin)
