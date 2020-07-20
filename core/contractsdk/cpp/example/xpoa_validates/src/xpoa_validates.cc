#include "xchain/json/json.h"
#include "xchain/xchain.h"

class jvectorFinder {
public:
    jvectorFinder(const std::string address) : address(address) {}
    bool operator ()(const std::vector<xchain::json>::value_type &value) { 
        return value.is_object() and value.value("Address", "") == address; 
    }
private:
    std::string address;
};

// XPoA 验证集合变更智能合约
class XpoaValidates : public xchain::Contract {
private:
    const std::string VALIDATES_KEY = "VALIDATES";

public:
    std::string changeFlag() const { return "CF_"; }
    
    bool checkArg(xchain::Context* ctx, const std::string& key, std::string& value) {
        value = ctx->arg(key);
        if (value.empty()) {
            ctx->error("missing required arg: " + key);
            return false;
        }
        return true;
    }

    void split(std::vector<std::string>& str_sets, const std::string& str, const std::string& separator) {
        std::string::size_type pos1, pos2;
        pos2 = str.find(separator);
        pos1 = 0;
        while (std::string::npos != pos2) {
            str_sets.push_back(str.substr(pos1, pos2 - pos1));
            pos1 = pos2 + separator.size();
            pos2 = str.find(separator, pos1);
        }
        if (pos1 != str.length()) {
            str_sets.push_back(str.substr(pos1));
        }
        return;
    }

    /*
     * func: 初始化函数，部署合约时默认被调用
     */
    void initialize() {
        xchain::Context* ctx = this->context();
        // 检查合约参数是否包含所需字段
        std::string addresss, neturls;
        if (!checkArg(ctx, "addresss", addresss) || !checkArg(ctx, "neturls", neturls)) {
            return;
        }
        std::vector<std::string> address_sets;
        split(address_sets, addresss, ";");
        std::vector<std::string> neturl_sets;
        split(neturl_sets, neturls, ";");
        if (!address_sets.size() || address_sets.size() != neturl_sets.size()) {
            ctx->error("initialize xpoa param error");
            return;
        }
        std::string buffer;
        if (ctx->get_object(VALIDATES_KEY, &buffer) && !buffer.empty()) {
            ctx->error("initialize xpoa validates already exist");
            return;
        }

        xchain::json jValidatesArray = xchain::json::array();
        for (int i = 0; i < address_sets.size(); ++i) {
            xchain::json jItem = {
                { "Address", address_sets[i] },
                { "PeerAddr", neturl_sets[i] }
            };
            jValidatesArray.push_back(jItem);
        }
        xchain::json jValidatesObject;
        jValidatesObject["proposers"] = jValidatesArray;
        auto validatesStr = jValidatesObject.dump();
        if (validatesStr.empty() || !ctx->put_object(VALIDATES_KEY, validatesStr)) {
            ctx->error("initialize fail to save validate");
            return;
        }
        if (!ctx->put_object(changeFlag(), "initialize")) {
            ctx->error("initialize fail to save validate change flag");
            return;
        }
        ctx->ok("initialize succeed:" + validatesStr);
    }

    /*
    * func: XPoA添加一个新的验证节点
    * 说明:
    * 通过合约方法权限控制谁可以增加XPoA共识的验证集合，此方法不应该是高频操作
    * @param: address: 节点地址
    * @param: neturl: 节点网络连接地址
    */
    void add_validate() {
        xchain::Context* ctx = this->context();
        // 检查合约参数是否包含所需字段
        std::string address, neturl;
        if (!checkArg(ctx, "address", address) || !checkArg(ctx, "neturl", neturl)) {
            return;
        }
        std::string buffer;
        if (!ctx->get_object(VALIDATES_KEY, &buffer) || buffer.empty()) {
            ctx->error("Invalid origin validates.");
            return;
        }
        auto jValidatesObject = xchain::json::parse(buffer);
        if (jValidatesObject.find("proposers") == jValidatesObject.end() || !jValidatesObject["proposers"].is_array()) {
            ctx->error("Invalid origin proposers.");
            return;
        }
        if (std::find_if(jValidatesObject["proposers"].begin(), jValidatesObject["proposers"].end(), jvectorFinder(address)) != jValidatesObject["proposers"].end()) {
            ctx->error("Validate already exists");
            return;
        }

        xchain::json jItem = {
            { "Address", address },
            { "PeerAddr", neturl}
        };
        jValidatesObject["proposers"].push_back(jItem);
        std::string value = jValidatesObject.dump();
        if (value.empty() || !ctx->put_object(VALIDATES_KEY, value)) {
           ctx->error("Add new validate Failed.");
           return;
        }
        if (!ctx->put_object(changeFlag(), "add")) {
            ctx->error("Add validate fail to save validate change flag");
            return;
        }
        ctx->ok(value);
    }

    /*
    * func: XPoA删除一个验证节点
    * 说明:
    * 通过合约方法权限控制谁可以减少XPoA共识的验证集合，此方法不应该是高频操作
    * @param: address: 节点地址
    */
    void del_validate() {
        xchain::Context* ctx = this->context();
        // 检查合约参数是否包含所需字段
        std::string address;
        if (!checkArg(ctx, "address", address)) {
            return;
        }
        std::string buffer;
        if (!ctx->get_object(VALIDATES_KEY, &buffer) || buffer.empty()) {
            ctx->error("Invalid origin validates.");
            return;
        }
        auto jValidatesObject = xchain::json::parse(buffer);
        if (jValidatesObject.find("proposers") == jValidatesObject.end() || !jValidatesObject["proposers"].is_array()) {
            ctx->error("Invalid origin proposers.");
            return;
        }
        auto proposerIter = std::find_if(jValidatesObject["proposers"].begin(), jValidatesObject["proposers"].end(), jvectorFinder(address));
        if (proposerIter == jValidatesObject["proposers"].end()) {
            ctx->error("Validate doesn't exist");
            return;
        }

        jValidatesObject["proposers"].erase(proposerIter);
        std::string value = jValidatesObject.dump();
        if (value.empty() || !ctx->put_object(VALIDATES_KEY, value)) {
           ctx->error("Delete validate Failed.");
           return;
        }
        if (!ctx->put_object(changeFlag(), "del")) {
            ctx->error("Del validate fail to save validate change flag");
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
    void update_validate() {
        xchain::Context* ctx = this->context();
        std::string address, neturl;
        if (!checkArg(ctx, "address", address) || !checkArg(ctx, "neturl", neturl)) {
            return;
        }
        std::string buffer;
        if (!ctx->get_object(VALIDATES_KEY, &buffer) || buffer.empty()) {
            ctx->error("Invalid origin validates.");
            return;
        }
        auto jValidatesObject = xchain::json::parse(buffer);
        if (jValidatesObject.find("proposers") == jValidatesObject.end() || !jValidatesObject["proposers"].is_array()) {
            ctx->error("Invalid origin proposers.");
            return;
        }
        auto proposerIter = std::find_if(jValidatesObject["proposers"].begin(), jValidatesObject["proposers"].end(), jvectorFinder(address));
        if (proposerIter == jValidatesObject["proposers"].end()) {
            ctx->error("Validate doesn't exist");
            return;
        }

        (*proposerIter)["PeerAddr"] = neturl;
        std::string value = jValidatesObject.dump();
        if (value.empty() || !ctx->put_object(VALIDATES_KEY, value)) {
           ctx->error("Update validate Failed.");
           return;
        }
        if (!ctx->put_object(changeFlag(), "update")) {
            ctx->error("update validate fail to save validate change flag");
            return;
        }
        ctx->ok(value);
    }

    /*
    * func: XPoA查询所有验证节点信息
    * 说明:
    * 查询当前XPoA共识所有验证的验证集合信息
    */
    void get_validates() {
        xchain::Context* ctx = this->context();
        std::string flag;
        if (!ctx->get_object(changeFlag(), &flag)) {
            ctx->error("get validate change flag error");
            return;
        }
        std::string buffer;
        if (!ctx->get_object(VALIDATES_KEY, &buffer) || buffer.empty()) {
            ctx->error("Invalid origin validates.");
            return;
        }
        auto jValidatesObject = xchain::json::parse(buffer);

        if (jValidatesObject.find("proposers") == jValidatesObject.end() || !jValidatesObject["proposers"].is_array() ||
            jValidatesObject["proposers"].empty() || !jValidatesObject["proposers"].size()) {
            ctx->error("Invalid origin proposers.");
            return;
        }
        xchain::json resultObject;
        resultObject["proposers"] = jValidatesObject["proposers"];
        ctx->ok(resultObject.dump());
    }
};

DEFINE_METHOD(XpoaValidates, initialize) { self.initialize(); }

DEFINE_METHOD(XpoaValidates, add_validate) { self.add_validate(); }

DEFINE_METHOD(XpoaValidates, del_validate) { self.del_validate(); }

DEFINE_METHOD(XpoaValidates, update_validate) { self.update_validate(); }

DEFINE_METHOD(XpoaValidates, get_validates) { self.get_validates(); }