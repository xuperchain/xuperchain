#include "xchain/xchain.h"

struct AntiYellow : public xchain::Contract {};

DEFINE_METHOD(AntiYellow, initialize) {
    xchain::Context* ctx = self.context();
    ctx->ok("initialize succeed");
}

DEFINE_METHOD(AntiYellow, call) {
    xchain::Context* ctx = self.context();
    ctx->ok("access permission succeed");
}

