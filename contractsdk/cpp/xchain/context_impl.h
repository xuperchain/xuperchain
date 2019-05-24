#ifndef XCHAIN_CONTEXT_IMPL_H
#define XCHAIN_CONTEXT_IMPL_H

#include "xchain/xchain.h"

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
    virtual bool get_object(const std::string& key, std::string* value);
    virtual bool put_object(const std::string& key, const std::string& value);
    virtual bool delete_object(const std::string& key);
    virtual void ok(const std::string& body);
    virtual void error(const std::string& body);
    virtual Response* mutable_response();

private:
    std::string _method;
    std::map<std::string, std::string> _args;
    Response _resp;
};
}  // namespace xchain

#endif
