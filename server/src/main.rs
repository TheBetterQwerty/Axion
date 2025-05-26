use crypto::{ read_key, Packet };
use std::io::{self, Write};

fn fgets(txt: &str) -> String {
    let mut input = String::new();
    print!("{}", txt);
    let _ = io::stdout().flush();
    io::stdin().read_line(&mut input).expect("[!] Error reading input");

    return input.trim().to_owned();
}

fn main() {}
