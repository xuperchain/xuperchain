#include "xvm.h"

extern wasm_rt_func_handle_t xvm_resolve_func(void* env, char* module, char* name);
extern double xvm_resolve_global(void* env, char* module, char* name);
extern uint32_t xvm_call_func(void* env, wasm_rt_func_handle_t handle, xvm_context_t* ctx, uint32_t* params, uint32_t param_len);

xvm_resolver_t make_resolver_t(void* env) {
	xvm_resolver_t r;
	r.env = env;
	r.resolve_func = xvm_resolve_func;
	r.resolve_global = xvm_resolve_global;
	r.call_func = xvm_call_func;
	return r;
}
