#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <sys/mman.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <unistd.h>

int main(int argc, char **argv) {
    if (argc < 2) {
        fprintf(stderr, "usage: %s <shellcode.bin>\n", argv[0]);
        return 1;
    }

    int fd = open(argv[1], O_RDONLY);
    if (fd < 0) { perror("open"); return 1; }

    struct stat st;
    if (fstat(fd, &st) < 0) { perror("fstat"); return 1; }
    size_t n = (size_t)st.st_size;

    unsigned char *buf = malloc(n);
    if (!buf) { perror("malloc"); return 1; }
    if (read(fd, buf, n) != (ssize_t)n) { perror("read"); return 1; }
    close(fd);

    void *mem = mmap(NULL, n, PROT_READ | PROT_WRITE | PROT_EXEC,
                     MAP_PRIVATE | MAP_ANONYMOUS, -1, 0);
    if (mem == MAP_FAILED) { perror("mmap"); return 1; }
    memcpy(mem, buf, n);
    free(buf);

    /* Piclang espera stackHint en RDI: un puntero al stack actual.
       El runtime lo usa para inicializar su heap. */
    volatile uint64_t anchor = 0;
    void *stackHint = (void *)&anchor;

    fprintf(stderr, "[runner] %zu bytes en %p, stackHint=%p\n", n, mem, stackHint);
    fflush(stderr);

    long ret = ((long (*)(void *))mem)(stackHint);

    printf("[runner] shellcode retorno = %ld (0x%lx)\n", ret, ret);
    return 0;
}

