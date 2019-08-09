#include <cstdio>
#include <cinttypes>
#include "xchain/xchain.h"


struct BuiltinTypes : public xchain::Contract {};

void print_tx(xchain::Transaction t);
void print_block(xchain::Block b);

DEFINE_METHOD(BuiltinTypes, initialize) {
    xchain::Context* ctx = self.context();
    ctx->ok("initialize succeed");
}

DEFINE_METHOD(BuiltinTypes, gettx) {
    xchain::Context* ctx = self.context();
    const std::string& txid = ctx->arg("txid");
    xchain::Transaction tx;
     
    if (ctx->query_tx(txid, &tx)) {
        print_tx(tx);
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
        print_block(block);
        ctx->ok("Obtaining block succeed");
    } else {
        ctx->error("blockid not found");
    }
}

void print_block(xchain::Block b) {
    printf("[Block]:\n");
    printf("blockid: %s\n", b.blockid.c_str());    
    printf("pre_hash: %s\n", b.pre_hash.c_str());    
    printf("proposer: %s\n", b.proposer.c_str());    
    printf("sign: %s\n", b.sign.c_str());    
    printf("pubkey: %s\n", b.pubkey.c_str());    
    printf("height: %" PRId64"\n", b.height);    
    for (auto v : b.txids) {
        printf("txid: %s\n", v.c_str());
    }
    printf("tx_count: %d\n", b.tx_count);    
    printf("in_trunk: %d\n", b.in_trunk);    
    printf("next_hash: %s\n", b.next_hash.c_str());    
}

void print_tx(xchain::Transaction t) {
    printf("[Transaction]:\n");
    printf("txid: %s\n", t.txid.c_str());    
    printf("blockid: %s\n", t.blockid.c_str());    
    printf("desc: %s\n", t.desc.c_str());    
    printf("initiator: %s\n", t.initiator.c_str());    
    for (auto v : t.auth_require) {
        printf("auth_require: %s\n", v.c_str());
    }
    for (int i = 0; i < t.tx_inputs.size(); i++) {
        printf("[tx_input[%d]]: ref_txid: %s\n", i, t.tx_inputs[i].ref_txid.c_str());
        printf("[tx_input[%d]]: ref_offset: %d\n", i, t.tx_inputs[i].ref_offset);
        printf("[tx_input[%d]]: from_addr: %s\n", i, t.tx_inputs[i].from_addr.c_str());
        printf("[tx_input[%d]]: amount: %s\n", i, t.tx_inputs[i].amount.c_str());
    }
    for (int i = 0; i < t.tx_outputs.size(); i++) {
        printf("[tx_output[%d]]: ammount: %s\n", i, t.tx_outputs[i].amount.c_str());
        printf("[tx_output[%d]]: to_addr: %s\n", i, t.tx_outputs[i].to_addr.c_str());
    }
}


