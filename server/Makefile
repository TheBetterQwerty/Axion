compiler = gcc

src = main.c 
output = main 
aes = ../AES/aes.c 
cflags = -Wextra -Wall -O3 
flags = -lssl -lcrypto 

all: 
        $(compiler) -o $(output) $(src) $(aes) $(flags) $(cflags) 

clean:
        rm $(output)
