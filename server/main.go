package main

import (
	"fmt"
	"net"
	"encoding/json"
	"axion/utils"
)

/* Stores the Active Users */
var Users = make(map[string]net.Conn);

func broadcast(pkt axion.Packet) {
	/* Send Everyone except sender */
	encoded, err := json.Marshal(pkt);
	if err != nil {
		fmt.Printf("[!] Error %x\n", err);
		return;
	}

	for username, fd := range Users {
		if username == pkt.Sender {
			continue;
		}

		_, err := fd.Write(encoded);
		if err != nil {
			fmt.Printf("[!] Error sending data to %s\n", username);
			continue;
		}
	}
}

func unicast(pkt axion.Packet) {
	/* Send only reciever */
}

func handler(fd net.Conn) {
	defer fd.Close();
	buffer := make([]byte, 4096);

	{
		/* Reads packet and logs user into the server */
		size, err := fd.Read(buffer);
		if err != nil {
			fmt.Printf("[!] Error reading from client!");
			return;
		}

		var login axion.Packet
		if err := json.Unmarshal(buffer[:size], &login); err != nil {
			fmt.Printf("[!] Error Unmarshal the packet!");
			return;
		}

		Users[login.Sender] = fd; /* Handle same username case and name cannot be SERVER */


		pkt := axion.New("SERVER", "SERVER");
		pkt.Data = fmt.Sprintf("%s joined the chat!", login.Sender);
		broadcast(pkt);
	}

	for {
		size, err := fd.Read(buffer);
		if err != nil {
			fmt.Printf("[!] Error reading from client!");
			return;
		}

		var packet axion.Packet;
		if err := json.Unmarshal(buffer[:size], &packet); err != nil {
			fmt.Printf("[!] Error Unmarshal the packet!");
			return;
		}

		if packet.Reciever == "SERVER" {
			broadcast(packet);
		} else {
			unicast(packet);
		}
	}

	// delete(Users,  # remove the user after the loop ends
}

func main() {
	listener, err := net.Listen("tcp", ":8080");
	if err != nil {
		fmt.Printf("[!] Error creating socket!");
		return;
	}
	defer listener.Close();

	for {
		socket, err := listener.Accept();
		if err != nil {
			fmt.Printf("[!] Error accepting client!");
			continue;
		}

		go handler(socket);
	}
}
