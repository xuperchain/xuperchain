#include "xchain/json/json.h"
#include "xchain/xchain.h"

#define CHECK_ARG(argKey)                             \
    std::string argKey = ctx->arg(#argKey);           \
    if (argKey == "") {                               \
        ctx->error("missing required arg: " #argKey); \
        return;                                       \
    }

// XPoA 验证集合变更智能合约
struct Hello : public xchain::Contract {};
std::string Validate(std::string address) { return "V_" + address; }
std::string ChangeFlag() { return "CF_"; }

const char delimiter_initialize = ';';
void split_str(const std::string& str, std::vector<std::string>& str_sets,
               const std::string& sub_str) {
    std::string::size_type pos1, pos2;
    pos2 = str.find(sub_str);
    pos1 = 0;
    while (std::string::npos != pos2) {
        str_sets.push_back(str.substr(pos1, pos2 - pos1));
        pos1 = pos2 + sub_str.size();
        pos2 = str.find(sub_str, pos1);
    }
    if (pos1 != str.length()) {
        str_sets.push_back(str.substr(pos1));
    }
}

/*
 * func: 初始化函数，部署合约时默认被调用
 */
DEFINE_METHOD(Hello, initialize) {
    xchain::Context* ctx = self.context();
    CHECK_ARG(addresss);
    CHECK_ARG(neturls);
    std::vector<std::string> address_sets;
    std::vector<std::string> neturl_sets;
    std::string sub_str = std::string(1, delimiter_initialize);
    split_str(addresss, address_sets, sub_str);
    split_str(neturls, neturl_sets, sub_str);
    if ((address_sets.size() != neturl_sets.size()) ||
        address_sets.size() == 0) {
        ctx->error("initialize param error");
        return;
    }
    for (int i = 0; i < address_sets.size(); i++) {
        xchain::json j;
        j["address"] = address_sets[i];
        j["neturl"] = neturl_sets[i];

        auto data = j.dump();
        std::string old_data;
        if (ctx->get_object(Validate(address_sets[i]), &old_data)) {
            ctx->error("initialize this validate already exists");
            return;
        };
        if (!ctx->put_object(Validate(address_sets[i]), data)) {
            ctx->error("initialize fail to save validate");
            return;
        }
    }
    if (!ctx->put_object(ChangeFlag(), "initialize")) {
        ctx->error("initialize fail to save validate change flag");
        return;
    }
    ctx->ok("initialize succeed");
}

/*
 * func: XPoA添加一个新的验证节点
 * 说明:
 * 通过合约方法权限控制谁可以增加XPoA共识的验证集合，此方法不应该是高频操作
 * @param: address: 节点地址
 * @param: neturl: 节点网络连接地址
 */
DEFINE_METHOD(Hello, add_validate) {
    xchain::Context* ctx = self.context();
    CHECK_ARG(address);
    CHECK_ARG(neturl);
    xchain::json j;
    j["address"] = address;
    j["neturl"] = neturl;
    auto data = j.dump();

    std::string old_data;
    if (ctx->get_object(Validate(address), &old_data)) {
        ctx->error("this validate already exists");
        return;
    };
    if (!ctx->put_object(Validate(address), data)) {
        ctx->error("fail to save validate");
        return;
    }
    if (!ctx->put_object(ChangeFlag(), "add")) {
        ctx->error("add validate fail to save validate change flag");
        return;
    }
    ctx->ok(data);
}

/*
 * func: XPoA删除一个验证节点
 * 说明:
 * 通过合约方法权限控制谁可以减少XPoA共识的验证集合，此方法不应该是高频操作
 * @param: address: 节点地址
 */
DEFINE_METHOD(Hello, del_validate) {
    xchain::Context* ctx = self.context();
    CHECK_ARG(address);
    std::string old_data;
    if (!ctx->get_object(Validate(address), &old_data)) {
        ctx->error("this validate does not exists");
        return;
    }

    if (!ctx->delete_object(Validate(address))) {
        ctx->error("fail to delete validate");
        return;
    }
    if (!ctx->put_object(ChangeFlag(), "del")) {
        ctx->error("del validate fail to save validate change flag");
        return;
    }
    ctx->ok("ok");
}

/*
 * func: XPoA更新一个验证节点信息
 * 说明:
 * 通过合约方法权限控制谁可以减少XPoA共识的验证集合，此方法不应该是高频操作
 * @param: address: 节点地址
 * @param: neturl: 节点网络连接地址
 */
DEFINE_METHOD(Hello, update_validate) {
    xchain::Context* ctx = self.context();
    CHECK_ARG(address);
    CHECK_ARG(neturl);
    std::string old_data;
    if (!ctx->get_object(Validate(address), &old_data)) {
        ctx->error("this validate does not exists");
        return;
    }
    xchain::json j;
    j["address"] = address;
    j["neturl"] = neturl;
    auto data = j.dump();

    if (!ctx->put_object(Validate(address), data)) {
        ctx->error("fail to update validate");
        return;
    }
    if (!ctx->put_object(ChangeFlag(), "update")) {
        ctx->error("update validate fail to save validate change flag");
        return;
    }
    ctx->ok(data);
}

/*
 * func: XPoA查询所有验证节点信息
 * 说明:
 * 查询当前XPoA共识所有验证的验证集合信息
 */
DEFINE_METHOD(Hello, get_validates) {
    xchain::Context* ctx = self.context();
    std::string flag;
    if (!ctx->get_object(ChangeFlag(), &flag)) {
        ctx->error("get validate change flag error");
        return;
    }
    xchain::json j;
    std::unique_ptr<xchain::Iterator> iter =
        ctx->new_iterator(Validate(""), Validate("~"));
    while (iter->next()) {
        std::pair<std::string, std::string> kv;
        iter->get(&kv);
        auto one = xchain::json::parse(kv.second);
        j["proposers"].push_back(one);
    }
    auto result = j.dump();
    ctx->ok(result);
}