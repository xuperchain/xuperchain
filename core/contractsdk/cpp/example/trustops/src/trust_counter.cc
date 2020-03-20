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

DEFINE_METHOD(Counter, increase) {
    xchain::Context* ctx = self.context();
    const std::string& key = ctx->arg("data");
    std::cout << "duanbing data: " << key.c_str() << std::endl;
    TrustOperators to(ctx->initiator());
    std::map<std::string, std::string> values;
    auto ok = to.store(0, key, &values);
    std::cout << "duanbing store ok: " << ok << std::endl; 
    std::string debug;
    for (auto &one : values) { 
        ctx->put_object(one.first, one.second);
        debug += one.first + ":" + one.second + ",";
    }
    std::cout << "duanbing increase done " << debug.c_str() << std::endl;
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
