#include "xchain/xchain.h"

struct CrossQueryDemo : public xchain::Contract {};

DEFINE_METHOD(CrossQueryDemo, initialize) {
    xchain::Context* ctx = self.context();
    ctx->ok("initialize succeed");
}

DEFINE_METHOD(CrossQueryDemo, cross_query) {
    xchain::Context* ctx = self.context();
    xchain::Response response;
    ctx->cross_query("xuper://test.xuper?module=wasm&bcname=xuper&contract_name=counter&method_name=get", {{"key", "zq"}}, &response);
    *ctx->mutable_response() = response;   
    ctx->ok("ok");
}
