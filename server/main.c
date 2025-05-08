#include <pthread.h>
#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <signal.h>
#include <sys/socket.h>
#include <unistd.h>
#include <netinet/in.h>
#include <arpa/inet.h>

#include "../AES/aes.h"

#define MAX 10
#define IP "127.0.0.1"
#define PORT 6000
#define BUFFER_SIZE 1024
#define MSG_LEN 32

typedef struct {
    int encrypted;
    char sender[MSG_LEN];
    char reciever[MSG_LEN];
    unsigned char iv[AES_BLOCK_SIZE];
    int msg_len;
    unsigned char msg[BUFFER_SIZE + MSG_LEN];
    char hash[MSG_LEN * 2];
} Packet;

typedef struct {
    int fd;
    char username[MSG_LEN];
} Client_info;

Client_info users[MAX];
int nClients = 0;
pthread_mutex_t mutex = PTHREAD_MUTEX_INITIALIZER;

void safe_strcpy(char* dst, const char* src, size_t dst_len) {
    if (!dst || !src || dst_len <= 0) return;
    size_t src_len = strlen(src);
    size_t len = (src_len < dst_len - 1) ? src_len : dst_len - 1;
    memcpy(dst, src, len);
    dst[len] = '\0';
}

// this is a debug function
void print_packet(Packet* pkt) {
    printf("-----------------------------------------------------\n");
    printf("sender -> %s\n", pkt->sender);
    printf("reciever -> %s\n", pkt->reciever);
    printf("msg -> %s\n", pkt->msg);
    printf("hash -> %s\n", pkt->hash);
    printf("-----------------------------------------------------\n");
}
// debug function ends
//
void broadcast(Packet* pkt) {
    print_packet(pkt);
    for (int i = 0; i < nClients; i++) {
        // doesnt send message to the one sending the message (ie checks if sender is in active client)
        if (!strncmp(pkt->sender, users[i].username, strlen(users[i].username))) continue;

        if (write(users[i].fd, pkt, sizeof(Packet)) < 0) {
            printf("[!] Err%d: Cannot send message to %s\n", __LINE__, users[i].username);
        }
    }
}

void unicast(Packet* pkt){
    int flag = 0;
    for (int i = 0; i < nClients; i++) { 
        if (memcmp(pkt->reciever, users[i].username, strlen(users[i].username))) continue;
        
        printf("sent packet to %s\n", pkt->reciever);
        flag = 1;
        if (write(users[i].fd, pkt, sizeof(Packet)) < 0)  {
            printf("[!] Err%d: Cannot send message to %s\n", __LINE__, users[i].username);
        }
    }
    
    if (flag) return;

    char msg[100], hash[MSG_LEN * 2];
    Packet* packet = malloc(sizeof(Packet));
    if (!packet) {
        printf("[!] Err%d: error allocating memory!\n", __LINE__);
        return;
    }
    
    packet->encrypted = 0;
    safe_strcpy(packet->sender, "SERVER", sizeof(packet->sender));
    safe_strcpy(packet->reciever, pkt->sender, sizeof(packet->reciever)); // send it to the sender
    memset(packet->iv, 0, MSG_LEN/2);
    snprintf(msg, sizeof(msg) - 1, "%s username doesn't exists!\n", pkt->reciever);
    msg[strcspn(msg, "\n")] = '\0';
    packet->msg_len = strlen(msg);
    safe_strcpy((char*) packet->msg, msg, sizeof(msg));
    Hash((const unsigned char*) msg, hash);
    safe_strcpy(packet->hash, hash, MSG_LEN * 2);
    // packet ends
    
    unicast(packet);
    free(packet);
}

void remove_user(int idx) {
    char msg[100], hash[MSG_LEN * 2];
    
    // Packet starts
    Packet* server_pkt = malloc(sizeof(Packet));
    if (!server_pkt) {
        printf("[!] Err%d: Cannot allocate memory!\n", __LINE__);
        return;
    }
    
    server_pkt->encrypted = 0;
    safe_strcpy(server_pkt->sender, "SERVER", sizeof(server_pkt->sender));
    safe_strcpy(server_pkt->reciever, "SERVER", sizeof(server_pkt->reciever));
    memset(server_pkt->iv, 0, MSG_LEN/2);
    snprintf(msg, sizeof(msg) - 1, "%s left the chat\n", users[idx].username);
    msg[strcspn(msg, "\n")] = '\0';
    server_pkt->msg_len = strlen(msg);
    safe_strcpy((char*) server_pkt->msg, msg, sizeof(msg));
    Hash((const unsigned char*) msg, hash);
    safe_strcpy(server_pkt->hash, hash, MSG_LEN * 2);
    // packet ends

    for (int i = idx; i < nClients; i++) {
        memset(users[i].username, 0, MSG_LEN);
        if (i + 1 >= nClients) continue;
        users[i].fd = users[i + 1].fd;
        safe_strcpy(users[i].username, users[i+1].username, MSG_LEN);
    }

    pthread_mutex_lock(&mutex);
    nClients--;
    pthread_mutex_unlock(&mutex);

    broadcast(server_pkt);
    free(server_pkt);
}

void disconnect(int signum __attribute__((unused))) {
    for (int i = 0; i < nClients; i++)
        close(users[i].fd);

    printf("\n[+] Gracefully closed all the connections\n");
    exit(EXIT_SUCCESS);
}

void* handle_connection(void* args) {
    int clientfd = *((int*) args);

    Packet* pkt = malloc(sizeof(Packet));
    if (!pkt) {
        printf("[!] Err%d: error Allocating memory!\n", __LINE__);
        return NULL;
    }

    if (read(clientfd, pkt, sizeof(Packet)) < 0) {
        printf("[!] Err%d: Reading username from client!\n", __LINE__);
        return NULL;
    }

    // add user to clients
    users[nClients].fd = clientfd;
    safe_strcpy(users[nClients].username, pkt->sender, MSG_LEN);
    
    pthread_mutex_lock(&mutex);
    int idx = nClients++;
    pthread_mutex_unlock(&mutex);

    // send packet to all users except one joining
    Packet* server_pkt = malloc(sizeof(Packet));
    if (!server_pkt) {
        printf("[!] Err%d: Cannot Allocating Memory!\n", __LINE__);
        return NULL;
    }

    char msg[50], hash[MSG_LEN * 2];
    safe_strcpy(server_pkt->sender, "SERVER", sizeof(server_pkt->sender));
    safe_strcpy(server_pkt->reciever, "SERVER", sizeof(server_pkt->reciever));
    memset(server_pkt->iv, 0, MSG_LEN/2);
    snprintf(msg, sizeof(msg) - 2, "%s Joined the chat\n", pkt->sender); // 32 + 20
    msg[strcspn(msg, "\n")] = '\0';
    printf("%s\n", pkt->sender); // debug
    memcpy(server_pkt->msg, msg, sizeof(msg));
    server_pkt->msg_len = strlen(msg);
    Hash((const unsigned char*) msg, hash);
    safe_strcpy(server_pkt->hash, hash, MSG_LEN * 2);
    
    broadcast(server_pkt);
    free(server_pkt);
    free(pkt);
    // sent packet to all user;
    
    Packet* client_pkt = malloc(sizeof(Packet));
    if (!client_pkt) {
        printf("[!] Err%d: Cannot allocate memory!\n", __LINE__);
        return NULL;
    }

    while (read(clientfd, client_pkt, sizeof(Packet)) > 0) {
        if (strncmp(client_pkt->reciever, "SERVER", MSG_LEN)) {
            unicast(client_pkt);
        } else {
            broadcast(client_pkt);
        }
        
        memset(client_pkt, 0, sizeof(Packet));
    }
    
    free(client_pkt);
    remove_user(idx);
    close(clientfd);

    pthread_detach(pthread_self());
    return NULL;
}

int main(void) {
    signal(SIGINT, disconnect);

    int sockfd = socket(AF_INET, SOCK_STREAM, 0);
    if (!sockfd) {
        printf("[!] Error Creating a socket!\n");
        return 1;
    }

    int optval = 1;
    if (setsockopt(sockfd, SOL_SOCKET, SO_REUSEADDR, &optval, sizeof(optval)) < 0) {
        printf("[!] Error setting up set sockopt!\n");
        close(sockfd);
        return 1;
    }

    struct sockaddr_in server, clientaddr;
    server.sin_family = AF_INET;
    server.sin_port = htons(PORT);
    server.sin_addr.s_addr = inet_addr(IP);
    if (bind(sockfd, (struct sockaddr*) &server, sizeof(server)) < 0) {
        printf("[!] Error binding to %s\n", IP);
        close(sockfd);
        return 1;
    }

    if (listen(sockfd, MAX) < 0) {
        printf("[!] Error setting up listener!\n");
        close(sockfd);
        return 1;
    }

    printf("[#] Listening on %s:%d ..\n", IP, PORT);

    while (1) {
        int _size_sock = sizeof(clientaddr);
        int clientfd = accept(sockfd, (struct sockaddr*) &clientaddr, (socklen_t*) &_size_sock);
        if (clientfd < 0) {
            printf("[!] Error accepting client!\n");
            continue;
        }

        if ((nClients + 1) >= MAX) {
            char* msg = "[ SERVER ] Max Clients Reached!";
            if (write(clientfd, msg, strlen(msg)) == -1) {
                printf("[!] Err%d: sending message\n", __LINE__);
                continue;
            }
            printf("%s\n", msg);
            close(clientfd);
            continue;
        }

        pthread_t tid;
        pthread_create(&tid, NULL, &handle_connection, (void*) &clientfd);
        
        sleep(1);
    }
    
    return EXIT_SUCCESS;
}
