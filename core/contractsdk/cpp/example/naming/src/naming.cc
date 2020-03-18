#include "xchain/xchain.h"

struct Naming : public xchain::Contract {};


#define CHECK_ARG(argKey) \
	std::string argKey = ctx->arg(#argKey);\
	if (argKey == "") { \
		ctx->error("missing required arg: " #argKey); \
		return;\
	}

DEFINE_METHOD(Naming, initialize) {
    xchain::Context* ctx = self.context();
    ctx->ok("initialize succeed");
}

DEFINE_METHOD(Naming, RegisterChain) {
    xchain::Context* ctx = self.context();
	CHECK_ARG(name);
	CHECK_ARG(chainMeta);
    ctx->ok(name);
}

DEFINE_METHOD(Naming, UpdateChain) {
    xchain::Context* ctx = self.context();
	CHECK_ARG(name);
	CHECK_ARG(chainMeta);
    ctx->ok("todo");
}

DEFINE_METHOD(Naming, GetChainMeta) {
    xchain::Context* ctx = self.context();
	CHECK_ARG(name);
    ctx->ok("todo");
}

DEFINE_METHOD(Naming, Resolve) {
    xchain::Context* ctx = self.context();
	CHECK_ARG(name);
}


DEFINE_METHOD(Naming, AddEndorsor) {
    xchain::Context* ctx = self.context();
	CHECK_ARG(name);
	CHECK_ARG(info);
    ctx->ok("todo");
}

DEFINE_METHOD(Naming, UpdateEndorsor) {
    xchain::Context* ctx = self.context();
	CHECK_ARG(name);
	CHECK_ARG(address);
	CHECK_ARG(info);
    ctx->ok("todo");
}


DEFINE_METHOD(Naming, DeleteEndorsor) {
    xchain::Context* ctx = self.context();
	CHECK_ARG(name);
	CHECK_ARG(address);
    ctx->ok("todo");
}

