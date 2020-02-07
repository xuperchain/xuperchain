#include "xchain/xchain.h"

const std::string UserBucket = "USER";
const int TOPIC_LENGTH_LIMIT = 36;
const int TITLE_LENGTH_LIMIT = 100;
const int CONTENT_LENGTH_LIMIT = 3000;

// 短内容存证API规范
// 参数由Context提供
class ShortContentDepositBasic {
public:
    // 初始化，基本不做任何工作
    virtual void initialize() = 0;
    // 将用户参数{user_id,title,topic}存储到磁盘
    virtual void storeShortContent() = 0;
    // 按照用户粒度查询内容
    virtual void queryByUser() = 0;
    // 按照标题粒度查询内容，用户名是必须的
    virtual void queryByTitle() = 0;
    // 按照主题粒度查询内容，用户名是必须的
    virtual void queryByTopic() = 0;
};

struct ShortContentDeposit : public ShortContentDepositBasic, public xchain::Contract {
public:
    void initialize() {
        xchain::Context* ctx = this->context();
        ctx->ok("initialize success");
    }
    void storeShortContent() {
        xchain::Context* ctx = this->context();
        std::string user_id = ctx->arg("user_id");
        std::string title = ctx->arg("title");
        std::string topic = ctx->arg("topic");
        std::string content = ctx->arg("content");
        const std::string userKey = UserBucket + "/" + user_id + "/" + topic + "/" + title;
        if (topic.length() > TOPIC_LENGTH_LIMIT || title.length() > TITLE_LENGTH_LIMIT ||
            content.length() > CONTENT_LENGTH_LIMIT) {
            ctx->error("The length of topic or title or content is more than limitation");
            return;
        }
        if (ctx->put_object(userKey, content)) {
            ctx->ok("storeShortContent success");
            return;
        }
        ctx->error("storeShortContent failed");
    }
    void queryByUser() {
        xchain::Context* ctx = this->context();
        std::string key = UserBucket + "/" + ctx->arg("user_id") + "/";
        std::unique_ptr<xchain::Iterator> iter = ctx->new_iterator(key, key + "～");
        std::string result;
        while (iter->next()) {
            std::pair<std::string, std::string> res;
            iter->get(&res);
            result += res.first + "\n" + res.second + "\n";
        }
        ctx->ok(result);
    }
    void queryByTitle() {
        xchain::Context* ctx = this->context();
        std::string user_id = ctx->arg("user_id");
        std::string topic = ctx->arg("topic");
        std::string title = ctx->arg("title");
        std::string key = UserBucket + "/" + user_id + "/" + topic + "/" + title;
        std::string value;
        bool ret = ctx->get_object(key, &value);
        if (ret) {
            ctx->ok(value);
            return;
        }
        ctx->error("queryByTitle failed");
    }
    void queryByTopic() {
        xchain::Context* ctx = this->context();
        std::string user_id = ctx->arg("user_id");
        std::string topic = ctx->arg("topic");
        std::string key = UserBucket + "/" + user_id + "/" + topic + "/";
        std::string result;
        std::unique_ptr<xchain::Iterator> iter = ctx->new_iterator(key, key + "～");
        while (iter->next()) {
            std::pair<std::string, std::string> res;
            iter->get(&res);
            result += res.first + "\n" + res.second + "\n";
        }
        ctx->ok(result);
    }
};

DEFINE_METHOD(ShortContentDeposit, initialize) {
    self.initialize();
}

DEFINE_METHOD(ShortContentDeposit, storeShortContent) {
    self.storeShortContent();
}

DEFINE_METHOD(ShortContentDeposit, queryByUser) {
    self.queryByUser();
}

DEFINE_METHOD(ShortContentDeposit, queryByTitle) {
    self.queryByTitle();
}

DEFINE_METHOD(ShortContentDeposit, queryByTopic) {
    self.queryByTopic();
}
