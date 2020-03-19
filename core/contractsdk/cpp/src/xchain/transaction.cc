#include "xchain/transaction.h"
#include <memory>
#include "xchain/contract.pb.h"

extern "C" int xvm_make_tx(const char* txptr, int txlen, char** outpptr);

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
        tx_inputs.emplace_back(
            pbtx.tx_inputs(i).ref_txid(), pbtx.tx_inputs(i).ref_offset(),
            pbtx.tx_inputs(i).from_addr(), pbtx.tx_inputs(i).amount());
    }

    for (int i = 0; i < pbtx.tx_outputs_size(); i++) {
        tx_outputs.emplace_back(pbtx.tx_outputs(i).amount(),
                                pbtx.tx_outputs(i).to_addr());
    }
}

bool Transaction::from_raw(const std::string& raw_tx) {
    char* buf = NULL;
    int ret = xvm_make_tx((const char*)&raw_tx[0], raw_tx.size(), &buf);
    if (ret != 0) {
        return false;
    }
    std::unique_ptr<char> mem_guard(buf);
    pb::Transaction tx;
    if (!tx.ParseFromString(buf)) {
        return false;
    }
    init(tx);
    return true;
}

}  // namespace xchain
