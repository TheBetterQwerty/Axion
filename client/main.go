package main

import (
	"fmt"
	"axion/utils"
	"encoding/json"
	"net"
	"os"
	"strings"
)

func HandleUser(sockfd net.Conn, username string, key []byte) {
	buffer := make([]byte, 4096);

	go func() {
		for {
			size, err := sockfd.Read(buffer);
			if err != nil {
				fmt.Printf("\n[!] Error reading from server !\n");
				os.Exit(1);
			}

			var pkt axion.Packet;
			if err := json.Unmarshal(buffer[:size], &pkt); err != nil {
				fmt.Printf("\n[!] Error parsing packet from server!");
				return;
			}

			/* Handle server messages that ain't encrypted */
			if !pkt.Encrypted {
				fmt.Printf("\r[ %s ] %s\n", pkt.Sender, pkt.Data);
				fmt.Printf("[ %s ] ", username);
				continue;
			}

			// decode first
			decode_data, err := pkt.Decrypt_data(key);
			if err != nil {
				fmt.Printf("\n[!] Error: decrypting message %x\n", err);
				return;
			}

			if pkt.Reciever != "SERVER" {
				fmt.Printf("\r[ %s (private) ] %s\n", pkt.Sender, decode_data);
			} else {
				fmt.Printf("\r[ %s ] %s\n", pkt.Sender, decode_data);
			}

			fmt.Printf("[ %s ] ", username);
		}
	}();

	for {
		reciever := "SERVER";

		fmt.Printf("\r[ %s ] ", username); // remove if problems
		input, err := axion.Fgets();
		if err != nil {
			fmt.Printf("\n[!] Error getting input %x\n", err);
			continue;
		}

		{
			/* Parse Input
			 * TODO: Add other commands like /help or /users
			 */

			if data, found := strings.CutPrefix(input, "/msg "); found {
				x := strings.Split(data, " ");
				reciever = x[0];
				input = strings.Join(x[1:], " ");
			}

			if _, found := strings.CutPrefix(input, "/exit"); found {
				fmt.Printf("\r[!] Exitting\n");
				return;
			}
		}

		pkt := axion.New(username, reciever);
		pkt.Set_data(key, input);

		encoded, err := json.Marshal(pkt);
		if err != nil {
			fmt.Printf("\n[!] Error marshalling packey!");
			continue;
		}

		if _, err := sockfd.Write(encoded); err != nil {
			fmt.Printf("\n[!] Error writting to socket %x\n", err);
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
	fmt.Printf("\n");
}
