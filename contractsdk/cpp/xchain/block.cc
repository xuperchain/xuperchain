#include "xchain/contract.pb.h"
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

}  // namespace xchain

