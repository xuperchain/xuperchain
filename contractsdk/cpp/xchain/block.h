#ifndef XCHAIN_BLOCK_H
#define XCHAIN_BLOCK_H

#include "xchain/contract.pb.h"

namespace xchain {

class Block {
public:
    Block();
    virtual ~Block();
    void init(pb::Block pbblock);

public:
    std::string blockid;
    std::string pre_hash;
    std::string proposer;
    std::string sign;
    std::string pubkey;
    int64_t height;
    std::vector<std::string> transactions;
    int32_t tx_count;
    bool in_trunk;
    std::string next_hash;
};
}  // namespace xchain

#endif
