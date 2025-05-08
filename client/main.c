#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <unistd.h>
#include <signal.h>
#include <pthread.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>

#include "../AES/aes.h"

#define PORT 6000
#define IP "127.0.0.1"
#define MSG_LEN 32
#define BUFFER_SIZE 1024

typedef struct {
    int encrypted;
    char sender[MSG_LEN];
    char reciever[MSG_LEN];
    unsigned char iv[MSG_LEN / 2];
    int len;
    unsigned char msg[BUFFER_SIZE + MSG_LEN];
    char hash[MSG_LEN * 2];
} Packet;

int fd = -1;
int server_state = 1;
pthread_mutex_t mutex = PTHREAD_MUTEX_INITIALIZER;
pthread_t server_listener, user_listener;
char user[MSG_LEN];
unsigned char key[MSG_LEN];

void kill_client(int signum __attribute__((unused))) {
    signal(SIGINT, SIG_IGN);

    pthread_mutex_lock(&mutex);
    server_state = 0;
    if (fd > 0) close(fd);
    pthread_mutex_unlock(&mutex);

    exit(EXIT_SUCCESS);
}

void copy(char* dst, const char* src, size_t dst_len) {
    if (!dst || !src || dst_len <= 0) return;
    size_t src_len = strlen(src);
    size_t len = (src_len < dst_len - 1) ? src_len : dst_len - 1;
    memcpy(dst, src, len);
    dst[len] = '\0';
}

void print_prompt() {
    char prompt[MSG_LEN + 8];
    snprintf(prompt, sizeof(prompt), "[ %s ] ", user);
    write(STDOUT_FILENO, prompt, strlen(prompt));
    fflush(stdout);
}

void decrypt_packet(Packet* pkt) {
    if (pkt->encrypted == 0) {
        printf("\33[2K\r[ %s ] %s\n", pkt->sender, pkt->msg);
        print_prompt();
        return;
    }
    
    char decrypted_msg[BUFFER_SIZE + MSG_LEN];
    if (decrypt(pkt->msg, pkt->len, key, pkt->iv, (unsigned char*) decrypted_msg) <= 0) {
        printf("\33[2K\r[ %s ] %s\n", "CLIENT", "ERROR decrypting msg");
        print_prompt();
        return;
    }
    
    /*
    char hash[MSG_LEN * 2];
    Hash((const unsigned char*) decrypted_msg, hash);
    if (memcmp(pkt->hash, hash, strlen(hash)) == 0) {
        printf("\33[2K\r[ %s ] %s hash -> %s\n", "CLIENT", "ERROR hashes doesnt match", hash);
        print_prompt();
        return;
    } */

    if (memcmp(pkt->reciever, "SERVER", 6) == 0) {
        printf("\33[2K\r[ %s ] %s %ld\n", pkt->sender, decrypted_msg, strlen(decrypted_msg));
    } else {
        printf("\33[2K\r[ %s (priv) ] %s\n", pkt->sender, decrypted_msg);
    }
    
    memset(decrypted_msg, 0, sizeof(decrypted_msg));
    print_prompt();
}

void* server_message(void* args __attribute__((unused))) {
    pthread_setcancelstate(PTHREAD_CANCEL_ENABLE, NULL);
    pthread_setcanceltype(PTHREAD_CANCEL_DEFERRED, NULL);

    Packet* pkt = malloc(sizeof(Packet));
    if (!pkt) {
        printf("[!] Err%d: error allocating memory!\n", __LINE__);
        return NULL;
    }

    while ((read(fd, pkt, sizeof(Packet)) > 0) && (server_state)) {
        // print_packet(pkt);
        decrypt_packet(pkt);
        memset(pkt, 0, sizeof(Packet));
    }

    free(pkt);
    kill_client(__LINE__);
    pthread_detach(pthread_self());
    return NULL;
}

void create_packet(Packet* pkt, char* msg, int encrypted) {
    pkt->encrypted = encrypted;
    copy(pkt->sender, user, MSG_LEN);

    int arr[2], idx = 0;
    for (int i = 0; i < (int) strlen(msg) && idx < 2; i++) {
        if (msg[i] != ' ') continue;
        arr[idx++] = i;
    }
    
    char hash[MSG_LEN * 2];

    if (memcmp(msg, "/msg", 4) != 0) {
        copy(pkt->reciever, "SERVER", MSG_LEN * 2);
        
        if (pkt->encrypted) {
            generate_random_iv(pkt->iv);
            pkt->len = encrypt((unsigned char*) msg, strlen(msg), 
                                key, pkt->iv, pkt->msg);
            if (pkt->len <= 0) {
                printf("\33[2K\r[ %s ] %s\n", "CLIENT", "ERROR encrypting msg");
                print_prompt();
                return;
            }
        }

        Hash((const unsigned char*) msg, hash);
        copy(pkt->hash, hash, strlen(hash));
    } else {
        copy(pkt->reciever, msg + arr[0] + 1, (arr[1] - arr[0] + 1)); // priv msg
        
        if (idx != 2) {
            printf("\33[2K\r[ %s ] %s\n", "CLIENT", "please enter the command properly");
            print_prompt();
            return;
        }

        if (pkt->encrypted) {
            generate_random_iv(pkt->iv);
            pkt->len = encrypt((unsigned char*) (msg + arr[1] + 1), strlen(msg + arr[1] + 1),
                                key, pkt->iv, pkt->msg);
            if (pkt->len <= 0) {
                printf("\33[2K\r[ %s ] %s\n", "CLIENT", "ERROR encrypting msg");
                print_prompt();
                return;
            }
        }

        Hash((const unsigned char*) msg + arr[1] + 1, hash);
        copy(pkt->hash, hash, strlen(hash));
    }
}

void* user_message(void* args __attribute__((unused))) {
    pthread_setcancelstate(PTHREAD_CANCEL_ENABLE, NULL);
    pthread_setcanceltype(PTHREAD_CANCEL_DEFERRED, NULL);

    Packet* pkt = malloc(sizeof(Packet));
    if (!pkt) {
        printf("[!] Err%d: error allocating memory!\n", __LINE__);
        return NULL;
    }
    
    char buffer[BUFFER_SIZE];
    while (server_state) {
        print_prompt();

        if (fgets(buffer, sizeof(buffer) - 1, stdin) == NULL) {
            printf("[!] Err%d: reading input!\n", __LINE__);
            continue;
        }

        if (!memcmp(buffer, "/exit", 5)) {
            printf("[ CLIENT ] Exitting\n");
            kill_client(__LINE__);
        }

        buffer[strcspn(buffer, "\n")] = '\0';
        if (!strlen(buffer)) continue;
        create_packet(pkt, buffer, 1); // change to encrypted afterwards

        if (write(fd, pkt, sizeof(Packet)) == -1) {
            printf("[!] Err%d: sending message to server!\n", __LINE__);
        }

        memset(pkt, 0, sizeof(Packet));
    }
    
    free(pkt);
    pthread_detach(pthread_self());
    return NULL;
}

int main(void) {
    signal(SIGINT, kill_client);

    fd = socket(AF_INET, SOCK_STREAM, 0);
    if (fd < 0) {
        printf("[!] Error60: creating socket!\n");
        return EXIT_FAILURE;
    }
    
    struct sockaddr_in client;
    client.sin_family = AF_INET;
    client.sin_port = htons(PORT);
    client.sin_addr.s_addr = inet_addr(IP);
    client.sin_zero[7] = '\0';
    if (connect(fd, (const struct sockaddr*) &client, (socklen_t) sizeof(client)) < 0) { 
        printf("[!] Error70: connecting to server!\n");
        close(fd);
        return EXIT_FAILURE;
    }
    
    printf("[>] Enter your username: ");
    fgets(user, sizeof(user), stdin);
    if (strlen(user) <= 0) {
        printf("[!] Err%d: Username cannot be empty!\n", __LINE__);
        return EXIT_FAILURE;
    }
    user[strcspn(user, "\n")] = '\0';
    get_key(key);

    Packet* login = malloc(sizeof(Packet));
    if (!login) {
        printf("[!] Err%d: error allocating memory!\n", __LINE__);
        return 1;
    }
    
    copy(login->sender, user, MSG_LEN);
    copy(login->reciever, "SERVER", MSG_LEN);

    if (write(fd, login, sizeof(Packet)) == -1) {
        printf("[!] Err%d: error sending message!\n", __LINE__);
        return 1;
    }

    free(login);
    pthread_create(&server_listener, NULL, &server_message, NULL);
    pthread_create(&user_listener, NULL, &user_message, NULL);

    pthread_join(server_listener, NULL);
    pthread_join(user_listener, NULL);

    printf("\n[#] Exitting gracefully!\n");
    return 0;
}
