#include "xchain/xchain.h"

struct ComplianceCheck : public xchain::Contract {};

DEFINE_METHOD(ComplianceCheck, initialize) {
    xchain::Context* ctx = self.context();
    ctx->ok("initialize succeed");
}

DEFINE_METHOD(ComplianceCheck, call) {
    xchain::Context* ctx = self.context();
    ctx->ok("access permission succeed");
}

