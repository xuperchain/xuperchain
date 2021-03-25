#include <iomanip>
#include <sstream>
#include "xchain/xchain.h"

// 慈善捐款公示模板
// 参数由xchain::Contract中的context提供
class Charity {
public:
    /*
     * func: 初始化慈善基金管理账户
     * @param: admin: 哪个address具有管理员权限
     */
    virtual void initialize() = 0;
    /*
     * func: 新增捐款
     * @param: donor:捐款人id（注意不能包含/字符）
     * @param: amount: 捐款金额
     * @param: timestamp: 捐款时间
     * @param: comments: 备注
     * @return: donateid: 捐款编号
     */
    virtual void donate() = 0;
    /*
     * func: 新增慈善花费
     * @param: to:善款受益人
     * @param: amount: 善款金额
     * @param: timestamp: 拨款时间
     * @param: comments: 备注，例如受益人接收证明(可以是收据链接)
     * @return: costid: 拨款编号
     */
    virtual void cost() = 0;
    /*
     * func: 获取善款综述
     * @return: totalDonates(总捐款金额),
     * totalCosts(总拨付善款),fundBalance(基金会善款余额)
     */
    virtual void statistics() = 0;
    /*
     * func: 查询某个用户的捐款记录
     * @param: donor:捐款人id
     */
    virtual void queryDonor() = 0;
    /*
     * func: 查询捐款记录
     * @param: startid: 起始记录id
     * @param: limit: 查询多少条(每次查询不超过100条)
     */
    virtual void queryDonates() = 0;
    /*
     * func: 查询拨款记录
     * @param: startid: 起始记录id
     * @param: limit: 查询多少条(每次查询不超过100条)
     */
    virtual void queryCosts() = 0;
};

struct CharityDemo : public Charity, public xchain::Contract {
private:
    const std::string USERDONATE = "UserDonate_";
    const std::string ALLDONATE = "AllDonate_";
    const std::string ALLCOST = "AllCost_";
    const std::string TOTALRECEIVED = "TotalDonates";
    const std::string TOTALCOSTS = "TotalCosts";
    const std::string BALANCE = "Balance";
    const std::string DONATECOUNT = "DonateCount";
    const std::string COSTCOUNT = "CostCount";
    const std::string ADMIN = "admin";
    const uint64_t MAX_LIMIT = 100;

    std::string getIDFromNum(uint64_t num) {
        std::ostringstream ss;
        ss << std::setw(20) << std::setfill('0') << num;
        return ss.str();
    }

    bool safe_stoull(const std::string in, uint64_t* out) {
        if (in.empty()) {
            return false;
        }
        for (int i = 0; i < in.size(); i++) {
            if (in[i] < '0' || in[i] > '9') {
                return false;
            }
        }
        std::string::size_type sz = 0;
        *out = std::stoull(in, &sz);
        if (sz != in.size()) {
            return false;
        }
        return true;
    }

    bool isAdmin(xchain::Context* ctx, const std::string& caller) {
        std::string admin;
        if (!ctx->get_object(ADMIN, &admin)) {
            return false;
        }
        return (admin == caller);
    }

public:
    void initialize() {
        xchain::Context* ctx = this->context();
        const std::string& admin = ctx->arg(ADMIN);
        if (admin.empty()) {
            ctx->error("missing admin address");
            return;
        }
        ctx->put_object(ADMIN, admin);
        ctx->put_object(TOTALRECEIVED, "0");
        ctx->put_object(TOTALCOSTS, "0");
        ctx->put_object(BALANCE, "0");
        ctx->put_object(DONATECOUNT, "0");
        ctx->put_object(COSTCOUNT, "0");
        ctx->ok("initialize success");
    }

    void donate() {
        xchain::Context* ctx = this->context();
        const std::string& caller = ctx->initiator();
        if (caller.empty()) {
            ctx->error("missing initiator");
            return;
        }

        if (!isAdmin(ctx, caller)) {
            ctx->error("only the admin can add donate record");
            return;
        }

        const std::string& donor = ctx->arg("donor");
        if (donor.empty()) {
            ctx->error("missing 'donor'");
            return;
        }

        const std::string& amountStr = ctx->arg("amount");
        if (amountStr.empty()) {
            ctx->error("missing 'amount'");
            return;
        }

        uint64_t amount;
        if (!safe_stoull(amountStr, &amount)) {
            ctx->error(
                "illegel 'amount', should be string of a positive number");
            return;
        }

        const std::string& timestamp = ctx->arg("timestamp");
        if (timestamp.empty()) {
            ctx->error("missing 'timestamp'");
            return;
        }

        const std::string& comments = ctx->arg("comments");
        if (comments.empty()) {
            ctx->error("missing 'comments'");
            return;
        }

        uint64_t totalRec, balance, donateCnt;
        std::string totalRecStr, balanceStr, donateCntStr;
        if (!ctx->get_object(DONATECOUNT, &donateCntStr) ||
            !ctx->get_object(TOTALRECEIVED, &totalRecStr) ||
            !ctx->get_object(BALANCE, &balanceStr)) {
            ctx->error("read history failed");
            return;
        }
        safe_stoull(totalRecStr, &totalRec);
        safe_stoull(balanceStr, &balance);
        safe_stoull(donateCntStr, &donateCnt);
        totalRec += amount;
        balance += amount;
        donateCnt++;

        totalRecStr = std::to_string(totalRec);
        balanceStr = std::to_string(balance);
        donateCntStr = std::to_string(donateCnt);

        std::string donateID = getIDFromNum(donateCnt);

        std::string userDonateKey = USERDONATE + donor + "/" + donateID;
        std::string allDonateKey = ALLDONATE + donateID;
        std::string donateDetail =
            "donor=" + donor + "," + "amount=" + amountStr + ", " +
            "timestamp=" + timestamp + ", " + "comments=" + comments;
        if (!ctx->put_object(userDonateKey, donateDetail) ||
            !ctx->put_object(allDonateKey, donateDetail) ||
            !ctx->put_object(DONATECOUNT, donateCntStr) ||
            !ctx->put_object(TOTALRECEIVED, totalRecStr) ||
            !ctx->put_object(BALANCE, balanceStr)) {
            ctx->error("failed to save donate record");
            return;
        }
        ctx->ok(donateID);
    }

    void cost() {
        xchain::Context* ctx = this->context();
        const std::string& caller = ctx->initiator();
        if (caller.empty()) {
            ctx->error("missing initiator");
            return;
        }

        if (!isAdmin(ctx, caller)) {
            ctx->error("only the admin can add cost record");
            return;
        }

        const std::string& to = ctx->arg("to");
        if (to.empty()) {
            ctx->error("missing 'to'");
            return;
        }

        const std::string& amountStr = ctx->arg("amount");
        if (amountStr.empty()) {
            ctx->error("missing 'amount'");
            return;
        }

        uint64_t amount;
        if (!safe_stoull(amountStr, &amount)) {
            ctx->error(
                "illegel 'amount', should be string of a positive number");
            return;
        }

        const std::string& timestamp = ctx->arg("timestamp");
        if (timestamp.empty()) {
            ctx->error("missing 'timestamp'");
            return;
        }

        const std::string& comments = ctx->arg("comments");
        if (comments.empty()) {
            ctx->error("missing 'comments'");
            return;
        }

        uint64_t totalCost, balance, costCnt;
        std::string totalCostStr, balanceStr, costCntStr;
        if (!ctx->get_object(COSTCOUNT, &costCntStr) ||
            !ctx->get_object(TOTALCOSTS, &totalCostStr) ||
            !ctx->get_object(BALANCE, &balanceStr)) {
            ctx->error("read history failed");
            return;
        }
        safe_stoull(totalCostStr, &totalCost);
        safe_stoull(balanceStr, &balance);
        safe_stoull(costCntStr, &costCnt);
        if (balance < amount) {
            ctx->error("fund balance is not enough");
            return;
        }

        totalCost += amount;
        balance -= amount;
        costCnt++;

        totalCostStr = std::to_string(totalCost);
        balanceStr = std::to_string(balance);
        costCntStr = std::to_string(costCnt);

        std::string costID = getIDFromNum(costCnt);

        std::string allCostKey = ALLCOST + costID;
        std::string costDetail = "to=" + to + "," + "amount=" + amountStr +
                                 ", " + "timestamp=" + timestamp + ", " +
                                 "comments=" + comments;
        if (!ctx->put_object(allCostKey, costDetail) ||
            !ctx->put_object(COSTCOUNT, costCntStr) ||
            !ctx->put_object(TOTALCOSTS, totalCostStr) ||
            !ctx->put_object(BALANCE, balanceStr)) {
            ctx->error("failed to save cost record");
            return;
        }
        ctx->ok(costID);
    }

    void statistics() {
        xchain::Context* ctx = this->context();
        uint64_t totalCost, balance, totalRec;
        std::string totalCostStr, balanceStr, totalRecStr;
        if (!ctx->get_object(TOTALRECEIVED, &totalRecStr) ||
            !ctx->get_object(TOTALCOSTS, &totalCostStr) ||
            !ctx->get_object(BALANCE, &balanceStr)) {
            ctx->error("read history failed");
            return;
        }

        std::string result = "totalDonates=" + totalRecStr + ", " +
                             "totalCosts=" + totalCostStr + ", " +
                             "fundBalance=" + balanceStr + "\n";
        ctx->ok(result);
    }

    void queryDonor() {
        xchain::Context* ctx = this->context();
        // admin can get the asset data of other users
        const std::string& donor = ctx->arg("donor");
        if (donor.empty()) {
            ctx->error("missing 'donor' in request");
            return;
        }

        if (donor.find("/") != std::string::npos) {
            ctx->error("donor should not contain /");
            return;
        }

        std::string userDonateKey = USERDONATE + donor + "/";
        std::unique_ptr<xchain::Iterator> iter =
            ctx->new_iterator(userDonateKey, userDonateKey + "~");
        std::string result;
        int donateCnt = 0;
        while (iter->next()) {
            std::pair<std::string, std::string> res;
            iter->get(&res);
            if (res.first.length() > userDonateKey.length()) {
                donateCnt++;
                std::string donateID = res.first.substr(userDonateKey.length());
                std::string content = res.second;
                result += "id=" + donateID + ", " + content + "\n";
            }
        }
        result =
            "total donate count:" + std::to_string(donateCnt) + "\n" + result;
        ctx->ok(result);
    }

    void queryDonates() {
        xchain::Context* ctx = this->context();
        const std::string& startID = ctx->arg("startid");
        if (startID.empty()) {
            ctx->error("missing 'startid' in request");
            return;
        }

        const std::string& limitStr = ctx->arg("limit");
        if (limitStr.empty()) {
            ctx->error("missing 'limit' in request");
            return;
        }
        uint64_t limit;
        safe_stoull(limitStr, &limit);
        if (limit > MAX_LIMIT) {
            ctx->error("'limit' is too large");
            return;
        }

        std::string donateKey = ALLDONATE + startID;
        std::string result;
        std::unique_ptr<xchain::Iterator> iter =
            ctx->new_iterator(donateKey, ALLDONATE + "~");
        int selected = 0;
        while (iter->next() && selected < limit) {
            std::pair<std::string, std::string> res;
            iter->get(&res);
            if (res.first.length() > ALLDONATE.length()) {
                selected++;
                std::string donateID = res.first.substr(ALLDONATE.length());
                std::string content = res.second;
                result += "id=" + donateID + ", " + content + '\n';
            }
        }
        ctx->ok(result);
    }

    void queryCosts() {
        xchain::Context* ctx = this->context();
        const std::string& startID = ctx->arg("startid");
        if (startID.empty()) {
            ctx->error("missing 'startid' in request");
            return;
        }

        const std::string& limitStr = ctx->arg("limit");
        if (limitStr.empty()) {
            ctx->error("missing 'limit' in request");
            return;
        }
        uint64_t limit;
        safe_stoull(limitStr, &limit);
        if (limit > MAX_LIMIT) {
            ctx->error("'limit' is too large");
            return;
        }

        std::string costKey = ALLCOST + startID;
        std::string result;
        std::unique_ptr<xchain::Iterator> iter =
            ctx->new_iterator(costKey, ALLCOST + "~");
        int selected = 0;
        while (iter->next() && selected < limit) {
            std::pair<std::string, std::string> res;
            iter->get(&res);
            if (res.first.length() > ALLCOST.length()) {
                selected++;
                std::string costID = res.first.substr(ALLCOST.length());
                std::string content = res.second;
                result += "id=" + costID + ", " + content + '\n';
            }
        }
        ctx->ok(result);
    };
};

DEFINE_METHOD(CharityDemo, initialize) { self.initialize(); }

DEFINE_METHOD(CharityDemo, donate) { self.donate(); }

DEFINE_METHOD(CharityDemo, cost) { self.cost(); }

DEFINE_METHOD(CharityDemo, statistics) { self.statistics(); }

DEFINE_METHOD(CharityDemo, queryDonor) { self.queryDonor(); }

DEFINE_METHOD(CharityDemo, queryDonates) { self.queryDonates(); }

DEFINE_METHOD(CharityDemo, queryCosts) { self.queryCosts(); }