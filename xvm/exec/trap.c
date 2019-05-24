#include "xvm.h"
extern void go_xvm_trap();

void init_go_trap() {
  wasm_rt_trap = go_xvm_trap;
}