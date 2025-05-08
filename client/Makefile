compiler = gcc 

src = main.c 
final = main 
crypto = ../AES/aes.c
cflags = -Wextra -Wall -O3
flags = -lssl -lcrypto 

all: 
        $(compiler) -g -o $(final) $(src) $(crypto) $(cflags) $(flags)

clean: 
        rm $(final) 
