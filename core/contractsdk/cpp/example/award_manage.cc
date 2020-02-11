#include "xchain/xchain.h"

const std::string BALANCEPRE = "balanceOf_";
const std::string ALLOWANCEPRE = "allowanceOf_";
const std::string MASTERPRE = "owner";

// 积分管理合约的基类
// 积分管理合约需要实现基类中指定的方法
// 参数由xchain::Contract中的context提供
class AwardBasic {
public:
    /*
     * func: 初始化积分管理账户以及总发行量
     * @param: initiator:交易发起者,也是初始化积分的owner
     * @param: totalSupply:发行总量,初始化时,积分全部归initiator
     */
    virtual void initialize() = 0;
    /*
     * func: 增发积分
     * @param: initiator:交易发起者,只有交易发起者等于积分owner时，才能增发
     * @param: amount:增发容量
     */
    virtual void addAward() = 0;
    /*
     * func: 获取积分总供应量
     */
    virtual void totalSupply() = 0;
    /*
     * func: 获取caller的积分余额
     * @param: caller: 合约调用者
     */
    virtual void balance() = 0;
    /*
     * func: 查询to用户能消费from用户的积分数量
     * @param: from: 被消费积分的一方
     * @param: to: 消费积分的一方
     */
    virtual void allowance() = 0;
    /*
     * func: from账户给to账户转token数量的积分
     * @param: from:转移积分的一方
     * @param: to:收积分的一方
     * @param: token:转移积分数量
     */
    virtual void transfer() = 0;
    /*
     * func: 从授权账户from转移数量为token的积分给to账户
      * @param: from:被转积分账户
      * @param: caller:合约调用者
      * @param: to:收积分账户
      * @param: token:转移的积分数量
      */
    virtual void transferFrom() = 0;
    /*
      * func: 允许to账户从from账户转移token数量的积分
     * @param: from:
     * @param: to:
     * @param: token
     */
    virtual void approve() = 0;
};

struct Award : public AwardBasic, public xchain::Contract {
public:
    void initialize() {
        xchain::Context* ctx = this->context();
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
        if (atoi(totalSupply.c_str()) <= 0) {
            ctx->error("totalSupply is overflow");
            return;
        }

        std::string key = BALANCEPRE + caller;
        ctx->put_object("totalSupply", totalSupply);
        ctx->put_object(key, totalSupply);

        std::string master = MASTERPRE;
        ctx->put_object(master, caller);
        ctx->ok("initialize success");
    }
    void addAward() {
        xchain::Context* ctx = this->context();
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
            ctx->error("only the person who created the contract can addAward");
            return;
        }

        const std::string& increaseSupply = ctx->arg("amount");
        if (increaseSupply.empty()) {
            ctx->error("missing amount");
            return;
        }
        if (atoi(increaseSupply.c_str()) <= 0) {
            ctx->error("amount is overflow");
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
        if (totalSupplyint <= 0) {
            ctx->error("amount+totalSupply is overflow");
            return;
        }
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
    void totalSupply() {
        xchain::Context* ctx = this->context();
        std::string value;
        if (ctx->get_object("totalSupply", &value)) {
            ctx->ok(value);
        } else {
            ctx->error("key not found");
        }
    }
    void balance() {
        xchain::Context* ctx = this->context();
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
    void allowance() {
        xchain::Context* ctx = this->context();
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
    void transfer() {
        xchain::Context* ctx = this->context();
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
        if (token <= 0) {
            ctx->error("token is overflow");
            return;
        }

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
    void transferFrom() {
        xchain::Context* ctx = this->context();
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
        if (token <= 0) {
            ctx->error("token is overflow");
            return;
        }

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
    void approve() {
        xchain::Context* ctx = this->context();
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
        if (token <= 0) {
            ctx->error("token is overflow");
            return;
        }

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
};


DEFINE_METHOD(Award, initialize) {
    self.initialize();
}

DEFINE_METHOD(Award, addAward) {
    self.addAward();
}

DEFINE_METHOD(Award, totalSupply) {
    self.totalSupply();
}

DEFINE_METHOD(Award, balance) {
    self.balance();
}

DEFINE_METHOD(Award, allowance) {
    self.allowance();
}

DEFINE_METHOD(Award, transfer) {
    self.transfer();
}

DEFINE_METHOD(Award, transferFrom) {
    self.transferFrom();
}

DEFINE_METHOD(Award, approve) {
    self.approve();
}
