#include "xchain/syscall.h"

extern "C" uint32_t call_method(const char* method, uint32_t method_len,
                                const char* request, uint32_t request_len);
extern "C" uint32_t fetch_response(char* response, uint32_t response_len);
extern "C" uint32_t call_method_v2(const char* method, uint32_t method_len,
                                   const char* request, uint32_t request_len,
                                   char* response, uint32_t response_len,
                                   uint32_t* success);

namespace xchain {

static bool syscall_raw(const std::string& method, const std::string& request,
                        std::string* response) {
    char buf[1024];
    uint32_t buf_len = sizeof(buf);

    uint32_t response_len = 0;
    uint32_t success = 0;

    response_len = call_method_v2(method.data(), uint32_t(method.size()),
                                  request.data(), uint32_t(request.size()),
                                  &buf[0], buf_len,
                                  &success);
    // method has no return and no error
    if (response_len <= 0) {
        return true;
    }

    // buf can hold the response
    if (response_len <= buf_len) {
      response->assign(buf, response_len);
      return success == 1;
    }

    // slow path
    response->resize(response_len + 1, 0);
    success = fetch_response(&(*response)[0u], response_len);
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
