# Axion

Axion is a lightweight and secure TCP-based chat system written in Go. It provides encrypted communication using AES without requiring key exchange, focusing on simplicity and security.

## Features

### Server
- Listens for incoming TCP connections.
- Supports broadcasting or unicasting of encrypted packets to connected clients.
- Shows Debug info 

### Client
- Sends encrypted packets to the server using AES.
- Decrypts received messages using a pre-shared password.
- If an incorrect password is used (e.g., by an intruder), messages cannot be decrypted and integrity checks will fail.
- After decryption, each message is hashed and compared to the hash sent by the server to verify authenticity and integrity.

## Planned Features

The following features are planned for future releases:

- **Colored Output:** Improve readability in the terminal by adding color-coded messages (e.g., by user, system messages, etc.).
- **Chat Rooms:** Add support for multiple chat rooms so users can join specific groups.
- **User Roles:** Implement role-based access (e.g., admin, guest, moderator) for better control and moderation.
- **HTAC Integrity Check:** Integrate Hash Tree-based Authentication Code (HTAC) to further ensure message integrity and tamper detection across the communication chain.

## Build Instructions

Make sure Go is installed (`go version` should return a valid version).

### To build and run the server:
```bash
make server
./server/main 
```

### To build and run the client:

```bash
make client
./client/main 
```

## License
```
This project is licensed under the terms provided in the [MIT License](LICENSE) file.
```


