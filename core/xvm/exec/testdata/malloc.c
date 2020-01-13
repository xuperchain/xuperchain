#include <malloc.h>
#include <emscripten/emscripten.h>

extern void print(char*);

EMSCRIPTEN_KEEPALIVE void run() {
    char* mem = malloc(1024);
    print(mem);
}