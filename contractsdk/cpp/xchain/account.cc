#include "xchain/contract.pb.h"
#include "xchain/account.h"
#include "xchain/syscall.h"

namespace pb = xchain::contract::sdk;

namespace xchain {

Account::Account(const std::string& name) {
    _name = name;
}

Account::~Account() {}

std::string Account::get_account() {
    return _name;
}

bool Account::transfer(const std::string& to, const std::string& amount) {
    pb::TransferRequest req;
    pb::TransferResponse rep;
    req.set_from(_name);
    req.set_to(to);
    req.set_amount(amount);
    bool ok = syscall("Transfer", req, &rep);
    if (!ok) {
        return false;
    }
    return true;
}

}  // namespace xchain

