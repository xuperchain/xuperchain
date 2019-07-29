#include "xchain/xchain.h"

struct BuiltinTypes : public xchain::Contract {};

DEFINE_METHOD(BuiltinTypes, initialize) {
    xchain::Context* ctx = self.context();
    ctx->ok("initialize succeed");
}

DEFINE_METHOD(BuiltinTypes, gettx) {
    xchain::Context* ctx = self.context();
    const std::string& txid = ctx->arg("txid");
    xchain::Transaction tx;
     
    if (ctx->query_tx(txid, &tx)) {
        tx.print();
        ctx->ok("Obtaining transaction succeed");
    } else {
        ctx->error("txid not found");
    }
}

DEFINE_METHOD(BuiltinTypes, getblock) {
    xchain::Context* ctx = self.context();
    const std::string& blockid = ctx->arg("blockid");
    xchain::Block block;

    if (ctx->query_block(blockid, &block)) {
        block.print();
        ctx->ok("Obtaining block succeed");
    } else {
        ctx->error("blockid not found");
    }
}
