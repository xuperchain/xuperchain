/*
 * Copyright 2018 WebAssembly Community Group participants
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

#ifndef XVM_H_
#define XVM_H_

#include "wasm-rt.h"

#ifdef __cplusplus
extern "C" {
#endif

struct xvm_context_t;
struct xvm_code_t;
struct xvm_resolver_t;

// Override this variable to define trap function
extern void (*wasm_rt_trap)(wasm_rt_trap_t code);

#define TRAP_NO_MEMORY "run out of memory"
// xvm_raise is used to raise internal trap error
extern void xvm_raise(char* msg);

typedef struct xvm_resolver_t {
  void* env;
  void* (*resolve_func)(void* env, char* module, char* name);
  int64_t (*resolve_global)(void* env, char* module, char* name);
  uint32_t (*call_func)(void* env, wasm_rt_func_handle_t hfunc, struct xvm_context_t* ctx,
                        uint32_t* params, uint32_t param_len);
} xvm_resolver_t;

struct FuncType;
typedef struct xvm_code_t {
  void* dlhandle;
  struct FuncType* func_types;
  uint32_t func_type_count;
  xvm_resolver_t resolver;
  void* (*new_handle_func)(void*);
  void (*init_func_types)(void*);
  void (*init_import_funcs)(void*);
} xvm_code_t;

xvm_code_t* xvm_new_code(char* module_path, xvm_resolver_t resolver);
int xvm_init_code(xvm_code_t* code);
void xvm_release_code(xvm_code_t* code);

typedef struct xvm_context_t {
  xvm_code_t* code;
  void* module_handle;
  wasm_rt_memory_t* mem;
  wasm_rt_table_t* table;
} xvm_context_t;

int xvm_init_context(xvm_context_t* ctx, xvm_code_t* code, uint64_t gas_limit);
void xvm_release_context(xvm_context_t* ctx);
uint32_t xvm_call(xvm_context_t* ctx, char* name, int64_t* params, int64_t param_len, int64_t* ret);
uint32_t xvm_mem_static_top(xvm_context_t* ctx);
void xvm_reset_gas_used(xvm_context_t* ctx);
uint64_t xvm_gas_used(xvm_context_t* ctx);

#ifdef __cplusplus
}
#endif

#endif // XVM_H_
