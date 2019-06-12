#include "xchain/xchain.h"
#include <vector>
#include <string>

class Banned : public xchain::Contract {};

const char delimiter = ',';

const std::vector<std::string> split(const std::string& rawStr) {
    std::vector<std::string> res;
    if (rawStr == "") {
        return res;
    }
    int i = 0;
    for (; i < rawStr.size(); ++i) {
        if (rawStr[i] == delimiter) {
            continue;
		}
        break;
    }
    if (i >= rawStr.size()) {
        return res;
    }
    std::string delimStr = std::string(1, delimiter);
    std::string str = rawStr.substr(i) + delimStr;
    size_t pos = std::string::npos;
    while ((pos=str.find(delimStr)) != std::string::npos) {
        std::string temp = str.substr(0, pos);
        if (temp != "") {
            res.push_back(temp);
        }
        str = str.substr(pos+1,str.size());
    }
    return res;
}

DEFINE_METHOD(Banned, initialize) {
    xchain::Context* ctx = self.context();
    ctx->ok("initialize success");
}

DEFINE_METHOD(Banned, ban) {
    xchain::Context* ctx = self.context();
    const std::string keys = ctx->arg("contract");
    const std::string value = "true";

    const std::vector<std::string> contracts = split(keys);

    for (auto iter = contracts.begin(); iter != contracts.end(); ++iter) {
        bool ret = ctx->put_object(*iter, value);
        if (!ret) {
            ctx->error("ban failed");
            return;
        }
    }
    ctx->ok("ban contract success");
}

DEFINE_METHOD(Banned, release) {
    xchain::Context* ctx = self.context();
    const std::string keys = ctx->arg("contract");

    const std::vector<std::string> contracts = split(keys);
    for (auto iter = contracts.begin(); iter != contracts.end(); ++iter) {
        bool ret = ctx->delete_object(*iter);
        if (!ret) {
            ctx->error("release failed");
            return;
        }
    }
    ctx->ok("release contract success");
}

DEFINE_METHOD(Banned, get) {
    xchain::Context* ctx = self.context();
    const std::string key = ctx->arg("contract");
    std::string value;
    if (ctx->get_object(key, &value)) {
        ctx->ok(value);
    } else {
        ctx->error("contract not found");
    }
}
