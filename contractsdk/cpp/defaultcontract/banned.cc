#include "xchain/xchain.h"

class Banned : public xchain::Contract {};

DEFINE_METHOD(Banned, initialize) {
    xchain::Context* ctx = self.context();
    ctx->ok("initialize success");
}

DEFINE_METHOD(Banned, add) {
    xchain::Context* ctx = self.context();
    const std::string key = ctx->arg("contract");
    const std::string value = "true";
    bool ret = ctx->put_object(key, value);
	std::string res = "add " + key;
	if (ret == false) {
		res = res + " failed";
	} else {
		res = res + " succeed";
	}
    ctx->ok(res);
}

DEFINE_METHOD(Banned, release) {
    xchain::Context* ctx = self.context();
    const std::string key = ctx->arg("contract");
    bool ret = ctx->delete_object(key);
	std::string res = "release " + key;
	if (ret == false) {
		res = res + " failed";
	} else {
		res = res + " succeed";
	}
    ctx->ok(res);
}

DEFINE_METHOD(Banned, get) {
    xchain::Context* ctx = self.context();
    const std::string key = ctx->arg("contract");
    std::string value;
    if (ctx->get_object(key, &value)) {
        ctx->ok(value);
    } else {
        ctx->error("contract not found");
    }
}
