#include <stdio.h>
#include <stdlib.h>
#include <dirent.h>
#include <sys/types.h>
#include <fcntl.h>
#include <errno.h>
#include <unistd.h>
#include "libpfs.h"

void cHello() {
    printf("Hello from C!\n");
}

void printMessage(char* message) {
    printf("pfs read %s\n", message);
}

int64_t pfs_read_file(char* path, char* buf, int64_t len) {
    int fd = open (path, O_RDONLY);
    int64_t res_len = read (fd, buf, len);
    return res_len;
}

int64_t pfs_write_file(char* path, char* buf, int64_t len) {
    int fd = open(path, O_WRONLY | O_CREAT, 0644);
    int64_t res_len = write (fd, buf, len);
    return res_len;
}

int64_t pfs_chunkstream_read(int64_t chunkid, char* buf, int64_t len) {
    for(int i = 0; i < len; i++){
        buf[i] = i % 255;
    }
    return len;
}

int64_t pfs_chunkstream_write(int64_t chunkid, const char* buf, int64_t len){
    printf("pfs write chunk:");
    for(int i = 0; i < len; i++){
        printf(" %d", buf[i]);
    }
    printf("\n");
    return len;
}

void resizeList(List* list){
    char** newlist = malloc(list->cap * 2 * sizeof(char*));
    memcpy(newlist, list->list, list->cap);
    list->cap = list->cap * 2;
    free(list->list);
    list->list = newlist;
}

void initList(List* list){
    list->list = malloc(512 * sizeof(char*));
    list->size = 0;
    list->cap = 512;
}

void addFile(List* list, const char *path){
    if(list->size == list->cap){
        resizeList(list);
    }
    int len = strlen(path) + 1;
    char* pathcopy = malloc(len);
    memcpy(pathcopy, path, len);
    list->list[list->size++] = pathcopy;
}

char** list_dir_wrapper(const char *path, int* len){
    List l;
    initList(&l);
    list_dir(path, &l);
    *len = l.size;
    return l.list;
}

int getLenOfList(char** list){
    return sizeof(list) / sizeof(char*);
}

void list_dir(const char *path, List* list)
{
    struct dirent *entry;
    DIR *dir = opendir(path);
    if (dir == NULL) {
        return;
    }

    while ((entry = readdir(dir)) != NULL) {
        if(entry->d_type == 8){
            char file[100] = {'\0'};
            sprintf(file, "%s/%s", path, entry->d_name);
            printf("file: %s\n", file);
            addFile(list, file);
        } else if(entry->d_type == 4){
            // ignore . and ..
            if((strlen(entry->d_name) == 1 && entry->d_name[0] == '.') ||
            (strlen(entry->d_name) == 2 && entry->d_name[0] == '.' && entry->d_name[1] == '.')){
                continue;
            }
            char subdir[80];
            sprintf(subdir, "%s/%s", path, entry->d_name);
            printf("dir: %s\n", subdir);
            list_dir(subdir, list);
        }
    }

    closedir(dir);
}

char* my_malloc(int64_t size) {
    return (char*)malloc(size);
}

char* my_convert(void* ptr) {
    return (char*)ptr;
}

int64_t my_int64(int i) {
    return (int64_t)i;
}