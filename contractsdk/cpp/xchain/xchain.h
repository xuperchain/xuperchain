#ifndef XCHAIN_XCHAIN_H
#define XCHAIN_XCHAIN_H

#include <map>
#include <string>
namespace xchain {

struct Response {
    int status;
    std::string message;
    std::string body;
};

class Context {
public:
    virtual ~Context() {}
    virtual const std::map<std::string, std::string>& args() const = 0;
    virtual const std::string& arg(const std::string& name) const = 0;
    virtual bool get_object(const std::string& key, std::string* value) = 0;
    virtual bool put_object(const std::string& key,
                            const std::string& value) = 0;
    virtual bool delete_object(const std::string& key) = 0;
    virtual void ok(const std::string& body) = 0;
    virtual void error(const std::string& body) = 0;
    virtual Response* mutable_response() = 0;
};

class Contract {
public:
    Contract();
    virtual ~Contract();
    Context* context() { return _ctx; };

private:
    Context* _ctx;
};

}  // namespace xchain

#define DEFINE_METHOD(contract_class, method_name)        \
    static void cxx_##method_name(contract_class&);       \
    extern "C" void __attribute__((used)) method_name() { \
        contract_class self;                              \
        cxx_##method_name(self);                          \
    };                                                    \
    static void cxx_##method_name(contract_class& self)

#endif
