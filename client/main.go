package main

import (
	"fmt"
	"axion/utils"
	"encoding/json"
	"net"
)

func HandleUser(sockfd net.Conn, username string, key []byte) {
	buffer := make([]byte, 4096);

	go func() {
		for {
			size, err := sockfd.Read(buffer);
			if err != nil {
				fmt.Printf("[!] Error reading from server !");
				return;
			}

			var pkt axion.Packet;
			if err := json.Unmarshal(buffer[:size], &pkt); err != nil {
				fmt.Printf("[!] Error parsing packet from server!");
				return;
			}

			// decode first
			decode_data, err := pkt.Decrypt_data(key);
			if err != nil {
				fmt.Printf("[!] Error: %x\n", err);
				return;
			}

			fmt.Printf("\r[ %s ] %s\n", pkt.Sender, decode_data);
		}
	}();

	for {
		fmt.Printf("[ %s ] ", username);
		input, err := axion.Fgets();
		if err != nil {
			fmt.Printf("[!] Error %x\n", err);
			continue;
		}

		pkt := axion.New(username, "SERVER"); // handle private msg later
		pkt.Set_data(key, input);

		encoded, err := json.Marshal(pkt);
		if err != nil {
			fmt.Printf("[!] Error marshalling packey!");
			continue;
		}

		if _, err := sockfd.Write(encoded); err != nil {
			fmt.Printf("[!] Error %x\n", err);
			return;
		}
	}
}

func main() {
	fmt.Printf("[+] Enter your username: ");
	username, err := axion.Fgets();
	if err != nil {
		fmt.Printf("[!] Error %x\n", err);
		return;
	}

	passwd := axion.GetKey();

	sockfd, err := net.Dial("tcp", "127.0.0.1:8080");
	if err != nil {
		fmt.Printf("[!] Error : %x\n", err);
		return;
	}
	defer sockfd.Close();

	{
		/* send server login packet! */
		pkt := axion.New(username, "SERVER");
		encoded, _ := json.Marshal(pkt);
		if _, err := sockfd.Write(encoded); err != nil {
			fmt.Printf("[!] Error sending data to server!");
			return;
		}
	}

	HandleUser(sockfd, username, passwd);
}
