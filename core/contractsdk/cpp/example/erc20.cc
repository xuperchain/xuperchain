#include "xchain/xchain.h"

struct ERC20 : public xchain::Contract {};

const std::string BALANCEPRE = "balanceOf_";
const std::string ALLOWANCEPRE = "allowanceOf_";
const std::string MASTERPRE = "owner";

DEFINE_METHOD(ERC20, initialize) {
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

    std::string key = BALANCEPRE + caller;
    ctx->put_object("totalSupply", totalSupply);
    ctx->put_object(key, totalSupply);

    std::string master = MASTERPRE;
    ctx->put_object(master, caller);
    ctx->ok("initialize success");
}

DEFINE_METHOD(ERC20, mint) {
    xchain::Context* ctx = self.context();
    const std::string& caller = ctx->initiator();
    if (caller.empty()) {
        ctx->error("missing caller");
        return;
    }

    std::string master;
    if (!ctx->get_object(MASTERPRE, &master)) {
        ctx->error("missing master");
        return;
    }
    if (master != caller) {
        ctx->error("only the person who created the contract can mint");
        return;
    }

    const std::string& increaseSupply = ctx->arg("amount");
    if (increaseSupply.empty()) {
        ctx->error("missing amount");
        return;
    }

    std::string value;
    if (!ctx->get_object("totalSupply", &value)) {
        ctx->error("get totalSupply error");
        return;
    }
    
    int increaseSupplyint = atoi(increaseSupply.c_str());
    int valueint = atoi(value.c_str());
    int totalSupplyint = increaseSupplyint + valueint;
    char buf[32];
    snprintf(buf, 32, "%d", totalSupplyint);
    ctx->put_object("totalSupply", buf); 
    
    std::string key = BALANCEPRE + caller;
    if (!ctx->get_object(key, &value)) {
        ctx->error("get caller balance error");
        return;
    }
    valueint = atoi(value.c_str());
    int callerint = increaseSupplyint + valueint;
    snprintf(buf, 32, "%d", callerint);
    ctx->put_object(key, buf); 
    
    ctx->ok(buf);
}

DEFINE_METHOD(ERC20, totalSupply) {
    xchain::Context* ctx = self.context();
    std::string value;
    if (ctx->get_object("totalSupply", &value)) {
        ctx->ok(value);
    } else {
        ctx->error("key not found");
    }
}

DEFINE_METHOD(ERC20, balance) {
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
        ctx->error("key not found");
    }
}

DEFINE_METHOD(ERC20, allowance) {
    xchain::Context* ctx = self.context();
    const std::string& from = ctx->arg("from");
    if (from.empty()) {
        ctx->error("missing from");
        return;
    }
   
    const std::string& to = ctx->arg("to");
    if (to.empty()) {
        ctx->error("missing to");
        return;
    }

    std::string key = ALLOWANCEPRE + from + "_" + to;
    std::string value;
    if (ctx->get_object(key, &value)) {
        ctx->ok(value);
    } else {
        ctx->error("key not found");
    }
}

DEFINE_METHOD(ERC20, transfer) {
    xchain::Context* ctx = self.context();
    const std::string& from = ctx->arg("from");
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
    int token = atoi(token_str.c_str());

    std::string from_key = BALANCEPRE + from;
    std::string value;
    int from_balance = 0;
    if (ctx->get_object(from_key, &value)) {
        from_balance = atoi(value.c_str()); 
        if (from_balance < token) {
            ctx->error("The balance of from not enough");
            return;
        }  
    } else {
        ctx->error("key not found");
        return;
    }

    std::string to_key = BALANCEPRE + to;
    int to_balance = 0;
    if (ctx->get_object(to_key, &value)) {
        to_balance = atoi(value.c_str());
    }
   
    from_balance = from_balance - token;
    to_balance = to_balance + token;
   
    char buf[32]; 
    snprintf(buf, 32, "%d", from_balance);
    ctx->put_object(from_key, buf);
    snprintf(buf, 32, "%d", to_balance);
    ctx->put_object(to_key, buf);

    ctx->ok("transfer success");
}

DEFINE_METHOD(ERC20, transferFrom) {
    xchain::Context* ctx = self.context();
    const std::string& from = ctx->arg("from");
    if (from.empty()) {
        ctx->error("missing from");
        return;
    }
  
    const std::string& caller = ctx->arg("caller");
    if (caller.empty()) {
        ctx->error("missing caller");
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
    int token = atoi(token_str.c_str());

    std::string allowance_key = ALLOWANCEPRE + from + "_" + caller;
    std::string value;
    int allowance_balance = 0;
    if (ctx->get_object(allowance_key, &value)) {
        allowance_balance = atoi(value.c_str()); 
        if (allowance_balance < token) {
            ctx->error("The allowance of from_to not enough");
            return;
        }  
    } else {
        ctx->error("You need to add allowance from_to");
        return;
    }

    std::string from_key = BALANCEPRE + from;
    int from_balance = 0;
    if (ctx->get_object(from_key, &value)) {
        from_balance = atoi(value.c_str()); 
        if (from_balance < token) {
            ctx->error("The balance of from not enough");
            return;
        }  
    } else {
        ctx->error("From no balance");
        return;
    }

    std::string to_key = BALANCEPRE + to;
    int to_balance = 0;
    if (ctx->get_object(to_key, &value)) {
        to_balance = atoi(value.c_str());
    }
   
    from_balance = from_balance - token;
    to_balance = to_balance + token;
    allowance_balance = allowance_balance - token;

    char buf[32]; 
    snprintf(buf, 32, "%d", from_balance);
    ctx->put_object(from_key, buf);
    snprintf(buf, 32, "%d", to_balance);
    ctx->put_object(to_key, buf);
    snprintf(buf, 32, "%d", allowance_balance);
    ctx->put_object(allowance_key, buf);

    ctx->ok("transferFrom success");
}

DEFINE_METHOD(ERC20, approve) {
    xchain::Context* ctx = self.context();
    const std::string& from = ctx->arg("from");
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
    int token = atoi(token_str.c_str());

    std::string from_key = BALANCEPRE + from;
    std::string value;
    if (ctx->get_object(from_key, &value)) {
        int from_balance = atoi(value.c_str()); 
        if (from_balance < token) {
            ctx->error("The balance of from not enough");
            return;
        }  
    } else {
        ctx->error("From no balance");
        return;
    }

    std::string allowance_key = ALLOWANCEPRE + from + "_" + to;
    int allowance_balance = 0;
    if (ctx->get_object(allowance_key, &value)) {
        allowance_balance = atoi(value.c_str()); 
    }

    allowance_balance = allowance_balance + token;
   
    char buf[32]; 
    snprintf(buf, 32, "%d", allowance_balance);
    ctx->put_object(allowance_key, buf);

    ctx->ok("approve success");
}



