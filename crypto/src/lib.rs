/* Learn &[u8], array, slice and how it works
 * Learn Deref, pointer in rust
 * */

/* Imports */
use std::io::{ self, Write};
use sha2::{ Sha256, Digest};
use hex::encode;
use aes_gcm::{
    aead::{ Aead, AeadCore, KeyInit, OsRng},
    Aes256Gcm, Key, Nonce
};

/* Packet of data */
pub struct Packet {
    pub encrypted: bool,
    pub sender: String,
    pub reciever: String,
    pub nonce: Vec<u8>,
    pub data: Vec<u8>,
    pub hash: String,
}

impl Packet {
    pub fn new(sender: String, reciever: String) -> Self {
        Packet {
            encrypted: false,
            sender,
            reciever,
            nonce: Vec::new(),
            data: Vec::new(),
            hash: String::new(),
        }
    }
    
    pub fn set_data(&mut self, key: &String, message: String) {
        self.hash = hash(&message);
        let (nonce, encrypted_data) = match encrypt(key, message) {
            Ok((nonce, data)) => (nonce, data),
            Err(_) => {
                println!("[!] Error encrypting data!");
                return;
            },
        };
        self.nonce = nonce;
        self.data = encrypted_data;
        self.encrypted = true;
    }

    pub fn decrypt_packet(&self, key: &String) -> Result<String, String> {
        let decrypted_message = match decrypt(key, &self.data, &self.nonce) {
            Ok(data) => data,
            Err(x) => {
                return Err(x);
            }
        };

        let decrypted_data_hash = hash(&decrypted_message);
        if self.hash == decrypted_data_hash {
            Ok(decrypted_message)
        } else {
            Err("Failed Integrity Check".to_owned())
        }
    }
}   

const KEY_SIZE: usize = 32;

#[allow(dead_code)]
pub fn read_key() -> String {
    let mut input = String::new();

    print!("[+] Enter your passphrase: ");
    let _ = io::stdout().flush();
    io::stdin().read_line(&mut input).expect("[!] Error reading line!");
    
    input = input.trim().to_owned();

    let len = input.len();
    if len < KEY_SIZE {
        let padding = "*".repeat(KEY_SIZE - len);
        return input + &padding;
    } else if len > KEY_SIZE {
        input.truncate(KEY_SIZE);
        return input;
    } else {
        return input;
    }
}

fn hash(text: &String) -> String {
    let result = Sha256::digest(text.as_bytes());
    return encode(result);
}

fn encrypt(key: &String, plaintext: String) -> Result<(Vec<u8>, Vec<u8>), String> {
    let key = Key::<Aes256Gcm>::from_slice(key.as_bytes());
    let nonce = Aes256Gcm::generate_nonce(&mut OsRng);
    let cipher = Aes256Gcm::new(key);

    let cipher_data = match cipher.encrypt(&nonce, plaintext.as_bytes()) {
        Ok(data) => data,
        Err(_) => {
            return Err("Error encrypting data".to_owned());
        },
    };

    return Ok((nonce.to_vec(), cipher_data));
}

fn decrypt(key: &String, encrypted_data: &[u8], nonce: &[u8]) -> Result<String, String> {
    let key = Key::<Aes256Gcm>::from_slice(key.as_bytes());
    let nonce = Nonce::from_slice(nonce);
    let cipher = Aes256Gcm::new(key);

    let plaintext = match cipher.decrypt(nonce, encrypted_data.as_ref()) {
        Ok(x) => x,
        Err(_) => {
            return Err("[!] Error decrypting message!".to_owned());
        }
    }; 

    match String::from_utf8(plaintext) {
        Ok(x) => Ok(x),
        Err(_) => Err("[!} Error converting message to string!".to_owned()),
    }
}
