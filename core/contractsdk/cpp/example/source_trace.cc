#include "xchain/xchain.h"

// 商品溯源合约模板
// 参数由xchain::Contract中的context提供
class SourceTrace {
public:
    /*
     * func: 初始化商品溯源合约
     * @param: admin: 哪个address具有商品管理员权限
     */
    virtual void initialize() = 0;
    /*
     * func: 创建一个新的商品
     * 说明: 仅有合约发起者为admin时才允许创建商品
     * @param: id: 商品id
     * @param: desc: 商品描述
     */
    virtual void createGoods() = 0;
    /*
     * func: 变更商品信息
     * 说明: 仅有合约发起者为admin时才允许变更商品
     * @param: id: 商品id
     * @param: reason: 变更原因
     */
    virtual void updateGoods() = 0;
    /*
     * func: 查询商品变更记录
     * @param: id: 商品id
     */
    virtual void queryRecords() = 0;
};

struct SourceTraceDemo : public SourceTrace, public xchain::Contract {
public:
    const std::string GOODS = "GOODS_";
    const std::string GOODSRECORD = "GOODSSRECORD_";
    const std::string GOODSRECORDTOP = "GOODSSRECORDTOP_";
    const std::string CREATE = "CREATE";

    void initialize() {
        xchain::Context* ctx = this->context();
        const std::string& admin = ctx->arg("admin");
        if (admin.empty()) {
            ctx->error("missing admin address");
            return;
        }

        ctx->put_object("admin", admin);
        ctx->ok("initialize success");
    }

    bool isAdmin(xchain::Context* ctx, const std::string& caller) {
        std::string admin;
        if (!ctx->get_object("admin", &admin)) {
            return false;
        }
        return (admin == caller);
    }

    void createGoods() {
        xchain::Context* ctx = this->context();
        const std::string& caller = ctx->initiator();
        if (caller.empty()) {
            ctx->error("missing initiator");
            return;
        }

        if (!isAdmin(ctx, caller)) {
            ctx->error("only the admin can create new goods");
            return;
        }

        const std::string& id = ctx->arg("id");
        if (id.empty()) {
            ctx->error("missing 'id' as goods identity");
            return;
        }

        const std::string& desc = ctx->arg("desc");
        if (desc.empty()) {
            ctx->error("missing 'desc' as goods desc");
            return;
        }

        std::string goodsKey = GOODS + id;
        std::string value;
        if (ctx->get_object(goodsKey, &value)) {
            ctx->error("the id is already exist, please check again");
            return;
        }
        ctx->put_object(goodsKey, desc);

        std::string goodsRecordsKey = GOODSRECORD + id + "_0";
        std::string goodsRecordsTopKey = GOODSRECORDTOP + id;
        ctx->put_object(goodsRecordsKey, CREATE);
        ctx->put_object(goodsRecordsTopKey, 0);
        ctx->ok(id);
    }

    void updateGoods() {
        xchain::Context* ctx = this->context();
        const std::string& caller = ctx->initiator();
        if (caller.empty()) {
            ctx->error("missing initiator");
            return;
        }

        if (!isAdmin(ctx, caller)) {
            ctx->error("only the admin can update goods");
            return;
        }

        const std::string& id = ctx->arg("id");
        if (id.empty()) {
            ctx->error("missing 'id' as goods identity");
            return;
        }

        const std::string& reason = ctx->arg("reason");
        if (reason.empty()) {
            ctx->error("missing 'reason' as update reason");
            return;
        }

        std::string goodsRecordsTopKey = GOODSRECORDTOP + id;
        std::string value;
        ctx->get_object(goodsRecordsTopKey, &value);
        if (value.length() == 0) {
            ctx->error("goods " + id + " not found");
            return;
        }
        int topRecord = 0;
        topRecord = atoi(value.c_str()) + 1;

        char topRecordVal[32];
        snprintf(topRecordVal, 32, "%d", topRecord);
        std::string goodsRecordsKey = GOODSRECORD + id + "_" + topRecordVal;

        ctx->put_object(goodsRecordsKey, reason);
        ctx->put_object(goodsRecordsTopKey, topRecordVal);
        ctx->ok(topRecordVal);
    }

    void queryRecords() {
        xchain::Context* ctx = this->context();
        const std::string& id = ctx->arg("id");
        if (id.empty()) {
            ctx->error("missing 'id' as goods identity");
            return;
        }

        std::string goodsKey = GOODS + id;
        std::string value;
        if (!ctx->get_object(goodsKey, &value)) {
            ctx->error("the id not exist, please check again");
            return;
        }

        std::string goodsRecordsKey = GOODSRECORD + id + "_";
        std::unique_ptr<xchain::Iterator> iter =
            ctx->new_iterator(goodsRecordsKey, goodsRecordsKey + "~");

        std::string result = "\n";
        while (iter->next()) {
            std::pair<std::string, std::string> res;
            iter->get(&res);
            if (res.first.length() > goodsRecordsKey.length()) {
                std::string goodsRecord =
                    res.first.substr(GOODSRECORD.length());
                std::string::size_type pos = goodsRecord.find("_");
                std::string goodsId = goodsRecord.substr(0, pos);
                std::string updateRecord =
                    goodsRecord.substr(pos + 1, goodsRecord.length());
                std::string reason = res.second;

                result += "goodsId=" + goodsId +
                          ",updateRecord=" + updateRecord +
                          ",reason=" + reason + '\n';
            }
        }
        ctx->ok(result);
    }
};

DEFINE_METHOD(SourceTraceDemo, initialize) { self.initialize(); }

DEFINE_METHOD(SourceTraceDemo, createGoods) { self.createGoods(); }

DEFINE_METHOD(SourceTraceDemo, updateGoods) { self.updateGoods(); }

DEFINE_METHOD(SourceTraceDemo, queryRecords) { self.queryRecords(); }
