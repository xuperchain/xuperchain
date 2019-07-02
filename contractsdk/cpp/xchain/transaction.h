#ifndef XCHAIN_TRANSACTION_H
#define XCHAIN_TRANSACTION_H

#include "xchain/contract.pb.h"
#include "xchain/xchain.h"
#include "xchain/xchain.pb.h"

namespace xchain {

struct TxInput {
    std::string ref_txid;
    int32_t ref_offset;
    std::string from_addr;
    std::string amount;
};

struct TxOutput {
    std::string amount;
    std::string to_addr;
};

struct TxInputExt {
    std::string bucket;
    std::string key;
    std::string ref_txid;
    int32_t ref_offset;
};

struct TxOutputExt {
    std::string bucket;
    std::string key;
    std::string value;
};

class Transaction {
public:
    Transaction();
    virtual ~Transaction();
    bool init(pb::Transaction* pbtx);

public:
    std::string _txid;
    std::string _blockid;
    std::string _desc;
    std::string _initiator;
    std::vector<std::string> _auth_require;
    std::vector<TxInput> _tx_inputs;
    std::vector<TxOutput> _tx_outputs;
    std::vector<TxInputExt> _tx_inputs_ext;
    std::vector<TxOutputExt> _tx_outputs_ext;
};

}  // namespace xchain

#endif
