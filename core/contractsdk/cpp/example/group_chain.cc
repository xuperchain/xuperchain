#include "xchain/xchain.h"

const std::string nodeBucket = "NODE";
const std::string chainBucket = "CHAIN";

const std::string endingSeparator = "\x01";

class GroupChainBasic {
public:
    virtual void initialize() = 0;
    virtual void listNode() = 0;
    virtual void addNode() = 0;
    virtual void delNode() = 0;
    virtual void getNode() = 0;
    virtual void changeIP() = 0;
    virtual void getChain() = 0;
    virtual void addChain() = 0;
    virtual void delChain() = 0;
    virtual void listChain() = 0;
};

class GroupChain : public GroupChainBasic, public xchain::Contract {
public:
    void initialize() {
        xchain::Context* ctx = this->context();
        ctx->ok("initialize success");
    }
    void listNode() {
        xchain::Context* ctx = this->context();
        std::string key = nodeBucket + ctx->arg("bcname") + endingSeparator;
        std::unique_ptr<xchain::Iterator> iter = ctx->new_iterator(key, key + "～");
        std::string result;
        while (iter->next()) {
            std::pair<std::string, std::string> res;
            iter->get(&res);
            std::string target = res.first;
            int offset = key.length();
            int length = target.length() - offset;
            result += target.substr(offset, length) + endingSeparator;
        }   
        ctx->ok(result);
    }
    void addNode() {
        xchain::Context* ctx = this->context();
        const std::string bcname  = ctx->arg("bcname");
        if (bcname == "xuper") {
            ctx->error("xuper is forbidden");
            return;
        }
        const std::string ip = ctx->arg("ip");
        std::string key = nodeBucket + bcname + endingSeparator + ip;
        std::string value = ctx->initiator();
        if (ctx->put_object(key, value)) {
            ctx->ok("add node succeed");
            return;
        }
        ctx->error("add node failed");
    }
    void delNode() {
        xchain::Context* ctx = this->context();
        std::string key = nodeBucket + ctx->arg("bcname") + endingSeparator + ctx->arg("ip");
        if (ctx->delete_object(key)) {
            ctx->ok("delete node succeed");
            return;
        }
        ctx->error("delete node failed");
    }
    void getNode() {
        xchain::Context* ctx = this->context();
        std::string key = nodeBucket + ctx->arg("bcname") + endingSeparator + ctx->arg("ip");
        std::string value;
        if (ctx->get_object(key, &value)) {
            ctx->ok(value);
            return;
        }
        ctx->error("ip not exist in white list");
    }
    void changeIP() {
        xchain::Context* ctx = this->context();
        std::string old_key = nodeBucket + ctx->arg("bcname") + endingSeparator + ctx->arg("old_ip");
        std::string new_key = nodeBucket + ctx->arg("bcname") + endingSeparator + ctx->arg("new_ip");
        std::string value;
        if (ctx->get_object(old_key, &value) && value == ctx->initiator() &&
            ctx->delete_object(old_key) && ctx->put_object(new_key, value)) {
            ctx->ok("change ip succeed");
            return;
        }
        ctx->error("change ip failed");
        
    }
    void getChain() {
        xchain::Context* ctx = this->context();
        std::string key = chainBucket + ctx->arg("bcname");
        std::string value;
        if (ctx->get_object(key, &value)) {
            ctx->ok(value);
            return;
        }
        ctx->error("get chain failed");
    }
    void addChain() {
        xchain::Context* ctx = this->context();
        std::string key = chainBucket + ctx->arg("bcname");
        if (ctx->put_object(key, "true")) {
            ctx->ok("add chain succeed");
            return;
        }
        ctx->error("add chain failed");
    }
    void delChain() {
        xchain::Context* ctx = this->context();
        std::string key = chainBucket + ctx->arg("bcname");
        if (ctx->delete_object(key)) {
            ctx->ok("delete chain succeed");
            return;
        }
        ctx->error("delete chain failed");
    }
    void listChain() {
        xchain::Context* ctx = this->context();
        std::string key = chainBucket;
        std::unique_ptr<xchain::Iterator> iter = ctx->new_iterator(key, key + "～");
        std::string result;
        while (iter->next()) {
            std::pair<std::string, std::string> res;
            iter->get(&res);
            std::string target = res.first;
            int offset = nodeBucket.length() + 1;
            int length = target.length() - offset;
            result += target.substr(offset, length);
        }   
        ctx->ok(result);
    }
};

DEFINE_METHOD(GroupChain, initialize) {
    self.initialize();
}

DEFINE_METHOD(GroupChain, listNode) {
    self.listNode();
}

DEFINE_METHOD(GroupChain, addNode) {
    self.addNode();
}

DEFINE_METHOD(GroupChain, delNode) {
    self.delNode();
}

DEFINE_METHOD(GroupChain, getNode) {
    self.getNode();
}

DEFINE_METHOD(GroupChain, changeIP) {
    self.changeIP();
}

DEFINE_METHOD(GroupChain, getChain) {
    self.getChain();
}

DEFINE_METHOD(GroupChain, addChain) {
    self.addChain();
}

DEFINE_METHOD(GroupChain, delChain) {
    self.delChain();
}

DEFINE_METHOD(GroupChain, listChain) {
    self.listChain();
}
