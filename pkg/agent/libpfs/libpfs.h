#ifndef LIBPFS_H
#define LIBPFS_H

typedef long int int64_t;

typedef struct List {
    char** list;
    int size;
    int cap;
}List;

void cHello();
void printMessage(char* message);
int64_t pfs_chunkstream_read(int64_t chunkid, char* buf, int64_t len);
int64_t pfs_chunkstream_write(int64_t chunkid, const char* buf, int64_t len);
char* my_malloc(int64_t size);
int64_t my_int64(int i);
char* my_convert(void* ptr);
char** list_dir_wrapper(const char *path, int* len);
int64_t pfs_read_file(char* path, char* buf, int64_t len);
int64_t pfs_write_file(char* path, char* buf, int64_t len);
int getLenOfList(char** list);

#endif