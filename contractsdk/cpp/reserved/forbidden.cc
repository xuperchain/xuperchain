#include "xchain/xchain.h"

class Forbidden : public xchain::Contract {};

DEFINE_METHOD(Forbidden, initialize) {
    xchain::Context* ctx = self.context();
    ctx->ok("initialize success");
}

DEFINE_METHOD(Forbidden, forbid) {
    xchain::Context* ctx = self.context();
    // txid to be forbidden
    const std::string key = ctx->arg("txid");
    if ("" == key) {
        ctx->error("forbid failed");
        return;
    }
    // the reason to be forbidden
    const std::string value = ctx->arg("value");
    bool ret = ctx->put_object(key, value);
    if (!ret) {
        ctx->error("forbid failed");
        return;
    }
    ctx->ok("forbid success");
}

DEFINE_METHOD(Forbidden, unforbid) {
    xchain::Context* ctx = self.context();
    // txid to be unforbidden
    const std::string key = ctx->arg("txid");
    if ("" == key) {
        ctx->error("unforbid failed");
        return;
    }
    bool ret = ctx->delete_object(key);
    if (!ret) {
        ctx->error("unforbid failed");
        return;
    }
    ctx->ok("unforbid success");
}

DEFINE_METHOD(Forbidden, get) {
    xchain::Context* ctx = self.context();
    // check if txid has been forbidden
    const std::string key = ctx->arg("txid");
    std::string value;
    bool ret = ctx->get_object(key, &value);
    if (ret) {
        ctx->ok("txid has been forbidden");
        return;
    }
    ctx->error("txid has not been forbidden");
}
