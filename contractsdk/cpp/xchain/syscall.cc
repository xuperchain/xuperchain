#include "xchain/syscall.h"
#include <vector>

extern "C" uint32_t call_method(const char* method, uint32_t method_len,
                                const char* request, uint32_t request_len);
extern "C" uint32_t fetch_response(char* response, uint32_t response_len);

namespace xchain {

static bool syscall_raw(const std::string& method, const std::string& request,
                        std::string* response) {
    uint32_t response_len;
    response_len = call_method(method.data(), uint32_t(method.size()),
                               request.data(), uint32_t(request.size()));
    if (response_len <= 0) {
        return true;
    }
    std::vector<char> buf(response_len);
    uint32_t success = fetch_response(buf.data(), response_len);
    response->resize(response_len);
    response->assign(buf.begin(), buf.end());
    return success == 1;
}

bool syscall(const std::string& method,
             const ::google::protobuf::MessageLite& request,
             ::google::protobuf::MessageLite* response) {
    std::string req;
    std::string rep;
    request.SerializeToString(&req);
    bool ok = syscall_raw(method, req, &rep);
    if (!ok) {
        return false;
    }
    response->ParseFromString(rep);
    return true;
}
}
