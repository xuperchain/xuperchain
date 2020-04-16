#include "xchain/trust_operators/tf.pb.h"
#include "xchain/trust_operators/trust_operators.h"

extern "C" uint32_t xvm_tfcall(const char* inputptr, size_t inputlen,
                         char** outputptr, size_t* outputlen);

bool tfcall(const TrustFunctionCallRequest& request, TrustFunctionCallResponse* response) {
    std::string req;
    char* out = nullptr;
    size_t out_size = 0;
    request.SerializeToString(&req);
    auto ok = xvm_tfcall(req.data(), req.size(), &out, &out_size);
    // 注意是返回1表示失败
    if (ok) {
        return false;
    }
    auto ret = response->ParseFromString(std::string(out, out_size));
    free(out);
    return ret;
}

TrustOperators::TrustOperators(xchain::Context* ctx, const uint32_t svn): _ctx(ctx), _svn(svn) {}

// map_to_string convert a <string, string>map to a string
std::string map_to_string(std::map<std::string, std::string> strMap) {
        std::map<std::string, std::string>::iterator it;
        std::string str = "{";
        for(it = strMap.begin(); it != strMap.end(); ++it) {
            if (it != strMap.begin()) {
                str = str + ",";
            }
            str = str + '"' + it->first + '"';
            str = str + ":";
            str = str + '"' + it->second + '"';
        }
        str = str + "}";
        return str;
}

/*
    ops supports encrypted data operations;
    op is one of {add, sub, mul};
    left_value is left operand(encrypted), right_value is right operand(encrypted);
    output_key is the key to put encrypted result.
*/
bool TrustOperators::ops(const std::string op, const std::string left_value, const std::string right_value, const std::string output_key) {
    TrustFunctionCallRequest req;
    req.set_method(op);
    std::map<std::string, std::string> argsMap;
    argsMap["l"] = left_value;
    argsMap["r"] = right_value;
    argsMap["o"] = output_key;
    req.set_args(map_to_string(argsMap));

    req.set_svn(_svn);
    req.set_address(_ctx->initiator());
    TrustFunctionCallResponse resp;
    if (!tfcall(req, &resp)) {
        return false;
    }
    assert(resp.has_kvs());
    // tfcall only returns one kv pair {output_key: encrypted_result}
    auto ok = _ctx->put_object(resp.kvs().kv(0).key(), resp.kvs().kv(0).value());
    if (!ok) {
         return false;
    }
    return true;
}
