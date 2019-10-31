#include "xchain/account.h"
#include "xchain/contract.pb.h"
#include "xchain/syscall.h"

namespace pb = xchain::contract::sdk;

namespace xchain {

Account::Account() {}

Account::Account(const std::string& name) : _name(name) {}

Account::~Account() {}

const std::string& Account::get_name() { return _name; }

bool Account::transfer(const std::string& amount) {
    pb::TransferRequest req;
    pb::TransferResponse rep;
    // req.set_from(_name);
    req.set_to(_name);
    req.set_amount(amount);
    return syscall("Transfer", req, &rep);
}

bool Account::get_addresses(std::vector<std::string>* out) {
    pb::GetAccountAddressesRequest req;
    pb::GetAccountAddressesResponse rep;
    req.set_account(_name);
    bool ret = syscall("GetAccountAddresses", req, &rep);
    if (!ret) {
        return false;
    }
    out->insert(out->end(), rep.addresses().begin(), rep.addresses().end());
    return true;
}

AccountType Account::type() const {
    if (_name.compare(0, 2, "XC") != 0) {
        return ADDRESS;
    }
    if (_name.rfind("@") == std::string::npos) {
        return ADDRESS;
    }
    return ACCOUNT;
}

}  // namespace xchain
