#ifndef XCHAIN_ACCOUNT_H
#define XCHAIN_ACCOUNT_H

namespace xchain {

class Account {
public:
    Account();
    virtual ~Account();
    void init(const std::string& name);
    std::string get_account();
    bool transfer(const std::string& to, const std::string& amount);

private:
    std::string _sender;
};

}  // namespace xchain

#endif
