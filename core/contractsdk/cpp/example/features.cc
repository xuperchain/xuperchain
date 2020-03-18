#include "xchain/xchain.h"

struct Features : xchain::Contract {};

DEFINE_METHOD(Features, initialize) {
    xchain::Context* ctx = self.context();
    ctx->ok("init");
}

DEFINE_METHOD(Features, logging) {
    xchain::Context* ctx = self.context();
    ctx->logf("log from contract");
    ctx->ok("ok");
}

DEFINE_METHOD(Features, put) {
    xchain::Context* ctx = self.context();
    for (const auto& elem : ctx->args()) {
        ctx->put_object(elem.first, elem.second);
    }
    ctx->ok("ok");
}

DEFINE_METHOD(Features, get) {
    xchain::Context* ctx = self.context();
    const std::string& key = ctx->arg("key");
    std::string value;
    if (ctx->get_object(key, &value)) {
        ctx->ok(value);
        return;
    }
    ctx->error("failed");
}

DEFINE_METHOD(Features, iterator) {
    xchain::Context* ctx = self.context();
    const std::string& start = ctx->arg("start");
    const std::string& limit = ctx->arg("limit");
    std::string ret;
    auto iter = ctx->new_iterator(start, limit);
    xchain::ElemType elem;
    while (iter->next()) {
        iter->get(&elem);
        ret += elem.first + ":" + elem.second + ", ";
    }
    ctx->ok(ret);
}

DEFINE_METHOD(Features, caller) {
    xchain::Context* ctx = self.context();
    ctx->ok(ctx->sender().get_name());
}

DEFINE_METHOD(Features, call) {
    xchain::Context* ctx = self.context();
    xchain::Response resp;
    const std::string contract = ctx->arg("contract");
    const std::string method = ctx->arg("method");
    bool ret = ctx->call("wasm", contract, method, ctx->args(), &resp);
    if (!ret) {
        ctx->error("call failed");
        return;
    }
    *ctx->mutable_response() = resp;
}
