#include "xchain/xchain.h"

const std::string TICKETID = "Luckid_";
const std::string USERID = "Userid_";
const std::string ADMIN = "admin";
const std::string RESULT = "result";
const std::string TICKETS = "tickets";

// 抽奖小游戏模板
// 参数由xchain::Contract中的context提供
class LuckDraw {
public:
    /*
     * func: 初始化游戏
     * @param: admin: 哪个address具有管理员权限
     */
    virtual void initialize() = 0;
    /*
     * func: 获得一个抽奖券
     * @param: initiator: 玩家的address，获得一个抽奖券
     */
    virtual void getLuckid() = 0;
    /*
     * func: 开始抽奖
     * @param: seed:
     * 传入一个随机数种子，可以是预言机生成的，也可以是游戏约定的，例如某天A股收盘价
     */
    virtual void startLuckDraw() = 0;
    /*
     * func: 查询抽奖结果
     */
    virtual void getResult() = 0;
};

struct LuckDrawDemo : public LuckDraw, public xchain::Contract {
public:
    void initialize() {
        xchain::Context* ctx = this->context();
        const std::string& admin = ctx->arg(ADMIN);
        if (admin.empty()) {
            ctx->error("missing admin address");
            return;
        }

        std::string key = ADMIN;
        ctx->put_object(key, admin);
        ctx->put_object(TICKETS, "0");
        ctx->ok("initialize success");
    }

    bool isAdmin(xchain::Context* ctx, const std::string& caller) {
        std::string admin;
        if (!ctx->get_object(ADMIN, &admin)) {
            return false;
        }
        return (admin == caller);
    }

    void getLuckid() {
        xchain::Context* ctx = this->context();
        const std::string& caller = ctx->initiator();
        if (caller.empty()) {
            ctx->error("missing initiator");
            return;
        }

        // 检查是否存在抽奖结果，如果存在则不再继续发放奖券
        std::string result;
        if (ctx->get_object(RESULT, &result)) {
            ctx->error("this luck draw is finished");
            return;
        }

        // 检查用户是否已经抽过奖券, 如果抽过直接返回上次抽的奖券号
        std::string userval;
        if (ctx->get_object(USERID + caller, &userval)) {
            ctx->ok(userval);
            return;
        }

        std::string lastidStr;
        if (!ctx->get_object(TICKETS, &lastidStr)) {
            ctx->error("get tickets count failed");
            return;
        }

        // 确定当前抽奖券编号，全局自增
        int lastid = std::atoi(lastidStr.c_str());
        if (lastid < 0) {
            ctx->error("tickets count is wrong");
            return;
        }
        lastid++;
        lastidStr = std::to_string(lastid);
        if (!ctx->put_object(USERID + caller, lastidStr) ||
            !ctx->put_object(TICKETID + lastidStr, caller) ||
            !ctx->put_object(TICKETS, lastidStr)) {
            ctx->error("save ticket failed");
            return;
        }
        ctx->ok(lastidStr);
    }

    void startLuckDraw() {
        xchain::Context* ctx = this->context();
        const std::string& caller = ctx->initiator();
        if (caller.empty()) {
            ctx->error("missing initiator");
            return;
        }

        if (!isAdmin(ctx, caller)) {
            ctx->error("only the admin can add new asset type");
            return;
        }

        const std::string& seedStr = ctx->arg("seed");
        if (seedStr.empty()) {
            ctx->error("missing seed");
            return;
        }
        int seed = std::atoi(seedStr.c_str());

        // 获取总奖券数
        std::string lastidStr;
        if (!ctx->get_object(TICKETS, &lastidStr)) {
            ctx->error("get tickets count failed");
            return;
        }
        int lastid = std::atoi(lastidStr.c_str());
        if (lastid == 0) {
            ctx->error("no luck draw tickets");
            return;
        }

        // 抽奖
        srand(seed);
        int lucknum = (rand() % lastid) + 1;
        std::string luckid = std::to_string(lucknum);

        std::string luckuser;
        if (!ctx->get_object(TICKETID + luckid, &luckuser)) {
            ctx->error("get luck ticket failed");
            return;
        }

        // 记录抽奖结果
        if (!ctx->put_object(RESULT, luckuser)) {
            ctx->error("save luck draw result failed");
            return;
        }
        ctx->ok(luckuser);
    }

    void getResult() {
        xchain::Context* ctx = this->context();
        // 获取总奖券数
        std::string luckuser;
        if (!ctx->get_object(RESULT, &luckuser)) {
            ctx->error("get luck draw result failed");
            return;
        }

        ctx->ok(luckuser);
    };
};

DEFINE_METHOD(LuckDrawDemo, initialize) { self.initialize(); }

DEFINE_METHOD(LuckDrawDemo, getLuckid) { self.getLuckid(); }

DEFINE_METHOD(LuckDrawDemo, startLuckDraw) { self.startLuckDraw(); }

DEFINE_METHOD(LuckDrawDemo, getResult) { self.getResult(); }