package axion

import (
	"crypto/aes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/hmac"
	"crypto/cipher"
	"encoding/hex"
	"strings"
	"io"
	"fmt"
	"bufio"
	"os"
)

type Packet struct {
	Encrypted	bool   // 1 byte only
	Sender 		string
	Reciever 	string
	Nonce		[]byte // 16 bytes only
	Data		string
	Hash 		string // 32 bytes only
}

/* Functions to be used with Packet */
func New(sender string, reciever string) Packet {
	return Packet {
		false,
		sender,
		reciever,
		[]byte{},
		"",
		"",
	};
}

func (pkt *Packet) Set_data(key []byte, data string) error {
	ciphertext, iv, err := encrypt_data(key, data);
	if err != nil {
		fmt.Printf("[!] Error %s\n", err);
		return err;
	}

	mac := hmac.New(sha256.New, key);
	mac.Write([]byte(pkt.Sender));
	mac.Write([]byte(pkt.Reciever));
	mac.Write(iv);
	mac.Write([]byte(data));

	pkt.Encrypted = true;
	pkt.Data = ciphertext;
	pkt.Nonce = iv;
	pkt.Hash = hex.EncodeToString(mac.Sum(nil));

	return nil;
}

func (pkt Packet) Decrypt_data(key []byte) (string, error) {
	if plaintext, err := decrypt_AES(key, pkt.Data, pkt.Nonce); err != nil {
		return "", err;
	} else {
		msg_mac, err := hex.DecodeString(pkt.Hash);
		if err != nil {
			return "", nil;
		}

		mac := hmac.New(sha256.New, key);
		mac.Write([]byte(pkt.Sender));
		mac.Write([]byte(pkt.Reciever));
		mac.Write(pkt.Nonce);
		mac.Write([]byte(plaintext));

		req_mac := mac.Sum(nil);

		if hmac.Equal(msg_mac, req_mac) {
			return plaintext, nil;
		}

		return "Hmac Doesnt match!", nil
	}
}
/* Packet functions ends */

func Fgets() (string, error) {
	reader := bufio.NewReader(os.Stdin);
	text, err := reader.ReadString('\n');
	if err != nil {
		return "", err;
	}

	return strings.TrimSuffix(text, "\n"), nil;
}

func GetKey() []byte {
	fmt.Printf("[+] Enter password : ");
	passwd, err := Fgets();
	if err != nil {
		panic(err);
	}

	return hash(passwd);
}

func hash(txt string) []byte {
	hash := sha256.Sum256([]byte(txt));
	return hash[:];
}

func encrypt_data(key []byte, data string) (string, []byte, error) {
	text := []byte(data);
	block, err := aes.NewCipher(key);
	if err != nil {
		return "", []byte{}, err;
	}

	iv := make([]byte, aes.BlockSize);
	ciphertext := make([]byte, len(text));

	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", []byte{}, err;
	}

	stream := cipher.NewCTR(block, iv);
	stream.XORKeyStream(ciphertext, text);

	return hex.EncodeToString(ciphertext), iv, nil;
}

func decrypt_AES(key []byte, data string, iv []byte) (string, error) {
	ciphertext, err := hex.DecodeString(data);
	if err != nil {
		return "", err;
	}

	block, err := aes.NewCipher(key);
	if err != nil {
		return "", err;
	}

	stream := cipher.NewCTR(block, iv);
	stream.XORKeyStream(ciphertext, ciphertext);

	return string(ciphertext), nil;
}
