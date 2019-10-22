#include <assert.h>
#include "xchain/xchain.h"
#include "xrc01.h"

bool str2bool(const std::string var, bool& var_bool ) {
     if(var == "true" || var == "TRUE") {
         var_bool = true;
         return true;
     } else if (var == "false" || var == "FALSE") {
         var_bool = false;
         return true;
     }
     return false;
}

struct XRC01_E1 : public xchain::Contract {};

DEFINE_METHOD(XRC01_E1, initialize) {
    xchain::Context* ctx = self.context();
    const std::string& creator = ctx->arg("creator");
    if (creator.empty()) {
        ctx->error("missing creator");
        return;
    }
    ctx->put_object("creator", creator);
    ctx->ok("initialize succeed");
}

DEFINE_METHOD(XRC01_E1, issue) {
    xchain::Context* ctx = self.context();
    XRC01 xrc01(ctx);
    XRC01::XRC01::token token;

    const std::string& fungible_str = ctx->arg("fungible");
    bool fungible;
    if (!str2bool(fungible_str, fungible)) {
        printf("issue token failed, fungible params illegal!");
        ctx->error("issue token failed fungible params illegal!");
        return;
    }

    uint64_t id= std::stoull(ctx->arg("id"));
    const std::string& name = ctx->arg("name");
    uint64_t supply = std::stoull(ctx->arg("supply"));
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

    if (!xrc01.issue(token)) {
        printf("issue token failed!");
        ctx->error("issue token failed!");
        return;
    }
    ctx->ok("issue succeed");   
}

DEFINE_METHOD(XRC01_E1, authorization) {
    xchain::Context* ctx = self.context();
    XRC01 xrc01(ctx);
    const std::string& to = ctx->arg("to");
    uint64_t token_id = std::stoull(ctx->arg("token_id"));
    uint64_t amount = std::stoull(ctx->arg("amount"));

    if (!xrc01.authorization(to, token_id, amount)) {
        printf("authorization token failed!");
        ctx->error("authorization token failed!");
        return;
    }    
    ctx->ok("authorization succeed");   
}

DEFINE_METHOD(XRC01_E1, withdraw_authorization) {
    xchain::Context* ctx = self.context();
    XRC01 xrc01(ctx);

    const std::string& from = ctx->arg("from");
    uint64_t token_id = std::stoull(ctx->arg("token_id"));
    uint64_t amount = std::stoull(ctx->arg("amount"));

    if (!xrc01.withdraw_authorization(from, token_id, amount)) {
        printf("withdraw_authorization token failed!");
        ctx->error("withdraw_authorization token failed!");
        return;
    }    
    ctx->ok("withdraw_authorization succeed");  
}

DEFINE_METHOD(XRC01_E1, transfer) {
    xchain::Context* ctx = self.context();
    XRC01 xrc01(ctx);

    const std::string& to = ctx->arg("to");
    uint64_t token_id = std::stoull(ctx->arg("token_id"));
    uint64_t amount = std::stoull(ctx->arg("amount"));

    if (!xrc01.transfer(to, token_id, amount)) {
        printf("transfer token failed!");
        ctx->error("transfer token failed!");
        return;
    }    
    ctx->ok("transfer succeed");  
}

DEFINE_METHOD(XRC01_E1, authorize_transfer) {
    xchain::Context* ctx = self.context();
    XRC01 xrc01(ctx);

    const std::string& from = ctx->arg("from");
    const std::string& to = ctx->arg("to");
    uint64_t token_id = std::stoull(ctx->arg("token_id"));
    uint64_t amount = std::stoull(ctx->arg("amount"));

    if (!xrc01.transfer_from(from, to, token_id, amount)) {
        printf("authorize_transfer token failed!");
        ctx->error("authorize_transfer token failed!");
        return;
    }    
    ctx->ok("authorize_transfer succeed");  
}


DEFINE_METHOD(XRC01_E1, get_balance) {
    xchain::Context* ctx = self.context();
    XRC01 xrc01(ctx);

    const std::string& account = ctx->arg("account");
    uint64_t token_id = std::stoull(ctx->arg("token_id"));
    uint64_t balance;

    if (!xrc01.get_balance( account, token_id, &balance)) {
        printf("get_balance failed!");
        ctx->error("get_balance failed!");
        return;
    }    
    ctx->ok(std::to_string(balance));  
}

DEFINE_METHOD(XRC01_E1, get_authorized) {
    xchain::Context* ctx = self.context();
    XRC01 xrc01(ctx);

    const std::string& account = ctx->arg("account");
    uint64_t token_id = std::stoull(ctx->arg("token_id"));
    uint64_t authorized;

    if (!xrc01.get_authorized( account, token_id, &authorized)) {
        printf("get_authorized failed!");
        ctx->error("get_authorized failed!");
        return;
    }    
    ctx->ok(std::to_string(authorized));  
}

DEFINE_METHOD(XRC01_E1, get_owner_of) {
    xchain::Context* ctx = self.context();
    XRC01 xrc01(ctx);

    uint64_t token_id = std::stoull(ctx->arg("token_id"));
    std::string owner;

    if (!xrc01.owner_of(token_id, &owner)) {
        printf("get_owner_of failed!");
        ctx->error("get_owner_of failed!");
        return;
    }    
    ctx->ok(owner);  
}

DEFINE_METHOD(XRC01_E1, get_authorize_infos) {
    xchain::Context* ctx = self.context();
    XRC01 xrc01(ctx);

    const std::string& account = ctx->arg("account");
    uint64_t token_id = std::stoull(ctx->arg("token_id"));
    std::vector<XRC01::authorize_info> authorize_infos;


    if (!xrc01.authorize_infos(account, token_id, authorize_infos)) {
        printf("get_authorize_infos failed!");
        ctx->error("get_authorize_infos failed!");
        return;
    }    
    for(auto iter = authorize_infos.begin(); iter != authorize_infos.end(); ++iter) {
        printf("get_authorize_infos result from=%s, to=%s, token_id=%llu, anount=%llu \n",
         iter->from().c_str(), iter->to().c_str(), iter->token_id(), iter->amount());
    }
    ctx->ok("get_authorize_infos success!");  
}

DEFINE_METHOD(XRC01_E1, get_authorized_infos) {
    xchain::Context* ctx = self.context();
    XRC01 xrc01(ctx);

    const std::string& account = ctx->arg("account");
    uint64_t token_id = std::stoull(ctx->arg("token_id"));
    std::vector<XRC01::authorize_info> authorized_infos;


    if (!xrc01.authorized_infos(account, token_id, authorized_infos)) {
        printf("get_authorized_infos failed!");
        ctx->error("get_authorized_infos failed!");
        return;
    }    
    for(auto iter = authorized_infos.begin(); iter != authorized_infos.end(); ++iter) {
        printf("get_authorized_infos result from=%s, to=%s, token_id=%llu, anount=%llu \n",
         iter->from().c_str(), iter->to().c_str(), iter->token_id(), iter->amount());
    }
    ctx->ok("get_authorized_infos success!");  
}