#include "xchain/xchain.h"
#include <vector>
#include <string>

class Identity : public xchain::Contract {};

const char delimiter_authrequire = ',';
const char delimiter_account = '/';

void split_str(const std::string& aks, std::vector<std::string>& ak_sets, const std::string& sub_str) {
    std::string::size_type pos1, pos2;
    pos2 = aks.find(sub_str);
    pos1 = 0;
    while(std::string::npos != pos2) {
        ak_sets.push_back(aks.substr(pos1, pos2-pos1));
        pos1 = pos2 + sub_str.size();
        pos2 = aks.find(sub_str, pos1);
    }
    if(pos1 != aks.length()) {
        ak_sets.push_back(aks.substr(pos1));
    }
}

DEFINE_METHOD(Identity, initialize) {
    xchain::Context* ctx = self.context();
    ctx->ok("initialize identity contract success");
}

// register_aks method register aks to identify contract
DEFINE_METHOD(Identity, register_aks) {
    xchain::Context* ctx = self.context();

    // aks register to identity contract
    const std::string aks = ctx->arg("aks");
    std::vector<std::string> ak_sets;
    std::string sub_str = std::string(1, delimiter_authrequire);
    split_str(aks, ak_sets, sub_str);

    for(auto iter = ak_sets.begin(); iter != ak_sets.end(); ++iter) {
        if (!ctx->put_object(*iter, "true")) {
            ctx->error("register aks to identify contract error");
            return;
        }
    }

    ctx->ok("register aks to identify contract success");
}

// unregister_aks method unregister aks from identify contract
DEFINE_METHOD(Identity, unregister_aks) {
    xchain::Context* ctx = self.context();

    // aks unregister form identity contract
    const std::string aks = ctx->arg("aks");
    std::vector<std::string> ak_sets;
    std::string sub_str = std::string(1, delimiter_authrequire);
    split_str(aks, ak_sets, sub_str);
  
    for (auto iter = ak_sets.begin(); iter != ak_sets.end(); ++iter) {
        if (!ctx->delete_object(*iter)) {
            ctx->error("unregister from identify contract error");
            return;
        }
    }

    ctx->ok("unregister aks from identify contract success");
}

// verify method verify whether the aks were identified
DEFINE_METHOD(Identity, verify) {
    xchain::Context* ctx = self.context();
    std::string value;

    // FIXME zq: @icexin context need to support initiator and if initiator is an account, should check account's aks.
    const std::string initiator;
    ctx->get_object(initiator, &value);
    if (value != "true") {
        ctx->error("verify initiator error");
        return;
    }

    // FIXME zq: @icexin context need to support auth_require
    std::vector<std::string> auth_require;
    std::vector<std::string> accounts;
    std::string sub_str = std::string(1, delimiter_account);
    for (auto iter = auth_require.begin(); iter != auth_require.end(); ++iter) {
        std::size_t found = (*iter).rfind(sub_str);
        std::string ak = (*iter).substr(found + 1, std::string::npos);
        ctx->get_object(ak, &value);
        if (value != "true") {
            ctx->error("verify auth_require error");
            return;
        }
    }
    ctx->ok("verify all aks success");
}
