#include <sstream>

#include "xchain/xchain.h"

// 游戏装备资产模板
// 参数由xchain::Contract中的context提供
class GameAssets {
public:
    /*
     * func: 初始化装备发行账户
     * @param: admin: 哪个address具有管理员权限
     */
    virtual void initialize() = 0;
    /*
     * func: 新增装备类型
     * @param: initiator:交易发起者,只有交易发起者等于admin账户才能执行成功
     * @param: type_id: 游戏状态的参数和属性描述
     * @param: typedesc: 游戏状态的参数和属性描述
     */
    virtual void addAssetType() = 0;
    /*
     * func: 获取所有的装备类型和参数信息
     */
    virtual void listAssetType() = 0;
    /*
     * func:
     * 按照用户查询装备资产，管理员可以查询任意用户，其他用户只能查询自己的资产
     * @param: userid: 管理员可以指定user进行查询
     */
    virtual void getAssetsByUser() = 0;
    /*
     * func: 系统新生成的新装备，发放给特定用户，只能由管理员调用
     * @param: type_id: 游戏装备类型id
     * @param: assetid:
     * 游戏装备唯一id(先从外部获取装备id,也可以实现成一个自增计数器)
     * @param: userid: 获得游戏装备的用户
     */
    virtual void newAssetToUser() = 0;
    /*
     * func: 玩家将自己的装备交易给其他用户, 默认装备持有者是交易发起者
     * @param: to: 装备接收者
     * @param: assetid: 装备id
     */
    virtual void tradeAsset() = 0;
};

struct GameDemo : public GameAssets, public xchain::Contract {
public:
    const std::string ASSETTYPE = "AssetType_";
    const std::string USERASSET = "UserAsset_";
    const std::string ASSET2USER = "Asset2User_";

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

    void addAssetType() {
        xchain::Context* ctx = this->context();
        const std::string& caller = ctx->initiator();
        if (caller.empty()) {
            ctx->error("missing initiator");
            return;
        }

        if (!isAdmin(ctx, caller)) {
            ctx->error("you do not have permission to call this method");
            return;
        }

        const std::string& typeId = ctx->arg("type_id");
        if (typeId.empty()) {
            ctx->error("missing 'type_id' as asset type identity");
            return;
        }

        const std::string& typeDesc = ctx->arg("type_desc");
        if (typeDesc.empty()) {
            ctx->error("missing type_desc");
            return;
        }

        std::string assetTypeKey = ASSETTYPE + typeId;
        std::string value;
        if (ctx->get_object(assetTypeKey, &value)) {
            ctx->error("the type_id is already exist, please check again");
            return;
        }
        ctx->put_object(assetTypeKey, typeDesc);
        ctx->ok(typeId);
    }

    void listAssetType() {
        xchain::Context* ctx = this->context();
        std::unique_ptr<xchain::Iterator> iter =
            ctx->new_iterator(ASSETTYPE, ASSETTYPE + "~");
        std::string result;
        while (iter->next()) {
            std::pair<std::string, std::string> res;
            iter->get(&res);
            if (res.first.length() > ASSETTYPE.length()) {
                result += res.first.substr(ASSETTYPE.length()) + ":" +
                          res.second + '\n';
            }
        }
        ctx->ok(result);
    }

    void getAssetsByUser() {
        xchain::Context* ctx = this->context();
        const std::string& caller = ctx->initiator();
        if (caller.empty()) {
            ctx->error("missing initiator");
            return;
        }
        std::string userId = caller;
        if (isAdmin(ctx, caller)) {
            // admin can get the asset data of other users
            const std::string& userId2 = ctx->arg("user_id");
            if (!userId2.empty()) {
                userId = userId2;
            }
        }

        std::string userAssetKey = USERASSET + userId + "_";
        std::unique_ptr<xchain::Iterator> iter =
            ctx->new_iterator(userAssetKey, userAssetKey + "~");
        std::string result;
        while (iter->next()) {
            std::pair<std::string, std::string> res;
            iter->get(&res);
            if (res.first.length() > userAssetKey.length()) {
                std::string assetId = res.first.substr(userAssetKey.length());
                std::string typeId = res.second;
                std::string assetTypeKey = ASSETTYPE + typeId;
                std::string assetDesc;
                if (!ctx->get_object(assetTypeKey, &assetDesc)) {
                    // asset type id not found ,skip this asset
                    continue;
                }
                result += "assetid=" + assetId + ",type_id=" + typeId +
                          ",assetDesc=" + assetDesc + '\n';
            }
        }
        ctx->ok(result);
    }

    void newAssetToUser() {
        xchain::Context* ctx = this->context();
        const std::string& caller = ctx->initiator();
        if (caller.empty()) {
            ctx->error("missing initiator");
            return;
        }

        if (!isAdmin(ctx, caller)) {
            ctx->error("you do not have permission to call this method");
            return;
        }

        const std::string& userId = ctx->arg("user_id");
        if (userId.empty()) {
            ctx->error("missing userid");
            return;
        }

        const std::string& typeId = ctx->arg("type_id");
        if (typeId.empty()) {
            ctx->error("missing type_id");
            return;
        }
        std::string assetTypeKey = ASSETTYPE + typeId;
        std::string value;

        std::stringstream ss;

        ctx->get_object(assetTypeKey, &value);
        if (value.empty()) {
            ss << "asset type " << typeId << " not found";
            ctx->error(ss.str());
            return;
        }

        const std::string& assetId = ctx->arg("asset_id");
        if (assetId.empty()) {
            ctx->error("missing assetid");
            return;
        }

        std::string assetKey = ASSET2USER + assetId;
        std::string tmp;
        if (ctx->get_object(assetKey, &tmp)) {
            ctx->error("the asset id is already exist, please check again");
            return;
        }

        std::string userAssetKey = USERASSET + userId + "_" + assetId;
        if (!ctx->put_object(userAssetKey, typeId) ||
            !ctx->put_object(assetKey, userId)) {
            ctx->error("failed to generate asset to user");
        }
        ctx->ok(assetId);
    }

    void tradeAsset() {
        xchain::Context* ctx = this->context();
        const std::string& from = ctx->initiator();
        if (from.empty()) {
            ctx->error("missing initiator");
            return;
        }

        const std::string& to = ctx->arg("to");
        if (to.empty()) {
            ctx->error("missing to");
            return;
        }

        const std::string& assetId = ctx->arg("asset_id");
        if (assetId.empty()) {
            ctx->error("missing assetid");
            return;
        }
        std::string userAssetKey = USERASSET + from + "_" + assetId;
        std::string assetType;
        if (!ctx->get_object(userAssetKey, &assetType)) {
            ctx->error("you don't have assetid:" + assetId);
            return;
        }

        if (!ctx->delete_object(userAssetKey)) {
            ctx->error("failed to delete assetid:" + assetId);
            return;
        }

        std::string assetKey = ASSET2USER + assetId;
        std::string newUserAssetKey = USERASSET + to + "_" + assetId;
        if (!ctx->put_object(newUserAssetKey, assetType) ||
            !ctx->put_object(assetKey, to)) {
            ctx->error("failed to save assetid:" + assetId);
            return;
        }
        ctx->ok(assetId);
    };
};

DEFINE_METHOD(GameDemo, initialize) { self.initialize(); }

DEFINE_METHOD(GameDemo, addAssetType) { self.addAssetType(); }

DEFINE_METHOD(GameDemo, listAssetType) { self.listAssetType(); }

DEFINE_METHOD(GameDemo, getAssetsByUser) { self.getAssetsByUser(); }

DEFINE_METHOD(GameDemo, newAssetToUser) { self.newAssetToUser(); }

DEFINE_METHOD(GameDemo, tradeAsset) { self.tradeAsset(); }