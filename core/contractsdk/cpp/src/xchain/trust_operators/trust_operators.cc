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

TrustOperators::TrustOperators(const std::string& addr):_address(addr){}

bool TrustOperators::store(xchain::Context* ctx, const uint32_t svn, const std::string& args) {
    TrustFunctionCallRequest req;
    req.set_method("store");
    req.set_args(args);
    req.set_svn(svn);
    req.set_address(_address);
    TrustFunctionCallResponse resp;
    if (!tfcall(req, &resp)) {
        return false;
    }
    assert(resp.has_kvs()); 
    auto kvs = resp.kvs();
    for (int i = 0; i < kvs.kv_size(); i++) {
        auto ok = ctx->put_object(kvs.kv(i).key(), kvs.kv(i).value());
	if (!ok) {
	    return false;
	}
    }
    return true;
}
