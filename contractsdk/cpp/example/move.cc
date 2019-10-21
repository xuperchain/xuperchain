#include <cstdio>
#include <cstdlib>
#include <climits>
#include "xchain/xchain.h"

struct Move : public xchain::Contract {};

const std::string BALANCEPRE = "balanceOf_";

enum ret_t {
    RET_SUCCESS = 0,  
    RET_ERROR_INVALID_NUM,
    RET_ERROR_OVERFLOW
};

ret_t string2num(const std::string& from, int64_t *to) {
    long long temp;
    temp = std::stoll(from.c_str(), NULL, 10);
    if (temp <= 0) {
        return RET_ERROR_INVALID_NUM;
    }

    if (temp >= LLONG_MAX) {
        return RET_ERROR_OVERFLOW;
    }

    *to = static_cast<int64_t> (temp);
    printf("The num:  %lld\n", *to);
    return RET_SUCCESS;
}

DEFINE_METHOD(Move, initialize) {
    xchain::Context* ctx = self.context();
    const std::string& caller = ctx->initiator();
    if (caller.empty()) {
        ctx->error("missing caller");
        return;
    }
    const std::string& totalSupply = ctx->arg("totalSupply");
    if (totalSupply.empty()) {
        ctx->error("missing totalSupply");
        return;
    }
    int64_t total;
    if (string2num(totalSupply, &total) != RET_SUCCESS) {
        ctx->error("totalSupply is not valid");
    }

    std::string key = BALANCEPRE + caller;
    ctx->put_object(key, totalSupply);
}

DEFINE_METHOD(Move, balance) {
    xchain::Context* ctx = self.context();
    const std::string& caller = ctx->arg("caller");
    if (caller.empty()) {
        ctx->error("missing caller");
        return;
    }
    
    std::string key = BALANCEPRE + caller;
    std::string value;
    if (ctx->get_object(key, &value)) {
        ctx->ok(value);
    } else {
        ctx->error("caller not found");
    }
}

DEFINE_METHOD(Move, transfer) {
    xchain::Context* ctx = self.context();
    const std::string& from = ctx->initiator();
    if (from.empty()) {
        ctx->error("missing from");
        return;
    }
   
    const std::string& to = ctx->arg("to");
    if (to.empty()) {
        ctx->error("missing to");
        return;
    }

    const std::string& token_str = ctx->arg("token");
    if (token_str.empty()) {
        ctx->error("missing token");
        return;
    }
    int64_t token = 0;
    if (string2num(token_str, &token) != RET_SUCCESS) {
        ctx->error("token is not valid");
    }

    std::string from_key = BALANCEPRE + from;
    std::string value;
    int64_t from_balance = 0;
    if (ctx->get_object(from_key, &value)) {
        if (string2num(value.c_str(), &from_balance) != RET_SUCCESS) { 
            ctx->error("The balance format of from is wrong");
            return;
        }
        if (from_balance < token) {
            ctx->error("The balance of from is not enough");
            return;
        }  
    } else {
        ctx->error("key not found");
        return;
    }

    std::string to_key = BALANCEPRE + to;
    int64_t to_balance = 0;
    if (ctx->get_object(to_key, &value)) {
        if (string2num(value.c_str(), &to_balance) != RET_SUCCESS) {
            ctx->error("The balance format of to is wrong");
            return;
        }
    }
  
    from_balance = from_balance - token;
    if (LLONG_MAX - to_balance > token) {
        ctx->error("If to is added the token, his amount will overflow");
    }
    to_balance = to_balance + token;
   
    char buf[64]; 
    snprintf(buf, 64, "%lld", from_balance);
    ctx->put_object(from_key, buf);
    snprintf(buf, 64, "%lld", to_balance);
    ctx->put_object(to_key, buf);

    ctx->ok("transfer success");
}

