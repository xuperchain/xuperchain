#include <cstdio>
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

void Transaction::print() {
    printf("[Transaction]:\n");
    printf("txid: %s\n", txid.c_str());    
    printf("blockid: %s\n", blockid.c_str());    
    printf("desc: %s\n", desc.c_str());    
    printf("initiator: %s\n", initiator.c_str());    
    for (auto v : auth_require) {
        printf("auth_require: %s\n", v.c_str());
    }
    for (int i = 0; i < tx_inputs.size(); i++) {
        printf("[tx_input[%d]]: ref_txid: %s\n", i, tx_inputs[i].ref_txid.c_str());
        printf("[tx_input[%d]]: ref_offset: %d\n", i, tx_inputs[i].ref_offset);
        printf("[tx_input[%d]]: from_addr: %s\n", i, tx_inputs[i].from_addr.c_str());
        printf("[tx_input[%d]]: amount: %s\n", i, tx_inputs[i].amount.c_str());
    }
    for (int i = 0; i < tx_outputs.size(); i++) {
        printf("[tx_output[%d]]: ammount: %s\n", i, tx_outputs[i].amount.c_str());
        printf("[tx_output[%d]]: to_addr: %s\n", i, tx_outputs[i].to_addr.c_str());
    }
}

}  // namespace xchain

