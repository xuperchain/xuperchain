#pragma once

#include <map>
#include <pthread.h>
#include "xchain/contract.pb.h"
#include "xchain/xchain.h"

namespace pb = xchain::contract::sdk;

namespace xchain { namespace cdt {

extern "C" {
    uint32_t call_method(const char* method, uint32_t method_len,
                            const char* request, uint32_t request_len);
    uint32_t fetch_response(char* response, uint32_t response_len);
}

const std::string kFileName = "/tmp/rwset.txt";
using MapType = std::map<std::string, std::string>;
class FakeContract {
public:
    FakeContract();
    ~FakeContract();
    uint32_t call_method(const std::string& m, const std::string& r);
    uint32_t fetch_response(char* res, uint32_t len);
    void init(MapType rwset, std::string initiator, std::vector<std::string> auth,
            std::string m, MapType args);
    bool _is_error;
    pb::Response resp;

private:
    std::map<std::string, std::function<std::string(std::string)>> _syscall;
    std::string _initiator;
    std::vector<std::string> _auth_require;
    std::string _method;
    MapType _args;

    std::string _buf;
    MapType _rwset;
};

void ctx_init(MapType rwset, std::string initiator, std::vector<std::string> auth,
        std::string m, MapType args);
void ctx_unlock();
void ctx_lock();

bool ctx_assert(int status, std::string message, std::string body);
bool ctx_assert(int status);

}}
