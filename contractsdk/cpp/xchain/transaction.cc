#include "xchain/transaction.h"

namespace xchain {

Transaction::Transaction() {}

Transaction::~Transaction() {}

void Transaction::init(pb::TransactionSDK pbtx) {
    txid = pbtx.txid();
    blockid = pbtx.blockid();
    desc = pbtx.desc();
    initiator = pbtx.initiator();
    for (int i = 0; i < pbtx.auth_require_size(); i++) {
        auth_require.emplace_back(pbtx.auth_require(i));
    }
    
    for (int i = 0; i < pbtx.tx_inputs_size(); i++) {
        TxInput ti;
        ti.ref_txid = pbtx.tx_inputs(i).ref_txid();
        ti.ref_offset = pbtx.tx_inputs(i).ref_offset();
        ti.from_addr = pbtx.tx_inputs(i).from_addr();
        ti.amount = pbtx.tx_inputs(i).amount();
        
        tx_inputs.emplace_back(ti); 
    }

    for (int i = 0; i < pbtx.tx_outputs_size(); i++) {
        TxOutput to;
        to.amount = pbtx.tx_outputs(i).amount();
        to.to_addr = pbtx.tx_outputs(i).to_addr();
        
        tx_outputs.emplace_back(to); 
    }
}

}  // namespace xchain

