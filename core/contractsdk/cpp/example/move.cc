#include <cstdio>
#include <cstdlib>
#include <climits>
#include <cassert>
#include "xchain/xchain.h"

struct Move : public xchain::Contract {};

const std::string BALANCEPRE = "balanceOf_";

enum ret_t {
    RET_SUCCESS = 0,  
    RET_ERROR_INVALID_NUM
};

ret_t string2num(const std::string& from, int64_t *to) {
    long long temp;
    char* p = nullptr;
    temp = std::strtoll(from.c_str(), &p, 10);

    if (temp >= LLONG_MAX || temp <= LLONG_MIN) {
        return RET_ERROR_INVALID_NUM;
    }

    if (temp < 0) {
        printf("The num is negtive: %lld\n", temp);
        return RET_ERROR_INVALID_NUM;
    }

    if (p && *p) {
        return RET_ERROR_INVALID_NUM;
    }

    *to = (int64_t) (temp);
    assert(sizeof(*to) == sizeof(temp));
    
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
    int64_t total = 0;
    if (string2num(totalSupply, &total) != RET_SUCCESS) {
        ctx->error("totalSupply is not valid");
        return;
    }

    std::string key = BALANCEPRE + caller;
    if (!ctx->put_object(key, totalSupply)) {
        ctx->error("initialize failed");
        return;
    }
    
    ctx->ok("initialize success");
}

DEFINE_METHOD(Move, balance) {
    xchain::Context* ctx = self.context();
    const std::string& caller = ctx->arg("caller");
    std::string key;
    if (caller.empty()) {
        const std::string& myself = ctx->initiator();
        key = BALANCEPRE + myself;
        if (myself.empty()) {
            ctx->error("missing caller");
            return;
        }
    } else {
        key = BALANCEPRE + caller;
    }
    
    std::string value;
    if (!ctx->get_object(key, &value)) {
        ctx->error("caller not found");
        return;
    }
    ctx->ok(value);
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
        return;
    }

    std::string from_key = BALANCEPRE + from;
    std::string value;
    int64_t from_balance = 0;
    if (!ctx->get_object(from_key, &value)) {
        ctx->error("the token you own is 0");
        return;
    }
    if (string2num(value.c_str(), &from_balance) != RET_SUCCESS) { 
        ctx->error("The balance format of from is wrong");
        return;
    }
    if (from_balance < token) {
        ctx->error("The balance of from is not enough");
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
    if (LLONG_MAX - to_balance < token) {
        ctx->error("If to is added the token, his amount will overflow");
        return;
    }
    to_balance = to_balance + token;
   
    char buf[64]; 
    snprintf(buf, 64, "%lld", from_balance);
    if (!ctx->put_object(from_key, buf)) {
        ctx->error("update from failed");
        return;
    }
    snprintf(buf, 64, "%lld", to_balance);
    if (!ctx->put_object(to_key, buf)) {
        ctx->error("update to failed");
        return;
    }

    ctx->ok("transfer success");
}

