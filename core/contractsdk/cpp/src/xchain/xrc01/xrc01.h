#pragma once

#include "xrc01.pb.h"
#include "xchain/table/table.tpl.h"
#include "xchain/table/types.h"
#include "xchain/xchain.h"

// XRC01 is the implement of XRC-01 multi token standard
class XRC01 {
public:
    XRC01(xchain::Context* ctx)
        : _token(ctx, "token"),
          _asset_info(ctx, "asset_info"),
          _authorize_info(ctx, "authorize_info") {
        _ctx = ctx;
    }

    struct token : public xrc01::token {
        DEFINE_ROWKEY(id);
        DEFINE_INDEX_BEGIN(0)
        DEFINE_INDEX_END();
    };

    struct asset_info : public xrc01::asset_info {
        DEFINE_ROWKEY(id);
        DEFINE_INDEX_BEGIN(1)
        DEFINE_INDEX_ADD(0, token_id)
        DEFINE_INDEX_END();
    };

    struct authorize_info : public xrc01::authorize_info {
        DEFINE_ROWKEY(id);
        DEFINE_INDEX_BEGIN(2)
        DEFINE_INDEX_ADD(0, from, token_id)
        DEFINE_INDEX_ADD(1, to, token_id)
        DEFINE_INDEX_END();
    };

private:
    // define token table
    xchain::cdt::Table<token> _token;
    // define asset_info table
    xchain::cdt::Table<asset_info> _asset_info;
    // define authorize_info table
    xchain::cdt::Table<authorize_info> _authorize_info;
    // _ctx of the contract call
    xchain::Context* _ctx;

public:
    // following interfaces declare the main interfaces of XRC01 multi token
    // standard issue a token
    bool issue(XRC01::token* token);
    // authorization token to someone
    bool authorization(const std::string& to, uint64_t token_id,
                       uint64_t amount);
    // withdraw authorization token from someone
    bool withdraw_authorization(const std::string& from, uint64_t token_id,
                                uint64_t amount);
    // transfer token to someone, from is the caller of the contract
    bool transfer(const std::string& to, uint64_t token_id, uint64_t amount);
    // authorize transfer token from someone to someone, the caller need to be
    // authorized by from
    bool transfer_from(const std::string& from, const std::string& to,
                       uint64_t token_id, uint64_t amount);
    // get_balance return the token balance of given account
    bool get_balance(const std::string& account, uint64_t token_id,
                     uint64_t* balance);
    // get_authorized return the token have authorized to others
    bool get_authorized(const std::string& account, uint64_t token_id,
                        uint64_t* authorized);
    // owner_of return the owner of an non-fungible token
    bool owner_of(const uint64_t token_id, std::string* owner);
    // authorize_infos return the token authorize infos of an account
    bool authorize_infos(const std::string& account, uint64_t token_id,
                         std::vector<XRC01::authorize_info>& authorize_infos);
    // authorized_infos return tokens the account be authorized
    bool authorized_infos(const std::string& account, uint64_t token_id,
                          std::vector<XRC01::authorize_info>& authorize_infos);
};
