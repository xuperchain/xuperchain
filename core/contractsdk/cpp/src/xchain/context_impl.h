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
    virtual bool query_tx(const std::string& txid, Transaction* tx);
    virtual bool query_block(const std::string& blockid, Block* block);
    virtual void ok(const std::string& body);
    virtual void error(const std::string& body);
    virtual Response* mutable_response();
    virtual std::unique_ptr<Iterator> new_iterator(const std::string& start,
                                                   const std::string& limit);
    virtual Account& sender();
    virtual const std::string& transfer_amount() const;
    virtual bool call(const std::string& module, const std::string& contract,
                      const std::string& method,
                      const std::map<std::string, std::string>& args,
                      Response* response);
    virtual bool cross_query(const std::string& uri,
                             const std::map<std::string, std::string>& args,
                             Response* response);
    virtual void logf(const char* fmt, ...);
    virtual bool emit_event(const std::string& name, const std::string& body);

private:
    pb::CallArgs _call_args;
    std::map<std::string, std::string> _args;
    Response _resp;
    Account _account;
};

}  // namespace xchain

#endif
