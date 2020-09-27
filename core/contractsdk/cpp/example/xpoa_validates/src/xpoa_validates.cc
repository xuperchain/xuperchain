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
    const std::string LAST_BLOCKID = "BLOCKID";
    const std::string LAST_HEIGHT = "HEIGHT";

public:
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

    bool parseValidates(xchain::Context* ctx, xchain::json* jValidatesObject) {
        std::string buffer;
        if (!ctx->get_object(VALIDATES_KEY, &buffer) || buffer.empty()) {
            ctx->error("Invalid origin validates.");
            return false;
        }
        *jValidatesObject = xchain::json::parse(buffer);
        auto proposersIter = jValidatesObject->find("proposers");
        if (proposersIter == jValidatesObject->end() || !(*proposersIter).is_array() || (*proposersIter).empty() || !(*proposersIter).size()) {
             ctx->error("Invalid origin proposers.");
             return false;
        }
        return true;
    }
       
    bool findItem(xchain::Context* ctx, xchain::json& jObject, const std::string& targetStr, xchain::json::iterator* iter) {
        *iter = std::find_if(jObject.begin(), jObject.end(), jvectorFinder(targetStr));
        return *iter != jObject.end();
    }

    /*
     * func: 初始化函数，部署合约时默认被调用
     */
    void initialize() {
        xchain::Context* ctx = this->context();
        // 检查合约参数是否包含所需字段
        std::string addresss, neturls, blockid;
        if (!checkArg(ctx, "addresss", addresss) || !checkArg(ctx, "neturls", neturls) || !checkArg(ctx, "blockid", blockid)) {
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

        // 此处写入blockid，通过blockid限制合约多tx并发
        xchain::Block block;
        if (!ctx->query_block(blockid, &block) || !block.in_trunk) {
            ctx->error("initialize fail to check blockid");
            return;
        }
        if (!ctx->put_object(LAST_BLOCKID, blockid) || !ctx->put_object(LAST_HEIGHT, std::to_string(block.height))) {
            ctx->error("initialize fail to save blockid");
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
        std::string address, neturl, blockid;
        if (!checkArg(ctx, "address", address) || !checkArg(ctx, "neturl", neturl) || !checkArg(ctx, "blockid", blockid)) {
            return;
        }
      // 此处检查用户blockid，通过blockid限制合约多tx并发
        std::string last_blockid, last_height;
        xchain::Block block;
        if (!ctx->get_object(LAST_BLOCKID, &last_blockid) || !ctx->get_object(LAST_HEIGHT, &last_height)|| !ctx->query_block(blockid, &block) || 
            !block.in_trunk || blockid == last_blockid || !block.next_hash.empty()) {
            ctx->error("add_validate fail to check blockid");
            return;
        }
        int64_t height = 0;
        height = std::atoll(last_height.c_str());
        if (!height || height > block.height) {
            ctx->error("add_validate fail to check blockid, please input tipID from ledger");
            return;
        }
        if (!ctx->put_object(LAST_BLOCKID, blockid) || !ctx->put_object(LAST_HEIGHT, std::to_string(block.height))) {
            ctx->error("add_validate fail to save blockid");
            return;
        }

        // 检查当前proposers是否合法
        xchain::json jValidatesObject;
        if (!parseValidates(ctx, &jValidatesObject)) {
            return;
        }
        xchain::json::iterator proposerIter;
        if (findItem(ctx, jValidatesObject["proposers"], address, &proposerIter)) {
            ctx->error("Proposer has exist");
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
        std::string address, blockid;
        if (!checkArg(ctx, "address", address) || !checkArg(ctx, "blockid", blockid)) {
            return;
        }

        // 此处检查用户blockid，通过blockid限制合约多tx并发
        std::string last_blockid, last_height;
        xchain::Block block;
        if (!ctx->get_object(LAST_BLOCKID, &last_blockid) || !ctx->get_object(LAST_HEIGHT, &last_height) || !ctx->query_block(blockid, &block) ||
            !block.in_trunk || blockid == last_blockid || !block.next_hash.empty()) {
            ctx->error("del_validate fail to check blockid");
            return;
        }
        int64_t height = 0;
        height = std::atoll(last_height.c_str());
        if (!height || height > block.height) {
            ctx->error("add_validate fail to check blockid, please input tipID from ledger");
            return;
        }
        if (!ctx->put_object(LAST_BLOCKID, blockid) || !ctx->put_object(LAST_HEIGHT, std::to_string(block.height))) {
            ctx->error("del_validate fail to save blockid");
            return;
        }

        xchain::json jValidatesObject;
        if (!parseValidates(ctx, &jValidatesObject)) {
            return;
        }
        xchain::json::iterator proposerIter;
        if (!findItem(ctx, jValidatesObject["proposers"], address, &proposerIter)) {
            ctx->error("Proposer doesn't exist");
            return;
        }
        jValidatesObject["proposers"].erase(proposerIter);
        std::string value = jValidatesObject.dump();
        if (value.empty() || !ctx->put_object(VALIDATES_KEY, value)) {
           ctx->error("Delete validate Failed.");
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
        std::string address, neturl, blockid;
        if (!checkArg(ctx, "address", address) || !checkArg(ctx, "neturl", neturl) || !checkArg(ctx, "blockid", blockid)) {
            return;
        }
        // 此处检查用户blockid，通过blockid限制合约多tx并发
        std::string last_blockid, last_height;
        xchain::Block block;
        if (!ctx->get_object(LAST_BLOCKID, &last_blockid)  || !ctx->get_object(LAST_HEIGHT, &last_height) || !ctx->query_block(blockid, &block) || 
            !block.in_trunk || blockid == last_blockid || !block.next_hash.empty()) {
            ctx->error("update_validate fail to check blockid");
            return;
        }
        int64_t height = 0;
        height = std::atoll(last_height.c_str());
        if (!height || height > block.height) {
            ctx->error("add_validate fail to check blockid, please input tipID from ledger");
            return;
        }
        if (!ctx->put_object(LAST_BLOCKID, blockid) || !ctx->put_object(LAST_HEIGHT, std::to_string(block.height))) {
            ctx->error("update_validate fail to save blockid");
            return;
        }

        xchain::json jValidatesObject;
        if (!parseValidates(ctx, &jValidatesObject)) {
            return;
        }
        xchain::json::iterator proposerIter;
        if (!findItem(ctx, jValidatesObject["proposers"], address, &proposerIter)) {
            ctx->error("Proposer doesn't exist");
            return;
        }

        (*proposerIter)["PeerAddr"] = neturl;
        std::string value = jValidatesObject.dump();
        if (value.empty() || !ctx->put_object(VALIDATES_KEY, value)) {
           ctx->error("Update validate Failed.");
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
        xchain::json jValidatesObject;
        if (!parseValidates(ctx, &jValidatesObject)) {
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
