#include "xchain/contract.pb.h"
#include "xchain/transaction.h"

namespace pb = xchain::contract::sdk;

namespace xchain {

Transaction::Transaction() {}

Transaction::~Transaction() {}

void Transaction::init(const pb::Transaction& pbtx) {
    txid = pbtx.txid();
    blockid = pbtx.blockid();
    desc = pbtx.desc();
    initiator = pbtx.initiator();
    for (int i = 0; i < pbtx.auth_require_size(); i++) {
        auth_require.emplace_back(pbtx.auth_require(i));
    }
    
    for (int i = 0; i < pbtx.tx_inputs_size(); i++) {
        tx_inputs.emplace_back(pbtx.tx_inputs(i).ref_txid(), pbtx.tx_inputs(i).ref_offset(),
            pbtx.tx_inputs(i).from_addr(), pbtx.tx_inputs(i).amount()); 
    }

    for (int i = 0; i < pbtx.tx_outputs_size(); i++) {
        tx_outputs.emplace_back(pbtx.tx_outputs(i).amount(), pbtx.tx_outputs(i).to_addr()); 
    }
}

}  // namespace xchain

