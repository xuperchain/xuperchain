#include "xchain/block.h"

namespace pb = xchain::contract::sdk;

namespace xchain {

Block::Block() {}

Block::~Block() {}

void Block::init(const pb::Block& pbblock) {
    blockid = pbblock.blockid();
    pre_hash = pbblock.pre_hash();
    proposer = pbblock.proposer();
    sign = pbblock.sign();
    pubkey = pbblock.pubkey();
    height = pbblock.height();
    tx_count = pbblock.tx_count();
    in_trunk = pbblock.in_trunk();
    next_hash = pbblock.next_hash();

    for (int i = 0; i < pbblock.txids_size(); i++) {
        txids.emplace_back(pbblock.txids(i));
    }
}

void Block::print() {
    printf("[Block]:\n");
    printf("blockid: %s\n", blockid.c_str());    
    printf("pre_hash: %s\n", pre_hash.c_str());    
    printf("proposer: %s\n", proposer.c_str());    
    printf("sign: %s\n", sign.c_str());    
    printf("pubkey: %s\n", pubkey.c_str());    
    printf("height: %ld\n", height);    
    for (auto v : txids) {
        printf("txid: %s\n", v.c_str());
    }
    printf("tx_count: %d\n", tx_count);    
    printf("in_trunk: %d\n", in_trunk);    
    printf("next_hash: %s\n", next_hash.c_str());    
}

}  // namespace xchain

