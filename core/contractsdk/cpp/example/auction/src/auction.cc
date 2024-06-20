#include "xchain/xchain.h"
#include "auction.pb.h"

// ����ģ��
// ������xchain::Contract�е�context�ṩ
class Auction {
public:
    /*
     * func: ��ʼ��
     * @param: admin: �ĸ�address���й���ԱȨ��
     */
    virtual void initialize() = 0;
    /*
     * func: ��������
     * @param: id: ����ƷΨһ��
     * @param: floor: �׼�
     */
    virtual void create() = 0;
    /*
     * func: �һ�����
     * @param: id: ����ƷΨһ��
     * @param: bidder: ������
     * @param: amount: �һ����
     * @return: amount: �һ����
     */
    virtual void chip() = 0;
    /*
     * func: ��س���
     * @param: id: ����ƷΨһ��
     * @param: bidder: ������
     * @return: amount: ��ؽ��
     */
    virtual void redeem() = 0;
    /*
     * func: ����
     * @param: id: ����ƷΨһ��
     * @param: bid: ����
     */
    virtual void bid() = 0;
    /*
     * func: �ɽ�
     * @param: id: ����ƷΨһ��
     */
    virtual void deal() = 0;
    /*
     * func: ��ѯ������Ϣ
     * @param: id: ����ƷΨһ��
     */
    virtual void query() = 0;
};

struct AuctionDemo : public Auction, public xchain::Contract {
private:
    const std::string PrefixLot = "Lot_";
    const std::string PrefixChip = "Chip_";
    const std::string PrefixRecord = "Record_";
    const std::string Admin = "admin";

    const std::string ErrorNoPermission = "no permission";
    const std::string ErrorAuctionIsOver = "auction is over";
    const std::string ErrorChipRedeemed = "chip redeemed";
    const std::string ErrorChipNotEnough = "chip not enough";
    const std::string ErrorLowerBid = "lower than current bid";

    const std::string ErrorInitiatorMissing = "initiator missing";
    const std::string ErrorAuctioneerMissing = "auctioneer missing";
    const std::string ErrorBidderMissing = "bidder missing";

    const std::string ErrorIdMissing = "id missing";
    const std::string ErrorIdConflict = "id conflict";
    const std::string ErrorIdNotExist = "id not exist";

    const std::string ErrorFloorMissing = "floor missing";
    const std::string ErrorFloorIllegal = "floor illegal";

    const std::string ErrorAmountMissing = "amount missing";
    const std::string ErrorAmountIllegal = "amount illegal";

    const std::string ErrorGetLot = "get lot error";
    const std::string ErrorSetLot = "set lot error";
    const std::string ErrorGetChip = "get chip error";
    const std::string ErrorSetChip = "set chip error";
    const std::string ErrorGetRecord = "get record error";
    const std::string ErrorSetRecord = "set record error";

    bool safe_stoull(const std::string& in, uint64_t* out) {
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

    std::string getLotKey(const std::string& id) {
        return PrefixLot + id;
    }

    std::string getChipKey(const std::string& id, const std::string& bidder) {
        return PrefixChip + id + "_" + bidder;
    }

    std::string getRecordKey(const std::string& id, uint64_t bid) {
        return PrefixRecord + id + "_" + std::to_string(bid);
    }

    std::string getRecordPrefix(const std::string& id) {
        return PrefixRecord + id + "_";
    }

public:
    void initialize() {
        xchain::Context* ctx = this->context();
        const std::string& caller = ctx->initiator();
        ctx->logf("caller: %s", caller.c_str());
        ctx->put_object(Admin, caller);
        ctx->ok("success");
    }

    void create() {
        xchain::Context* ctx = this->context();
        const std::string& initiator = ctx->initiator();
        if (initiator.empty()) {
            ctx->error(ErrorInitiatorMissing);
            return;
        }

        const std::string& id = ctx->arg("id");
        if (id.empty()) {
            ctx->error(ErrorIdMissing);
            return;
        }
        std::string lotKey = getLotKey(id);
        std::string lotStr;
        if (ctx->get_object(lotKey, &lotStr) || !lotStr.empty()) {
            ctx->error(ErrorIdConflict);
            return;
        }

        const std::string& floorStr = ctx->arg("floor");
        if (floorStr.empty()) {
            ctx->error(ErrorFloorMissing);
            return;
        }
        uint64_t floor;
        if (!safe_stoull(floorStr, &floor)) {
            ctx->error(ErrorFloorIllegal);
            return;
        }

        std::unique_ptr<auction::Lot> lot(new auction::Lot);
        lot->set_id(id);
        lot->set_auctioneer(initiator);
        lot->set_floor(floor);
        lot->set_status(auction::PROGRESS);
        if (!lot->SerializeToString(&lotStr) || !ctx->put_object(lotKey, lotStr)) {
            ctx->error(ErrorSetLot);
            return;
        }

        ctx->ok("success");
    }

    void chip() {
        xchain::Context* ctx = this->context();

        const std::string& id = ctx->arg("id");
        if (id.empty()) {
            ctx->error(ErrorIdMissing);
            return;
        }
        std::string lotKey = getLotKey(id);
        std::string lotStr;
        std::unique_ptr<auction::Lot> lot(new auction::Lot);
        if (!ctx->get_object(lotKey, &lotStr) || lotStr.empty() || !lot->ParseFromString(lotStr)) {
            ctx->error(ErrorIdNotExist);
            return;
        }

        // ���Ľ���
        if (lot->status() == auction::ENDED) {
            ctx->ok(ErrorAuctionIsOver);
            return;
        }

        const std::string& bidder = ctx->arg("bidder");
        if (bidder.empty()) {
            ctx->error(ErrorBidderMissing);
            return;
        }

        const std::string& amountStr = ctx->arg("amount");
        if (amountStr.empty()) {
            ctx->error(ErrorAmountMissing);
            return;
        }
        uint64_t amount;
        if (!safe_stoull(amountStr, &amount)) {
            ctx->error(ErrorAmountIllegal);
            return;
        }

        std::string chipKey = getChipKey(id, bidder);
        std::string chipStr;
        std::unique_ptr<auction::Chip> chip(new auction::Chip);
        if (ctx->get_object(chipKey, &chipStr) && !chipStr.empty() && chip->ParseFromString(chipStr)) {
            // ׷�ӳ���
            chip->set_amount(chip->amount()+amount);
        } else {
            // �³���
            chip->set_bidder(bidder);
            chip->set_amount(amount);
            chip->set_status(auction::PROGRESS);
        }

        if (!chip->SerializeToString(&chipStr) || !ctx->put_object(chipKey, chipStr)) {
            ctx->error(ErrorSetChip);
            return;
        }

        ctx->logf("chip: bidder:%s amount=%lld", chip->bidder().c_str(), chip->amount());
        ctx->ok("success");
    }

    void redeem() {
        xchain::Context* ctx = this->context();

        const std::string& id = ctx->arg("id");
        if (id.empty()) {
            ctx->error(ErrorIdMissing);
            return;
        }
        std::string lotKey = getLotKey(id);
        std::string lotStr;
        std::unique_ptr<auction::Lot> lot(new auction::Lot);
        if (!ctx->get_object(lotKey, &lotStr) || lotStr.empty() || !lot->ParseFromString(lotStr)) {
            ctx->error(ErrorIdNotExist);
            return;
        }

        const std::string& bidder = ctx->arg("bidder");
        if (bidder.empty()) {
            ctx->error(ErrorBidderMissing);
            return;
        }
        std::string chipKey = getChipKey(id, bidder);
        std::string chipStr;
        std::unique_ptr<auction::Chip> chip(new auction::Chip);
        if (!ctx->get_object(chipKey, &chipStr) || chipStr.empty() || !chip->ParseFromString(chipStr)) {
            ctx->error(ErrorGetChip);
            return;
        }

        if (chip->status() == auction::ENDED) {
            ctx->ok(ErrorChipRedeemed);
            return;
        }

        uint64_t amount = chip->amount();
        if (lot->bidder() == bidder) {
            amount = chip->amount() - chip->bid();
        }

        chip->set_status(auction::ENDED);
        if (!chip->SerializeToString(&chipStr) || !ctx->put_object(chipKey, chipStr)) {
            ctx->error(ErrorSetChip);
            return;
        }

        ctx->ok(std::to_string(amount));
    }

    void bid() {
        xchain::Context* ctx = this->context();

        const std::string& id = ctx->arg("id");
        if (id.empty()) {
            ctx->error(ErrorIdMissing);
            return;
        }
        std::string lotKey = getLotKey(id);
        std::string lotStr;
        std::unique_ptr<auction::Lot> lot(new auction::Lot);
        if (!ctx->get_object(lotKey, &lotStr) || lotStr.empty() || !lot->ParseFromString(lotStr)) {
            ctx->error(ErrorGetLot);
            return;
        }

        // ���Ľ���
        if (lot->status() == auction::ENDED) {
            ctx->ok(ErrorAuctionIsOver);
            return;
        }

        const std::string& initiator = ctx->initiator();
        if (initiator.empty()) {
            ctx->error(ErrorInitiatorMissing);
            return;
        }

        std::string chipKey = getChipKey(id, initiator);
        std::string chipStr;
        std::unique_ptr<auction::Chip> chip(new auction::Chip);
        if (!ctx->get_object(chipKey, &chipStr) || chipStr.empty() || !chip->ParseFromString(chipStr)) {
            ctx->error(ErrorGetChip);
            return;
        }

        const std::string& amountStr = ctx->arg("amount");
        if (amountStr.empty()) {
            ctx->error(ErrorAmountMissing);
            return;
        }
        uint64_t amount;
        if (!safe_stoull(amountStr, &amount)) {
            ctx->error(ErrorAmountIllegal);
            return;
        }

        // С�ڵ�ǰ��߾���
        if (amount <= lot->bid() || amount < lot->floor()) {
            ctx->error(ErrorLowerBid);
            return;
        }

        // ���벻��
        if (chip->amount() < amount) {
            ctx->logf("bidder:%s, amount=%lld", chip->bidder().c_str(), chip->amount());
            ctx->error(ErrorChipNotEnough);
            return;
        }

        // ���澺�ļ�¼
        std::string recordKey = getRecordKey(id, amount);
        if (!ctx->put_object(recordKey, initiator)) {
            ctx->error(ErrorSetRecord);
            return;
        }

        // ���³���
        chip->set_bid(amount);
        if (!chip->SerializeToString(&chipStr) || !ctx->put_object(chipKey, chipStr)) {
            ctx->error(ErrorSetChip);
            return;
        }

        // ��������Ʒ��Ϣ
        lot->set_bidder(initiator);
        lot->set_bid(amount);
        if (!lot->SerializeToString(&lotStr) || !ctx->put_object(lotKey, lotStr)) {
            ctx->error(ErrorSetLot);
            return;
        }

        std::string result =
                "amount=" + std::to_string(chip->amount()) + "\t"
                "bid=" + std::to_string(chip->bid());
        ctx->ok(result);
    }

    void deal() {
        xchain::Context* ctx = this->context();
        const std::string& caller = ctx->initiator();
        if (caller.empty()) {
            ctx->error(ErrorInitiatorMissing);
            return;
        }

        const std::string& id = ctx->arg("id");
        if (id.empty()) {
            ctx->error(ErrorIdMissing);
            return;
        }
        std::string lotKey = getLotKey(id);
        std::string lotStr;
        std::unique_ptr<auction::Lot> lot(new auction::Lot);
        if (!ctx->get_object(lotKey, &lotStr) || lotStr.empty() || !lot->ParseFromString(lotStr)) {
            ctx->error(ErrorGetLot);
            return;
        }

        if (lot->status() == auction::ENDED) {
            ctx->ok(ErrorAuctionIsOver);
            return;
        }

        if (lot->auctioneer() != caller) {
            ctx->error(ErrorNoPermission);
            return;
        }

        lot->set_status(auction::ENDED);
        if (!lot->SerializeToString(&lotStr) || !ctx->put_object(lotKey, lotStr)) {
            ctx->error(ErrorSetLot);
            return;
        }

        std::string result = lot->bidder() + ": " + std::to_string(lot->bid());
        ctx->ok(result);
    }

    void query() {
        xchain::Context* ctx = this->context();

        const std::string& id = ctx->arg("id");
        if (id.empty()) {
            ctx->error(ErrorIdMissing);
            return;
        }
        std::string lotKey = getLotKey(id);
        std::string lotStr;
        std::unique_ptr<auction::Lot> lot(new auction::Lot);
        if (!ctx->get_object(lotKey, &lotStr) || lotStr.empty() || !lot->ParseFromString(lotStr)) {
            ctx->error(ErrorGetLot);
            return;
        }

        std::string result;
        result += "id=" + id +
                "\tauctioneer=" + lot->auctioneer() +
                "\tbidder=" + lot->bidder() +
                "\tbid=" + std::to_string(lot->bid()) +
                "\tfloor=" + std::to_string(lot->floor()) +
                "\n";

        std::string recordKey = getRecordPrefix(id);
        std::unique_ptr<xchain::Iterator> iter =
                ctx->new_iterator(recordKey, recordKey + "~");
        int bidCnt = 0;
        while (iter->next()) {
            std::pair<std::string, std::string> res;
            iter->get(&res);
            if (res.first.length() > recordKey.length()) {
                bidCnt++;
                std::string bid = res.first.substr(recordKey.length());
                std::string bidder = res.second;
                result += "bidder=" + bidder + "\tbid=" + bid + "\n";
            }
        }
        result ="total bid count:" + std::to_string(bidCnt) + "\n" + result;
        ctx->ok(result);
    };
};

DEFINE_METHOD(AuctionDemo, initialize) { self.initialize(); }

DEFINE_METHOD(AuctionDemo, create) { self.create(); }

DEFINE_METHOD(AuctionDemo, chip) { self.chip(); }

DEFINE_METHOD(AuctionDemo, redeem) { self.redeem(); }

DEFINE_METHOD(AuctionDemo, bid) { self.bid(); }

DEFINE_METHOD(AuctionDemo, deal) { self.deal(); }

DEFINE_METHOD(AuctionDemo, query) { self.query(); }