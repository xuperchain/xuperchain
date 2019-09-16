#ifndef XCHAIN_CONTEXT_IMPL_H
#define XCHAIN_CONTEXT_IMPL_H

#include "xchain/contract.pb.h"
#include "xchain/xchain.h"

namespace pb = xchain::contract::sdk;

namespace xchain {

class ContextImpl : public Context {
public:
    ContextImpl();
    virtual ~ContextImpl();
    bool init();
    const Response& get_response();
    virtual const std::string& method();
    virtual const std::map<std::string, std::string>& args() const;
    virtual const std::string& arg(const std::string& name) const;
    virtual const std::string& initiator() const;
    virtual int auth_require_size() const;
    virtual const std::string& auth_require(int idx) const;
    virtual bool get_object(const std::string& key, std::string* value);
    virtual bool put_object(const std::string& key, const std::string& value);
    virtual bool delete_object(const std::string& key);
    virtual bool query_tx(const std::string &txid, Transaction* tx);
    virtual bool query_block(const std::string &blockid, Block* block);
    virtual void ok(const std::string& body);
    virtual void error(const std::string& body);
    virtual Response* mutable_response();
    virtual std::unique_ptr<Iterator> new_iterator(const std::string& start, const std::string& limit);

private:
    pb::CallArgs _call_args;
    std::map<std::string, std::string> _args;
    Response _resp;
};

}  // namespace xchain

#endif
