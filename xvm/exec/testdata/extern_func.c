#include <emscripten/emscripten.h>

extern void print(char*);

EMSCRIPTEN_KEEPALIVE void run() {
    print("hello world");
}