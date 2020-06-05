#include "xchain/json/json.h"
#include "xchain/xchain.h"

#define CHECK_ARG(argKey)                             \
    std::string argKey = ctx->arg(#argKey);           \
    if (argKey == "") {                               \
        ctx->error("missing required arg: " #argKey); \
        return;                                       \
    }

std::string Meta(std::string name) { return "M" + name; }

std::string Endorsor(std::string name, std::string address) {
    return "E" + name + "/" + address;
}

struct Naming : public xchain::Contract {};

DEFINE_METHOD(Naming, initialize) {
    xchain::Context* ctx = self.context();
    ctx->ok("initialize succeed");
}

DEFINE_METHOD(Naming, RegisterChain) {
    xchain::Context* ctx = self.context();
    CHECK_ARG(name);
    CHECK_ARG(type);
    CHECK_ARG(min_endorsor_num);
    xchain::json j;
    j["type"] = type;
    j["min_endorsor_num"] = std::atoi(min_endorsor_num.c_str());
    j["name"] = name;
    if (j["min_endorsor_num"] <= 0) {
        ctx->error("invalid min_endorsor_num, it should be greater than 0");
        return;
    }
    auto data = j.dump();
    std::string old_data;
    if (ctx->get_object(Meta(name), &old_data)) {
        ctx->error("chain name already exists");
        return;
    }
    if (!ctx->put_object(Meta(name), data)) {
        ctx->error("fail to save chain meta");
        return;
    }
    ctx->ok(data);
}

DEFINE_METHOD(Naming, UpdateChain) {
    xchain::Context* ctx = self.context();
    CHECK_ARG(name);
    CHECK_ARG(type);
    CHECK_ARG(min_endorsor_num);
    xchain::json j;
    j["type"] = type;
    j["min_endorsor_num"] = std::atoi(min_endorsor_num.c_str());
    j["name"] = name;
    if (j["min_endorsor_num"] <= 0) {
        ctx->error("invalid min_endorsor_num, it should be greater than 0");
        return;
    }
    auto data = j.dump();
    std::string old_data;
    if (!ctx->get_object(Meta(name), &old_data)) {
        ctx->error("chain name does not exist");
        return;
    }
    if (!ctx->put_object(Meta(name), data)) {
        ctx->error("fail to save chain meta");
        return;
    }
    ctx->ok(data);
}

DEFINE_METHOD(Naming, Resolve) {
    xchain::Context* ctx = self.context();
    CHECK_ARG(name);
    std::string chain_meta;
    xchain::json j;
    if (!ctx->get_object(Meta(name), &chain_meta)) {
        ctx->error("chain name does not exist");
        return;
    }
    j["chain_meta"] = xchain::json::parse(chain_meta);
    std::unique_ptr<xchain::Iterator> iter =
        ctx->new_iterator(Endorsor(name, ""), Endorsor(name, "~"));
    while (iter->next()) {
        std::pair<std::string, std::string> kv;
        iter->get(&kv);
        auto one = xchain::json::parse(kv.second);
        j["endorsors"].push_back(one);
    }
    auto result = j.dump();
    ctx->ok(result);
}

DEFINE_METHOD(Naming, AddEndorsor) {
    xchain::Context* ctx = self.context();
    CHECK_ARG(name);
    CHECK_ARG(address);
    CHECK_ARG(pub_key);
    CHECK_ARG(host);
    std::string meta;
    if (!ctx->get_object(Meta(name), &meta)) {
        ctx->error("chain name does not exist");
        return;
    }
    std::string _;
    if (ctx->get_object(Endorsor(name, address), &_)) {
        ctx->error("endorsor already exists");
        return;
    }
    xchain::json j;
    j["address"] = address;
    j["pub_key"] = pub_key;
    j["host"] = host;
    auto info = j.dump();
    if (!ctx->put_object(Endorsor(name, address), info)) {
        ctx->error("fail to save endorsor");
        return;
    }
    ctx->ok(info);
}

DEFINE_METHOD(Naming, UpdateEndorsor) {
    xchain::Context* ctx = self.context();
    CHECK_ARG(name);
    CHECK_ARG(address);
    CHECK_ARG(host);
    std::string old_info;
    if (!ctx->get_object(Endorsor(name, address), &old_info)) {
        ctx->error("endorsor does not exist");
        return;
    }
    auto j = xchain::json::parse(old_info);
    j["host"] = host;
    auto info = j.dump();
    if (!ctx->put_object(Endorsor(name, address), info)) {
        ctx->error("fail to save endorsor");
        return;
    }
    ctx->ok(info);
}

DEFINE_METHOD(Naming, DeleteEndorsor) {
    xchain::Context* ctx = self.context();
    CHECK_ARG(name);
    CHECK_ARG(address);
    std::string old_info;
    if (!ctx->get_object(Endorsor(name, address), &old_info)) {
        ctx->error("endorsor does not exist");
        return;
    }
    if (!ctx->delete_object(Endorsor(name, address))) {
        ctx->error("fail to delete endorsor");
        return;
    }
    ctx->ok("ok");
}
