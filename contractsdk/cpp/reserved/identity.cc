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
    const std::string& creator = ctx->arg("creator");
    if (creator.empty()) {
        ctx->error("missing creator");
        return;
    }
    ctx->put_object(creator, "true");
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

    const std::string initiator = ctx->initiator();
    ctx->get_object(initiator, &value);
    if (value != "true") {
        ctx->error("verify initiator error");
        return;
    }

    int auth_require_size = ctx->auth_require_size();
    std::vector<std::string> accounts;
    std::string sub_str = std::string(1, delimiter_account);
    for ( int iter = 0; iter < auth_require_size; ++iter) {
        std::string auth_require = ctx->auth_require(iter);
        std::string ak;
        std::string auth_value;
        std::size_t found = auth_require.rfind(sub_str);
        if (found != std::string::npos) {
            ak = auth_require.substr(found + 1, std::string::npos);
        } else {
            ak = auth_require;
        }
        ctx->get_object(ak, &auth_value);
        if (auth_value != "true") {
            ctx->error("verify auth_require error");
            return;
        }
    }
    ctx->ok("verify all aks success");
}
