#ifndef XCHAIN_BLOCK_H
#define XCHAIN_BLOCK_H

namespace xchain {
namespace contract {
namespace sdk {
    class Block;
}}}

namespace xchain {

class Block {
public:
    Block();
    virtual ~Block();
    void init(const xchain::contract::sdk::Block& pbblock);

public:
    std::string blockid;
    std::string pre_hash;
    std::string proposer;
    std::string sign;
    std::string pubkey;
    int64_t height;
    std::vector<std::string> txids;
    int32_t tx_count;
    bool in_trunk;
    std::string next_hash;
};
}  // namespace xchain

#endif
