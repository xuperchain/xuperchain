#include "xchain/xchain.h"
#include <string>

const std::string UserBucket = "USER";
const std::string HashBucket = "HASH";

// 文件Hash存证API规范
// 参数由Context提供
class HashDepositBasic {
public:
    // 初始化，基本不做任何工作
    virtual void initialize() = 0;
    // 将用户参数{user_id,file_name,hash_id}存储到磁盘
    virtual void storeFileInfo() = 0;
    // 查询当前合约下所有用户
    virtual void queryUserList() = 0;
    // 查询某个User下所有信息
    virtual void queryFileInfoByUser() = 0;
    // 按照Hash查询文件信息,需要指定用户
    virtual void queryFileInfoByHash() = 0;
};

struct HashDeposit : public HashDepositBasic, public xchain::Contract {
public:
    void initialize() {
        xchain::Context* ctx = this->context();
        ctx->ok("initialize success");
    }
    void storeFileInfo() {
        xchain::Context* ctx = this->context();    
        std::string user_id = ctx->arg("user_id");
        std::string hash_id = ctx->arg("hash_id");
        std::string file_name = ctx->arg("file_name");
        const std::string userKey = UserBucket + "/" + user_id + "/" + hash_id;
        const std::string hashKey = HashBucket + "/" + hash_id;
        std::string value = user_id + "\t" + hash_id + "\t" + file_name;
        std::string tempVal;
        if (ctx->get_object(hashKey, &tempVal)) {
            ctx->error("storeFileInfo failed, such hash has existed already");
            return;
        }
        if (ctx->put_object(userKey, value) && ctx->put_object(hashKey, value)) {
            ctx->ok("storeFileInfo success");
            return;
        }
        ctx->error("storeFileInfo failed");
    }
    
    void queryUserList() {
        xchain::Context* ctx = this->context();
        const std::string key = UserBucket + "/";
        std::unique_ptr<xchain::Iterator> iter = ctx->new_iterator(key, key + "～");
        std::string result;
        while (iter->next()) {
            std::pair<std::string, std::string> res;
            iter->get(&res);    
            if (res.first.length() > UserBucket.length() + 1) {
                result += res.first.substr(UserBucket.length() + 1) + "\n";
            }
        }
        ctx->ok(result);
    }
    void queryFileInfoByUser() {
        xchain::Context* ctx = this->context();
        const std::string key = UserBucket + "/" + ctx->arg("user_id");
        std::unique_ptr<xchain::Iterator> iter = ctx->new_iterator(key, key + "～");
        std::string result;
        while (iter->next()) {
            std::pair<std::string, std::string> res;
            iter->get(&res);
            result += res.second + "\n";
        }
        ctx->ok(result);
    }
    void queryFileInfoByHash() {
        xchain::Context* ctx = this->context();
        
        const std::string key = HashBucket + "/" + ctx->arg("hash_id");
        std::string value;
        bool ret = ctx->get_object(key, &value);
        if (ret) {
            ctx->ok(value);
            return;
        }
        ctx->error("queryFileInfoByHash error");
    }
};

DEFINE_METHOD(HashDeposit, initialize) {
    self.initialize();
}

DEFINE_METHOD(HashDeposit, storeFileInfo) {
    self.storeFileInfo();
}

DEFINE_METHOD(HashDeposit, queryUserList) {
    self.queryUserList();
}

DEFINE_METHOD(HashDeposit, queryFileInfoByUser) {
    self.queryFileInfoByUser();
}

DEFINE_METHOD(HashDeposit, queryFileInfoByHash) {
    self.queryFileInfoByHash();
}
