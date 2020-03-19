#include <assert.h>
#include "xchain/xchain.h"
#include "xchain/xrc01/xrc01.h"

// XRC是超级链协议标准族
// XRC_01是XRC协议家族中的第一个协议，支持在超级链中发行通用资产
// 该协议同时支持可分割和不可分割两种资产的发行、转账、授权、授权转账、查询余额、授权关系等行为；

// 通用资产协议XRC_01使用示例
// 参数由xchain::Contract中的context提供
class XRC01_Example {
public:
    /*
     * func: 初始化智能合约
     * @param: creator: 合约创建者
     */
    virtual void initialize() = 0;
    /*
     * func: 发行一个通用资产
     * @param: id: 所发行的通用资产id
     * @param: name: 所发行的通用资产名称
     * @param: fungible: 所发行资产是否可分割
     * @param: supply: 所发行资产的资产量，当为不可分割资产时，只能为1
     * @param: issue_account: 初始化发行账户
     * @param: profile_desc: 资产描述
     */
    virtual void issue() = 0;
    /*
     * func: 资产所有者将部分资产授权别人代为管理
     * 说明: 授权的原账户为合约的发起者
     * @param: to: 授权给的账户
     * @param: token_id: 授权的资产id
     * @param: amount: 授权金额
     */
    virtual void authorization() = 0;
    /*
     * func: 撤销之前授权的资产
     * 说明: 原授权的账户才有权撤销授权, 撤销授权发起者为合约调用者
     * @param: from: 被撤销授权的账户
     * @param: token_id: 被撤销授权的资产id
     * @param: amount: 被撤销授权的金额
     */
    virtual void withdraw_authorization() = 0;
    /*
     * func: 发起转账
     * 说明: 转出账户为合约调用者
     * @param: to: 转账接收者
     * @param: token_id: 资产id
     * @param: amount: 转账金额
     */
    virtual void transfer() = 0;
    /*
     * func: 被授权的账户进行代为转账
     * @param: from: 转账发起账户
     * @param: to: 转账接收者
     * @param: token_id: 资产id
     * @param: amount: 转账金额
     */
    virtual void authorize_transfer() = 0;
    /*
     * func: 查询余额
     * @param: account: 被查询账户
     * @param: token_id: 资产id
     */
    virtual void get_balance() = 0;
    /*
     * func: 查询账户被授权的金额
     * @param: account: 被查询账户 
     * @param: token_id: 资产id
     */
    virtual void get_authorized() = 0;
    /*
     * func: 查询不可分割资产所属账户
     * @param: token_id: 资产id
     */
    virtual void get_owner_of() = 0;
    /*
     * func: 查询账户某资产的授权详情
     * @param: account: 被查询账户 
     * @param: token_id: 资产id
    */
    virtual void get_authorize_infos() = 0;
    /*
     * func: 查询账户某资产被授权详情
     * @param: account: 被查询账户 
     * @param: token_id: 资产id
    */
    virtual void get_authorized_infos() = 0;
};

struct XRC01_Demo : public XRC01_Example, public xchain::Contract {
public:
    bool str2bool(const std::string var, bool& var_bool) {
        if (var == "1") {
            var_bool = true;
            return true;
        } else if (var == "0") {
            var_bool = false;
            return true;
        }
        return false;
    }

    bool safe_stoull(const std::string in, uint64_t* out) {
        if (in.empty()) {
            return false;
        }
        for (int i = 0; i < in.size(); i++) {
            if (in[i] < '0' || in[i] > '9') {
                return false;
            }    
        }
        std::string::size_type sz = 0;
        (*out) = std::stoull(in, &sz);
        if (sz != in.size()) {
            return false;
        }
        return true;
    }

    void initialize() {
        xchain::Context* ctx = this->context();
        const std::string& creator = ctx->arg("creator");
        if (creator.empty()) {
            ctx->error("missing creator");
            return;
        }
        ctx->put_object("creator", creator);
        ctx->ok("initialize succeed");
    }

    void issue() {
        xchain::Context* ctx = this->context();
        XRC01 xrc01(ctx);
        XRC01::token token;

        const std::string& fungible_str = ctx->arg("fungible");
        bool fungible;
        if (!str2bool(fungible_str, fungible)) {
            printf("issue token failed, fungible params illegal! \n");
            ctx->error("issue token failed fungible params illegal!");
            return;
        }

        uint64_t id;
        if (!safe_stoull(ctx->arg("id"), &id)) {
            ctx->error("issue error, param id error");
            return;
        }
        const std::string& name = ctx->arg("name");
        uint64_t supply;
        if (!safe_stoull(ctx->arg("supply"), &supply)) {
            ctx->error("issue error, param supply error");
            return;
        }
        const std::string& issue_account = ctx->arg("issue_account");
        const std::string& profile_desc = ctx->arg("profile_desc");
        if (name.empty() || issue_account.empty() || profile_desc.empty()) {
            ctx->error("issue error, param error");
            return;
        }

        token.set_id(id);
        token.set_name(name);
        token.set_fungible(fungible);
        token.set_supply(supply);
        token.set_issue_account(issue_account);
        token.set_profile_desc(profile_desc);

        if (!xrc01.issue(&token)) {
            printf("issue token failed! \n");
            ctx->error("issue token failed!");
            return;
        }
        ctx->ok("issue succeed");
    }

    void authorization() {
        xchain::Context* ctx = this->context();
        XRC01 xrc01(ctx);
        const std::string& to = ctx->arg("to");
        uint64_t token_id;
        if (!safe_stoull(ctx->arg("token_id"), &token_id)) {
            ctx->error("authorization error, param token_id error");
            return;
        }

        uint64_t amount;
        if (!safe_stoull(ctx->arg("amount"), &amount)) {
            ctx->error("authorization error, param amount error");
            return;
        }
        if (!xrc01.authorization(to, token_id, amount)) {
            printf("authorization token failed! \n");
            ctx->error("authorization token failed!");
            return;
        }
        ctx->ok("authorization succeed");
    }

    void withdraw_authorization() {
        xchain::Context* ctx = this->context();
        XRC01 xrc01(ctx);

        const std::string& from = ctx->arg("from");
        uint64_t token_id;
        if (!safe_stoull(ctx->arg("token_id"), &token_id)) {
            ctx->error("withdraw_authorization error, param token_id error");
            return;
        }

        uint64_t amount;
        if (!safe_stoull(ctx->arg("amount"), &amount)) {
            ctx->error("withdraw_authorization error, param amount error");
            return;
        }

        if (!xrc01.withdraw_authorization(from, token_id, amount)) {
            printf("withdraw_authorization token failed! \n");
            ctx->error("withdraw_authorization token failed!");
            return;
        }
        ctx->ok("withdraw_authorization succeed");       
    }

    void transfer() {
        xchain::Context* ctx = this->context();
        XRC01 xrc01(ctx);

        const std::string& to = ctx->arg("to");
        uint64_t token_id;
        if (!safe_stoull(ctx->arg("token_id"), &token_id)) {
            ctx->error("transfer error, param token_id error");
            return;
        }

        uint64_t amount;
        if (!safe_stoull(ctx->arg("amount"), &amount)) {
            ctx->error("transfer error, param amount error");
            return;
        }

        if (!xrc01.transfer(to, token_id, amount)) {
            printf("transfer token failed! \n");
            ctx->error("transfer token failed!");
            return;
        }
        ctx->ok("transfer succeed"); 
    }

    void authorize_transfer() {
        xchain::Context* ctx = this->context();
        XRC01 xrc01(ctx);

        const std::string& from = ctx->arg("from");
        const std::string& to = ctx->arg("to");
        uint64_t token_id;
        if (!safe_stoull(ctx->arg("token_id"), &token_id)) {
            ctx->error("authorize_transfer error, param token_id error");
            return;
        }

        uint64_t amount;
        if (!safe_stoull(ctx->arg("amount"), &amount)) {
            ctx->error("authorize_transfer error, param amount error");
            return;
        }

        if (!xrc01.transfer_from(from, to, token_id, amount)) {
            printf("authorize_transfer token failed! \n");
            ctx->error("authorize_transfer token failed!");
            return;
        }
        ctx->ok("authorize_transfer succeed");
    }

    void get_balance() {
        xchain::Context* ctx = this->context();
        XRC01 xrc01(ctx);

        const std::string& account = ctx->arg("account");
        uint64_t token_id;
        if (!safe_stoull(ctx->arg("token_id"), &token_id)) {
            ctx->error("get_balance error, param token_id error");
            return;
        }
        uint64_t balance;

        if (!xrc01.get_balance(account, token_id, &balance)) {
            printf("get_balance failed! \n");
            ctx->error("get_balance failed!");
            return;
        }
        ctx->ok(std::to_string(balance));
    }

    void get_authorized() {
        xchain::Context* ctx = this->context();
        XRC01 xrc01(ctx);

        const std::string& account = ctx->arg("account");
        uint64_t token_id;
        if (!safe_stoull(ctx->arg("token_id"), &token_id)) {
            ctx->error("get_authorized error, param token_id error");
            return;
        }
        uint64_t authorized;

        if (!xrc01.get_authorized(account, token_id, &authorized)) {
            printf("get_authorized failed! \n");
            ctx->error("get_authorized failed!");
            return;
        }
        ctx->ok(std::to_string(authorized));
    }

    void get_owner_of() {
        xchain::Context* ctx = this->context();
        XRC01 xrc01(ctx);

        uint64_t token_id;
        if (!safe_stoull(ctx->arg("token_id"), &token_id)) {
            ctx->error("get_owner_of error, param token_id error");
            return;
        }

        std::string owner;

        if (!xrc01.owner_of(token_id, &owner)) {
            printf("get_owner_of failed! \n");
            ctx->error("get_owner_of failed!");
            return;
        }
        ctx->ok(owner);
    }

    void get_authorize_infos() {
        xchain::Context* ctx = this->context();
        XRC01 xrc01(ctx);

        const std::string& account = ctx->arg("account");
        uint64_t token_id;
        if (!safe_stoull(ctx->arg("token_id"), &token_id)) {
            ctx->error("get_authorize_infos error, param token_id error");
            return;
        }

        std::vector<XRC01::authorize_info> authorize_infos;

        if (!xrc01.authorize_infos(account, token_id, authorize_infos)) {
            printf("get_authorize_infos failed!");
            ctx->error("get_authorize_infos failed!");
            return;
        }
        for (auto iter = authorize_infos.begin(); iter != authorize_infos.end();
            ++iter) {
            printf(
                "get_authorize_infos result from=%s, to=%s, token_id=%llu, "
                "anount=%llu \n",
                iter->from().c_str(), iter->to().c_str(), iter->token_id(),
                iter->amount());
        }
        ctx->ok("get_authorize_infos success!");
    }

    void get_authorized_infos() {
        xchain::Context* ctx = this->context();
        XRC01 xrc01(ctx);

        const std::string& account = ctx->arg("account");
        uint64_t token_id;
        if (!safe_stoull(ctx->arg("token_id"), &token_id)) {
            ctx->error("get_authorized_infos error, param token_id error");
            return;
        }
        std::vector<XRC01::authorize_info> authorized_infos;
        if (!xrc01.authorized_infos(account, token_id, authorized_infos)) {
            printf("get_authorized_infos failed!");
            ctx->error("get_authorized_infos failed!");
            return;
        }
        for (auto iter = authorized_infos.begin(); iter != authorized_infos.end();
            ++iter) {
            printf(
                "get_authorized_infos result from=%s, to=%s, token_id=%llu, "
                "anount=%llu \n",
                iter->from().c_str(), iter->to().c_str(), iter->token_id(),
                iter->amount());
        }
        ctx->ok("get_authorized_infos success!");     
    }
};

DEFINE_METHOD(XRC01_Demo, initialize) { self.initialize(); }

DEFINE_METHOD(XRC01_Demo, issue) { self.issue(); }

DEFINE_METHOD(XRC01_Demo, authorization) { self.authorization(); }

DEFINE_METHOD(XRC01_Demo, withdraw_authorization) { self.withdraw_authorization(); }

DEFINE_METHOD(XRC01_Demo, transfer) { self.transfer(); }

DEFINE_METHOD(XRC01_Demo, authorize_transfer) { self.authorize_transfer(); }

DEFINE_METHOD(XRC01_Demo, get_balance) { self.get_balance(); }

DEFINE_METHOD(XRC01_Demo, get_authorized) { self.get_authorized(); }

DEFINE_METHOD(XRC01_Demo, get_owner_of) { self.get_owner_of(); }

DEFINE_METHOD(XRC01_Demo, get_authorize_infos) { self.get_authorize_infos(); }

DEFINE_METHOD(XRC01_Demo, get_authorized_infos) { self.get_authorized_infos(); }
