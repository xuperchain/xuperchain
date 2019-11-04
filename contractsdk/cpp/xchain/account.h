#ifndef XCHAIN_ACCOUNT_H
#define XCHAIN_ACCOUNT_H

#include <string>
#include <vector>

namespace xchain {

enum AccountType {
    ACCOUNT = 0,
    ADDRESS = 1,
};

class Account {
public:
    Account();
    Account(const std::string& name);
    virtual ~Account();
    const std::string& get_name();
    AccountType type() const;
    bool get_addresses(std::vector<std::string>* out);
    bool transfer(const std::string& amount);

private:
    std::string _name;
};

}  // namespace xchain

#endif
