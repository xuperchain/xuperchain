#ifndef XCHAIN_ACCOUNT_H
#define XCHAIN_ACCOUNT_H

namespace xchain {

class Account {
public:
    Account();
    Account(const std::string& name);
    virtual ~Account();
    const std::string& get_name();
    bool transfer(const std::string& to, const std::string& amount);

private:
    std::string _name;
};

}  // namespace xchain

#endif
