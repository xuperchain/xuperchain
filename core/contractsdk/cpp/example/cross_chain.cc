#include "xchain/xchain.h"

struct CrossChain : public xchain::Contract {};

DEFINE_METHOD(CrossChain, initialize) {
    xchain::Context* ctx = self.context();
    ctx->ok("ok");
}

DEFINE_METHOD(CrossChain, verifyTx) {
    xchain::Context* ctx = self.context();
    const std::string& relay = ctx->arg("relay");
    const std::string& to_addr = ctx->initiator();
    const std::string& amount = ctx->arg("amount");
    xchain::Transaction tx;
    if (!tx.from_raw(ctx->arg("tx"))) {
        ctx->error("parse tx error");
        return;
    }
    xchain::Response resp;
    int ok = ctx->call("wasm", relay, "verifyTx",
                       {{"blockid", ctx->arg("blockid")},
                        {"txid", tx.txid},
                        {"txIndex", ctx->arg("txIndex")},
                        {"proofPath", ctx->arg("proofPath")}},
                       &resp);
    if (!ok) {
        ctx->error("call error");
        return;
    }
    if (resp.status != 200) {
        ctx->error("verify failed");
        return;
    }

    for (auto& tx_out : tx.tx_outputs) {
        if (tx_out.amount == amount && tx_out.to_addr == to_addr) {
            ctx->ok("ok");
            return;
        }
    }
    ctx->error("expect address and amount not found");
}
