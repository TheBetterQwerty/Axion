package axion

import (
	"crypto/aes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/cipher"
	"encoding/hex"
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
	_hash := hash(data);
	ciphertext, iv, err := encrypt_data(key, data);
	if err != nil {
		fmt.Printf("[!] Error %s\n", err);
		return err;
	}
	pkt.Encrypted = true;
	pkt.Data = ciphertext;
	pkt.Nonce = iv;
	pkt.Hash = hex.EncodeToString(_hash);
	return nil;
}

func (pkt Packet) Decrypt_data(key []byte) (string, error) {
	if plaintext, err := decrypt_AES(key, pkt.Data, pkt.Nonce); err != nil {
		return "", err;
	} else {
		_hash := hex.EncodeToString(hash(plaintext));
		if _hash == pkt.Hash {
			return plaintext, nil;
		} else {
			return "Hashes dont match", nil;
		}
	}
}

/* Packet functions ends */

func Fgets() (string, error) {
	reader := bufio.NewReader(os.Stdin);
	text, err := reader.ReadString('\n');
	if err != nil {
		return "", err;
	}

	return text, nil;
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
