package main

import (
	"fmt"
	"net"
	"encoding/json"
	"axion/utils"
)

/* Stores the Active Users */
var Users = make(map[string]net.Conn);

func broadcast(pkt *axion.Packet) {
	/* Send Everyone except sender */
	encoded, err := json.Marshal(*pkt);
	if err != nil {
		fmt.Printf("[!] Error %v\n", err);
		return;
	}

	for username, fd := range Users {
		if username == (*pkt).Sender {
			continue;
		}

		if _, err := fd.Write(encoded); err != nil {
			fmt.Printf("[!] Error sending data to %s\n", username);
		} else {
			fmt.Printf("[+] Sent packet to %s\n", username); // DBG
		}
	}
}

func unicast(pkt *axion.Packet) {
	/* Send only reciever */
	encoded, err := json.Marshal(*pkt);
	if err != nil {
		fmt.Printf("[!] Error: %v\n", err);
		return;
	}

	username := (*pkt).Reciever;
	fd, found := Users[username];
	if !found {
		// send info that the reciever doesnt exist
		new_pkt := axion.New("SERVER", (*pkt).Sender);
		new_pkt.Data = fmt.Sprintf("\"%s\" user doesn't exists", username);

		if _fd, exists := Users[pkt.Sender]; exists {
			pkt_encoded, err := json.Marshal(new_pkt);
			if err != nil {
				fmt.Printf("[!] Error marshaling text!\n");
				return;
			}

			if _, err := _fd.Write(pkt_encoded); err != nil {
				fmt.Printf("[!] Error writting to client!\n");
				return;
			}
		}

		fmt.Printf("%s\n", new_pkt.Data); // DBG
		return;
	}

	if _, err := fd.Write(encoded); err != nil {
		fmt.Printf("[!] Error sending data to %s\n", username);
	} else {
		fmt.Printf("[+] Sent packet to %s\n", username); // DBG
	}
}

func handler(fd net.Conn) {
	defer fd.Close();

	var username string;
	buffer := make([]byte, 4096);

	{
		/* Reads packet and logs user into the server */
		size, err := fd.Read(buffer);
		if err != nil {
			fmt.Printf("[!] Error reading from client!\n");
			return;
		}

		var login axion.Packet
		if err := json.Unmarshal(buffer[:size], &login); err != nil {
			fmt.Printf("[!] Error Unmarshal the packet!\n");
			return;
		}
		username = login.Sender;

		/* If username same as SERVER */
		if username == "SERVER" {
			pkt := axion.New("SERVER", username);
			pkt.Data = "\"SERVER\" is reserved for system use";
			encoded, _ := json.Marshal(pkt);
			if _, err := fd.Write(encoded); err != nil {
				fmt.Printf("[!] Error sending data to %s\n", username);
			} else {
				fmt.Printf("[+] Sent packet to %s\n", username); // DBG
				fmt.Printf("[$] client tried to use reserved name\n");
			}
			return;
		}

		/* If username exists */
		if _, exists := Users[username]; exists {
			pkt := axion.New("SERVER", username);
			pkt.Data = "Username already taken";
			encoded, _ := json.Marshal(pkt);
			if _, err := fd.Write(encoded); err != nil {
				fmt.Printf("[!] Error sending data to %s\n", username);
			} else {
				fmt.Printf("[+] Sent packet to %s\n", username); // DBG
				fmt.Printf("[$] client tried to use existing name\n");

			}
			return;
		}

		Users[username] = fd;

		pkt := axion.New("SERVER", "SERVER");
		pkt.Data = fmt.Sprintf("%s joined the chat!", username);
		go broadcast(&pkt);
	}

	for {
		size, err := fd.Read(buffer);
		if err != nil {
			fmt.Printf("[!] Error reading from client!\n");
			break;
		}

		var packet axion.Packet;
		if err := json.Unmarshal(buffer[:size], &packet); err != nil {
			fmt.Printf("[!] Error Unmarshal the packet!\n");
			break;
		}

		if packet.Reciever == "SERVER" {
			go broadcast(&packet);
		} else {
			go unicast(&packet);
		}
	}

	delete(Users, username);

	{
		/* Broadcast the message */
		pkt := axion.New("SERVER", "SERVER");
		pkt.Data = fmt.Sprintf("%s left the chat!", username);
		go broadcast(&pkt);
	}
}

func main() {
	listener, err := net.Listen("tcp", ":8080");
	if err != nil {
		fmt.Printf("[!] Error creating socket\n!");
		return;
	} else {
		fmt.Printf("[#] Listening on 127.0.0.1:8080\n");
	}
	defer listener.Close();

	for {
		socket, err := listener.Accept();
		if err != nil {
			fmt.Printf("[!] Error accepting client!\n");
			continue;
		}

		go handler(socket);
	}
}
