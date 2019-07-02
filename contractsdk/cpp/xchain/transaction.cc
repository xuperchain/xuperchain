//#include <google/protobuf/message.h>
#include "xchain/transaction.h"
//#include "xchain/contract.pb.h"
//#include "xchain/xchain.h"

namespace xchain {

Transaction::Transaction() {}

Transaction::~Transaction() {}

bool Transaction::init(pb::Transaction* pbtx) {
    _txid = pbtx->txid();
    _blockid = pbtx->blockid();
    _desc = pbtx->desc();
    _initiator = pbtx->initiator();
    for (int i = 0; i < pbtx->auth_require_size(); i++) {
        _auth_require.push_back(pbtx->auth_require(i));
    }
    
    for (int i = 0; i < pbtx->tx_inputs_size(); i++) {
        TxInput ti;
        ti.ref_txid = pbtx->tx_inputs(i).ref_txid();
        ti.ref_offset = pbtx->tx_inputs(i).ref_offset();
        ti.from_addr = pbtx->tx_inputs(i).from_addr();
        ti.amount = pbtx->tx_inputs(i).amount();
        
        _tx_inputs.push_back(ti); 
    }

    for (int i = 0; i < pbtx->tx_outputs_size(); i++) {
        TxOutput to;
        to.amount = pbtx->tx_outputs(i).amount();
        to.to_addr = pbtx->tx_outputs(i).to_addr();
        
        _tx_outputs.push_back(to); 
    }
    
    for (int i = 0; i < pbtx->tx_inputs_ext_size(); i++) {
        TxInputExt tie;
        tie.bucket = pbtx->tx_inputs_ext(i).bucket();
        tie.key = pbtx->tx_inputs_ext(i).key();
        tie.ref_txid = pbtx->tx_inputs_ext(i).ref_txid();
        tie.ref_offset = pbtx->tx_inputs_ext(i).ref_offset();
        
        _tx_inputs_ext.push_back(tie); 
    }

    for (int i = 0; i < pbtx->tx_outputs_ext_size(); i++) {
        TxOutputExt toe;
        toe.bucket = pbtx->tx_outputs_ext(i).bucket();
        toe.key = pbtx->tx_outputs_ext(i).key();
        toe.value = pbtx->tx_outputs_ext(i).value();
        
        _tx_outputs_ext.push_back(toe); 
    }

    return true;
}

}  // namespace xchain

