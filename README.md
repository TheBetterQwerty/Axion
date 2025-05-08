# Axion - Secure Chat Client (C)

Axion is a secure AES chat client-server communication system built in C, using encryption techniques to ensure confidentiality. The client communicates with the server over a socket connection, supporting both encrypted and plaintext messages. It features AES encryption for securing messages, with support for public and private messaging.

## Table of Contents

* [Features](#features)
* [Installation](#installation)
* [Usage](#usage)
* [File Structure](#file-structure)
* [Dependencies](#dependencies)
* [Contributing](#contributing)
* [License](#license)

## Features

* **AES Encryption:** Encrypt messages using AES encryption to ensure security.
* **Private Messaging:** Send private messages between users.
* **Server Communication:** Communicate with the server or send direct messages to users.
* **Graceful Shutdown:** Handles interruptions (SIGINT) properly, ensuring the client shuts down gracefully.
* **Simple Interface:** Easy-to-use text-based prompt for sending messages.

## Installation

### Prerequisites

Make sure you have the following installed:

* A C compiler (e.g., `gcc`)
* Make tool (optional, for building the project)
* OpenSSL library for AES encryption (if you're linking external libraries)
* pthread library for multi-threading support

### Steps

1. Clone the repository to your local machine:

   ```bash
   git clone https://github.com/your-repo/axion.git
   cd axion
   ```

2. Ensure that the AES encryption files (like `aes.h` and other related code) are available in the `../AES/` directory, or adjust the include path accordingly.

3. Compile the program:

   ```bash
   make
   ```

4. Run the server (make sure you have a server running on `127.0.0.1:6000`):

   ```bash
   ./server/main # Assuming you have the server code
   ```

5. Start the client:

   ```bash
   ./client/main
   ```

## Usage

### Running the Client

1. When you start the client, it will prompt you to enter a username.

2. Once connected to the server, you can send messages to either the server or a specific user:

   * **Public messages**: Just type your message and hit enter.
   * **Private messages**: Use the command:

     ```
     /msg <username> <message>
     ```

3. The client can encrypt messages by default, or you can modify the code to disable encryption.

4. To exit the client, type:

   ```
   /exit
   ```

### Example Interaction

```bash
[>] Enter your username: Alice
[ Alice ] Hello, world!
[ SERVER ] Server message to Alice
[ Alice (priv) ] Private message to Bob
```

## File Structure

* `client.c`: Main client code for interacting with the server and sending messages.
* `../AES/aes.h`: Header file for AES encryption (ensure it's available or modify the path).
* `server.c`: (Not included here, but assumed to be part of the project) The server that listens to and handles client connections.
* `Makefile` (optional): Automates the build process (if used).

## Dependencies

* **pthread**: Required for multi-threading in the client.
* **OpenSSL**: Required for AES encryption and decryption.
* **Standard C Libraries**: Uses `stdio.h`, `string.h`, `stdlib.h`, `unistd.h`, `signal.h`, and others.

To install the necessary libraries:

* On Ubuntu/Debian:

  ```bash
  sudo apt-get install libssl-dev 
  ```

## Contributing

We welcome contributions to **Axion**! If you find a bug or have a feature idea, feel free to fork the repository and submit a pull request.

### Steps to contribute:

1. Fork the repository.
2. Create a new branch for your changes (`git checkout -b feature-name`).
3. Commit your changes (`git commit -am 'Add new feature'`).
4. Push to the branch (`git push origin feature-name`).
5. Open a pull request with a description of your changes.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---
