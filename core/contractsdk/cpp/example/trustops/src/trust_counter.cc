#include <iostream>
#include "xchain/xchain.h"
#include "xchain/trust_operators/trust_operators.h"

struct Counter : public xchain::Contract {};

DEFINE_METHOD(Counter, initialize) {
    xchain::Context* ctx = self.context();
    const std::string& creator = ctx->arg("creator");
    if (creator.empty()) {
        ctx->error("missing creator");
        return;
    }
    ctx->put_object("creator", creator);
    ctx->ok("initialize succeed");
}

DEFINE_METHOD(Counter, store) {
    xchain::Context* ctx = self.context();
    const std::string& key = ctx->arg("data");
    TrustOperators to(ctx->initiator());
    auto ok = to.store(ctx, 0, key);
    std::string debug = "done";
    if (!ok) {  debug = "error"; } 
    ctx->ok(debug);
}

DEFINE_METHOD(Counter, get) {
    xchain::Context* ctx = self.context();
    const std::string& key = ctx->arg("key");
    std::string value;
    if (ctx->get_object(key, &value)) {
        ctx->ok(value);
    } else {
        ctx->error("key not found");
    }
}
