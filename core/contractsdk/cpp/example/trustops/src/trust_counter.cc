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

// get a number by key
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

// store saves encrypted data directly
DEFINE_METHOD(Counter, store) {
    xchain::Context* ctx = self.context();
    std::string debug = "done";
    for(auto it = ctx->args().begin(); it != ctx->args().end(); ++it) {
        // args are already encrypted, just put
        auto ok = ctx->put_object(it->first, it->second);
        if (!ok) { debug = "error"; }
    }
    ctx->ok(debug);
}

// add adds two encrypted numbers
// input is in format {"l":"key_l", "r":"key_r", "o":"key_output"}
DEFINE_METHOD(Counter, add) {
    xchain::Context* ctx = self.context();
    TrustOperators to(ctx, 0);

    std::map<std::string, std::string> argsMap;
    for(auto it = ctx->args().begin(); it != ctx->args().end(); ++it) {
        std::string value;
        // get each encrypted value
        if (it->first != "o") {
            ctx->get_object(it->second, &value);
            argsMap[it->first] = value;
        } else {
            argsMap[it->first] = it->second;
        }
    }

    auto ok = to.ops("add", argsMap["l"], argsMap["r"], argsMap["o"]);
    std::string debug = "done";
    if (!ok) { debug = "error"; }
    ctx->ok(debug);
}

// sub substracts one number from another
// input is in format {"l":"key_l", "r":"key_r", "o":"key_output"}
DEFINE_METHOD(Counter, sub) {
   xchain::Context* ctx = self.context();
    TrustOperators to(ctx, 0);

    std::map<std::string, std::string> argsMap;
    for(auto it = ctx->args().begin(); it != ctx->args().end(); ++it) {
        std::string value;
        // get each encrypted value
        if (it->first != "o") {
            ctx->get_object(it->second, &value);
            argsMap[it->first] = value;
        } else {
            argsMap[it->first] = it->second;
        }
    }

    auto ok = to.ops("sub", argsMap["l"], argsMap["r"], argsMap["o"]);
    std::string debug = "done";
    if (!ok) { debug = "error"; }
    ctx->ok(debug);
}

// mul multiplies two encrypted numbers
// input is in format {"l":"key_l", "r":"key_r", "o":"key_output"}
DEFINE_METHOD(Counter, mul) {
   xchain::Context* ctx = self.context();
    TrustOperators to(ctx, 0);

    std::map<std::string, std::string> argsMap;
    for(auto it = ctx->args().begin(); it != ctx->args().end(); ++it) {
        std::string value;
        // get each encrypted value
        if (it->first != "o") {
            ctx->get_object(it->second, &value);
            argsMap[it->first] = value;
        } else {
            argsMap[it->first] = it->second;
        }
    }

    auto ok = to.ops("mul", argsMap["l"], argsMap["r"], argsMap["o"]);
    std::string debug = "done";
    if (!ok) { debug = "error"; }
    ctx->ok(debug);
}
