package main

import (
	"fmt"
	"axion/utils"
)

func print_pkt(pkt *axion.Packet) {
	fmt.Printf("[ PACKET ]\n");
	fmt.Printf("[ENCRYPTED] %d\n", (*pkt).Encrypted);
	fmt.Printf("[SENDER] -> %s\n", (*pkt).Sender);
	fmt.Printf("[RECIEVER] -> %s\n", (*pkt).Reciever);
	fmt.Printf("[NONCE] -> %x\n", (*pkt).Nonce);
	fmt.Printf("[DATA] -> %x\n", (*pkt).Data);
	fmt.Printf("[HASH] -> %s\n", (*pkt).Hash);
}

func main() {
	pkt := axion.New("nigger", "SERVER");
	key := axion.GetKey();

	pkt.Set_data(key, "hello niggers how are you");
	print_pkt(&pkt);

	x, _ := pkt.Decrypt_data(key);
	fmt.Printf("Data is %s\n", x);
}
