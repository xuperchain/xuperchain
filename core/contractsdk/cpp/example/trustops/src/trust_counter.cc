#include <iostream>
#include "xchain/xchain.h"
#include "xchain/trust_operators/tf.pb.h"
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

DEFINE_METHOD(Counter, store) {
    xchain::Context* ctx = self.context();
    for(auto it = ctx->args().begin(); it != ctx->args().end(); ++it) {
        // args are already encrypted, just put
        ctx->put_object(it->first, it->second);
    }

    ctx->ok("done");
}

DEFINE_METHOD(Counter, debug) {
    xchain::Context* ctx = self.context();
    TrustOperators to(ctx->initiator());

    std::map<std::string, std::string> argsMap;
    for(auto it = ctx->args().begin(); it != ctx->args().end(); ++it) {
        std::string value;
        // get each encrypted value
        ctx->get_object(it->second, &value);
        argsMap[it->second] = value;
    }

    std::string args = TrustOperators::MapToString(argsMap);
    auto ok = to.debug(ctx, 0, args);
    ctx->ok(ok);
}

DEFINE_METHOD(Counter, add) {
    xchain::Context* ctx = self.context();
    TrustOperators to(ctx->initiator());

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
    std::string args = TrustOperators::MapToString(argsMap);
    auto ok = to.add(ctx, 0, args);
    ctx->ok(ok);
}

DEFINE_METHOD(Counter, sub) {
   xchain::Context* ctx = self.context();
    TrustOperators to(ctx->initiator());

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
    std::string args = TrustOperators::MapToString(argsMap);
    auto ok = to.sub(ctx, 0, args);
    ctx->ok(ok);
}


DEFINE_METHOD(Counter, mul) {
   xchain::Context* ctx = self.context();
    TrustOperators to(ctx->initiator());

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
    std::string args = TrustOperators::MapToString(argsMap);
    auto ok = to.mul(ctx, 0, args);
    ctx->ok(ok);
}
