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

#include "xvm.h"

#include <assert.h>
#include <stdarg.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include <sys/mman.h>
#include <dlfcn.h>
#include <stdio.h>

#define PAGE_SIZE 65536

void (*wasm_rt_trap)(wasm_rt_trap_t code) = NULL;

static void* xvm_malloc(size_t size);
static void* xvm_realloc(void* ptr, size_t size);
static void* xvm_free(void* ptr);

static void _wasm_rt_trap(wasm_rt_trap_t code) {
  fprintf(stderr, "panic:%d\n", code);
  abort();
}

struct FuncType {
  wasm_rt_type_t* params;
  wasm_rt_type_t* results;
  uint32_t param_count;
  uint32_t result_count;
};

static bool func_types_are_equal(struct FuncType* a, struct FuncType* b) {
  if (a->param_count != b->param_count || a->result_count != b->result_count)
    return 0;
  int i;
  for (i = 0; i < a->param_count; ++i)
    if (a->params[i] != b->params[i])
      return 0;
  for (i = 0; i < a->result_count; ++i)
    if (a->results[i] != b->results[i])
      return 0;
  return 1;
}

void free_func_type(struct FuncType* ftype) {
  if (ftype->params != NULL) {
    xvm_free(ftype->params);
  }
  if (ftype->results != NULL) {
    xvm_free(ftype->results);
  }
}

static uint32_t wasm_rt_register_func_type(void* context,
                                    uint32_t param_count,
                                    uint32_t result_count,
                                    ...) {
  
  xvm_code_t* code = context;
  struct FuncType func_type;
  func_type.param_count = param_count;
  func_type.params = xvm_malloc(param_count * sizeof(wasm_rt_type_t));
  func_type.result_count = result_count;
  func_type.results = xvm_malloc(result_count * sizeof(wasm_rt_type_t));

  va_list args;
  va_start(args, result_count);

  uint32_t i;
  for (i = 0; i < param_count; ++i)
    func_type.params[i] = va_arg(args, wasm_rt_type_t);
  for (i = 0; i < result_count; ++i)
    func_type.results[i] = va_arg(args, wasm_rt_type_t);
  va_end(args);

  for (i = 0; i < code->func_type_count; ++i) {
    if (func_types_are_equal(&code->func_types[i], &func_type)) {
      xvm_free(func_type.params);
      xvm_free(func_type.results);
      return i + 1;
    }
  }

  uint32_t idx = code->func_type_count++;
  code->func_types = xvm_realloc(code->func_types, code->func_type_count * sizeof(struct FuncType));
  code->func_types[idx] = func_type;
  return idx + 1;
}

static void wasm_rt_allocate_memory(void* context,
                             wasm_rt_memory_t* memory,
                             uint32_t initial_pages,
                             uint32_t max_pages) {
  if (initial_pages == 0) {
    initial_pages = 1;
  }
  memory->pages = initial_pages;
  memory->max_pages = max_pages;
  memory->size = initial_pages * PAGE_SIZE;
  if (memory->size != 0) {
    memory->data = mmap(0, memory->size, PROT_READ|PROT_WRITE,
    MAP_PRIVATE|MAP_ANONYMOUS, -1, 0);
    if (memory->data == MAP_FAILED) {
      xvm_raise(TRAP_NO_MEMORY);
    }
  }
  xvm_context_t* ctx = context;
  ctx->mem = memory;
}

static void wasm_rt_free_memory(wasm_rt_memory_t* mem) {
    munmap(mem->data, mem->size);
}

static uint32_t wasm_rt_grow_memory(void* context, wasm_rt_memory_t* memory, uint32_t delta) {
  // do not support grow memory
  wasm_rt_trap(WASM_RT_TRAP_OOB);

  uint32_t old_pages = memory->pages;
  uint32_t new_pages = memory->pages + delta;
  if (new_pages < old_pages || new_pages > memory->max_pages) {
    return (uint32_t)-1;
  }
  memory->pages = new_pages;
  memory->size = new_pages * PAGE_SIZE;
  memory->data = xvm_realloc(memory->data, memory->size);
  memset(memory->data + old_pages * PAGE_SIZE, 0, delta * PAGE_SIZE);
  return old_pages;
}

static void wasm_rt_allocate_table(void* context,
                            wasm_rt_table_t* table,
                            uint32_t elements,
                            uint32_t max_elements) {
  if (elements == 0) {
    elements = 10;
  }
  table->size = elements;
  table->max_size = max_elements;
  if (table->size != 0) {
    table->data = xvm_malloc(table->size*sizeof(wasm_rt_elem_t));
  }
  xvm_context_t* ctx = context;
  ctx->table = table;
}

static void* wasm_rt_malloc(void* context, uint32_t size) {
  return xvm_malloc(size);
}

static wasm_rt_func_handle_t wasm_rt_resolve_func(void* context, char* module, char* name) {
  xvm_code_t* code = context;
  return code->resolver.resolve_func(code->resolver.env, module, name);
}

static uint32_t wasm_rt_call_func(void* context, wasm_rt_func_handle_t hfunc, uint32_t* params, uint32_t param_len) {
  xvm_context_t* ctx = context;
  xvm_code_t* code = ctx->code;
  void* env = code->resolver.env;
  return code->resolver.call_func(env, hfunc, ctx, params, param_len);
}

// TODO: 每个context有单独的全局变量设置?
static double wasm_rt_resolve_global(void* context, char* module, char* name) {
  xvm_context_t* ctx = context;
  return ctx->code->resolver.resolve_global(ctx->code->resolver.env, module, name);
}

static wasm_rt_ops_t make_wasm_rt_ops() {
  wasm_rt_ops_t ops = {0};
  ops.wasm_rt_register_func_type = wasm_rt_register_func_type;
  ops.wasm_rt_allocate_memory = wasm_rt_allocate_memory;
  ops.wasm_rt_grow_memory = wasm_rt_grow_memory;
  ops.wasm_rt_allocate_table = wasm_rt_allocate_table;
  ops.wasm_rt_malloc = wasm_rt_malloc;
  ops.wasm_rt_resolve_func = wasm_rt_resolve_func;
  ops.wasm_rt_call_func = wasm_rt_call_func;
  ops.wasm_rt_resolve_global = wasm_rt_resolve_global;
  if (wasm_rt_trap == NULL) {
    ops.wasm_rt_trap = _wasm_rt_trap;
  } else {
    ops.wasm_rt_trap = wasm_rt_trap;
  }
  return ops;
}

/*
 * xvm_code_t相关代码
 */
xvm_code_t* xvm_new_code(char* module_path, xvm_resolver_t resolver) {
  void* dlhandle = dlopen(module_path, RTLD_NOW|RTLD_LOCAL);
  if (dlhandle == NULL) {
    fprintf(stderr, "dlopen:%s\n", dlerror());
    return NULL;
  }
  void (*init_rt_ops_func)(void*) = dlsym(dlhandle, "init_rt_ops");
  if (init_rt_ops_func == NULL) {
    fprintf(stderr, "function init_rt_ops not found\n");
    dlclose(dlhandle);
    return NULL;
  }
  wasm_rt_ops_t ops = make_wasm_rt_ops();
  (*init_rt_ops_func)(&ops);

  void (*init_func_types)(void*) = dlsym(dlhandle, "init_func_types");
  if (init_func_types == NULL) {
    fprintf(stderr, "function init_func_types not found\n");
    dlclose(dlhandle);
    return NULL;
  }

  void (*init_import_funcs)(void*) = dlsym(dlhandle, "init_import_funcs");
  if (init_import_funcs == NULL) {
    fprintf(stderr, "function init_import_funcs not found\n");
    dlclose(dlhandle);
    return NULL;
  }
  xvm_code_t* code = xvm_malloc(sizeof(xvm_code_t));
  code->dlhandle = dlhandle;
  code->resolver = resolver;
  (*init_func_types)(code);
  (*init_import_funcs)(code);
  return code;
}

void xvm_release_code(xvm_code_t* code) {
  if (code->func_types != NULL) {
    int i = 0;
    for (; i<code->func_type_count; i++) {
      free_func_type(code->func_types + i);
    }
    xvm_free(code->func_types);
  }
  if (code->dlhandle != NULL) {
    dlclose(code->dlhandle);
  }
  xvm_free(code);
  memset((void*)code, 0, sizeof(xvm_code_t));
}

/*
 * xvm_context_t相关代码
 */

// FIXME: 修改访问暴露变量的方式
struct _wasm_rt_handle_t {
  void* user_ctx;
  wasm_rt_gas_t gas;
};

xvm_context_t* xvm_new_context(xvm_code_t* code) {
  xvm_context_t* ctx = xvm_malloc(sizeof(xvm_context_t));
  ctx->code = code;
  return ctx;
}

int xvm_init_context(xvm_context_t* ctx, xvm_code_t* code) {
  void*(*new_handle_func)(void*) = dlsym(code->dlhandle, "new_handle");
  if (new_handle_func == NULL) {
    fprintf(stderr, "new_handle function not found\n");
    return 0;
  }
  ctx->code = code;
  ctx->module_handle = new_handle_func(ctx);
  return 1;
}

void xvm_release_context(xvm_context_t* ctx) {
  if (ctx->mem != NULL) {
    wasm_rt_free_memory(ctx->mem);
  }
  if (ctx->table != NULL) {
    xvm_free(ctx->table->data);
  }
  if (ctx->module_handle != NULL) {
    xvm_free(ctx->module_handle);
  }
  memset((void*)ctx, 0, sizeof(xvm_context_t));
}

uint32_t xvm_call(xvm_context_t* ctx, char* name, int64_t* params, int64_t param_len, wasm_rt_gas_t* gas, int64_t* ret) {
  void* func = dlsym(ctx->code->dlhandle, name);
  if (func == NULL) {
    return 0;
  }
  struct _wasm_rt_handle_t* _handle = ctx->module_handle;
  if (gas != NULL) {
    _handle->gas.limit = gas->limit;
  }
  int64_t (*real_func)(void*, int64_t*, int64_t) = func;
  *ret = real_func(ctx->module_handle, params, param_len);
  if (gas != NULL) {
    gas->used = _handle->gas.used;
  }
  return 1;
}

static void* xvm_malloc(size_t size) {
  void* ptr = calloc(size, 1);
  if (ptr == NULL) {
    xvm_raise(TRAP_NO_MEMORY);
  }
  return ptr;
}

static void* xvm_realloc(void* ptr, size_t size) {
  void* new_ptr = realloc(ptr, size);
  if (new_ptr == NULL) {
    xvm_raise(TRAP_NO_MEMORY);
  }
  return new_ptr;
}

static void* xvm_free(void* ptr) {
  free(ptr);
}
