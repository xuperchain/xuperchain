#include "xchain/block.h"

namespace xchain {

Block::Block() {}

Block::~Block() {}

bool Block::init(pb::Block pbblock) {
    blockid = pbblock.blockid();
    pre_hash = pbblock.pre_hash();
    proposer = pbblock.proposer();
    sign = pbblock.sign();
    pubkey = pbblock.pubkey();
    height = pbblock.height();
    tx_count = pbblock.tx_count();
    in_trunk = pbblock.in_trunk();
    next_hash = pbblock.next_hash();

    for (int i = 0; i < pbblock.transactions_size(); i++) {
        transactions.emplace_back(pbblock.transactions(i).txid());
    }
    
    return true;
}
}  // namespace xchain

