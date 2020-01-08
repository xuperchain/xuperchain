/*
 * XRC-01 is a multi token standard used for smart contracts on the XuperChain
 * blockchain. It's can include combination of fungible and non-fungible tokens
 * in one contract.
 */
#include "xrc01.h"

#include "xchain/safemath.h"

std::string make_asset_info_id(const std::string account, uint64_t token_id) {
    return account + std::to_string(token_id);
}

std::string make_authorize_info_id(const std::string from, const std::string to,
                                   uint64_t token_id) {
    return from + to + std::to_string(token_id);
}

bool XRC01::issue(XRC01::token* token) {
    uint64_t token_id = token->id();
    XRC01::token token_get;
    if (_token.find({{"id", std::to_string(token_id)}}, &token_get)) {
        printf("Issued failed, the token with id %llu has been created! \n",
               token_id);
        return false;
    }
    if (!token->fungible() && token->supply() != 1) {
        printf("Issued failed, non-Fungible token need supply only one \n");
        return false;
    }
    if (!_token.put(*token)) {
        printf("Issued failed, store token failed, token id %llu \n", token_id);
        return false;
    }

    XRC01::asset_info asset_info;
    std::string asset_info_id =
        make_asset_info_id(token->issue_account(), token_id);
    asset_info.set_id(asset_info_id);
    asset_info.set_account(token->issue_account());
    asset_info.set_token_id(token_id);
    asset_info.set_amount(token->supply());
    asset_info.set_authorized(0);
    if (!_asset_info.put(asset_info)) {
        printf("Issued failed, store asset info failed, token id %llu \n ",
               token_id);
        return false;
    }
    return true;
}

bool XRC01::authorization(const std::string& to, uint64_t token_id,
                          uint64_t amount) {
    XRC01::token token;
    if (!_token.find({{"id", std::to_string(token_id)}}, &token)) {
        printf("Token with token_id=%llu haven't issued. \n", token_id);
        return false;
    }

    xchain::Account account = _ctx->sender();
    std::string caller = account.get_name();
    if (caller == to) {
        printf(
            "Authorization failed, can not authorization to self, \
         from=%s, to=%s \n",
            caller.c_str(), to.c_str());
        return false;
    }

    XRC01::asset_info asset_info;
    XRC01::authorize_info authorize_info;
    std::string asset_info_id = make_asset_info_id(caller, token_id);
    if (!_asset_info.find({{"id", asset_info_id}}, &asset_info)) {
        printf(
            "Authorization failed, you do not have token, \
         from=%s, to=%s, token_id=%llu, amount=%llu \n",
            caller.c_str(), to.c_str(), token_id, amount);
        return false;
    }

    if (!xchain::safe_assert(asset_info.amount() >= asset_info.authorized())) {
        printf("safemath assert failed");
        return false;
    }
    if (asset_info.amount() - asset_info.authorized() < amount) {
        printf(
            "Authorization failed, asset remained not enough, \
         from=%s, to=%s, token_id=%llu, amount=%llu, authorized=%llu \n",
            caller.c_str(), to.c_str(), token_id, amount,
            asset_info.authorized());
        return false;
    }

    // Update asset_info
    if (!xchain::safe_assert(asset_info.authorized() + amount >= amount)) {
        printf("safemath assert failed");
        return false;
    }
    asset_info.set_authorized(asset_info.authorized() + amount);
    if (!_asset_info.update(asset_info)) {
        printf(
            "Authorization failed, update asset_info error,\
         from=%s, to=%s, token_id=%llu, amount=%llu \n",
            caller.c_str(), to.c_str(), token_id, amount);
        return false;
    }

    // Update authorize_info
    std::string authorize_info_id =
        make_authorize_info_id(caller, to, token_id);
    if (!_authorize_info.find({{"id", authorize_info_id}}, &authorize_info)) {
        authorize_info.set_id(authorize_info_id);
        authorize_info.set_from(caller);
        authorize_info.set_to(to);
        authorize_info.set_token_id(token_id);
        authorize_info.set_amount(amount);
        if (!_authorize_info.put(authorize_info)) {
            printf(
                "Authorization failed, store authorization info failed, \
                from=%s, to=%s, token_id=%llu, amount=%llu \n",
                caller.c_str(), to.c_str(), token_id, amount);
            return false;
        }
    } else {
        if (!xchain::safe_assert(amount + authorize_info.amount() >=
                                 authorize_info.amount())) {
            printf("safemath assert failed");
            return false;
        }
        authorize_info.set_amount(amount + authorize_info.amount());
        if (!_authorize_info.update(authorize_info)) {
            printf(
                "Authorization failed, update authorization info failed, \
                from=%s, to=%s, token_id=%llu, amount=%llu \n",
                caller.c_str(), to.c_str(), token_id, amount);
            return false;
        }
    }
    return true;
}

bool XRC01::withdraw_authorization(const std::string& to, uint64_t token_id,
                                   uint64_t amount) {
    XRC01::token token;
    if (!_token.find({{"id", std::to_string(token_id)}}, &token)) {
        printf("Token with token_id=%llu haven't issued. \n", token_id);
        return false;
    }

    xchain::Account account = _ctx->sender();
    const std::string caller = account.get_name();

    XRC01::asset_info asset_info;
    XRC01::authorize_info authorize_info;

    std::string asset_info_id = make_asset_info_id(caller, token_id);
    if (!_asset_info.find({{"id", asset_info_id}}, &asset_info)) {
        printf(
            "Withdraw_authorization failed, you do not have token, "
            "token_id=%llu \n",
            token_id);
        return false;
    }

    std::string authorize_info_id =
        make_authorize_info_id(caller, to, token_id);
    if (!_authorize_info.find({{"id", authorize_info_id}}, &authorize_info)) {
        printf(
            "Withdraw_authorization failed, you haven't authorized "
            "token_id=%llu to %s \n",
            token_id, to.c_str());
        return false;
    }

    if (authorize_info.amount() < amount) {
        printf(
            "Withdraw_authorization failed, you haven't authorized enough "
            "token \n");
        return false;
    }

    if (!xchain::safe_assert(asset_info.authorized() >= amount)) {
        printf("safemath assert failed");
        return false;
    }
    asset_info.set_authorized(asset_info.authorized() - amount);
    if (!_asset_info.update(asset_info)) {
        printf("Withdraw_authorization failed, store asset_info error \n");
        return false;
    }

    if (!xchain::safe_assert(authorize_info.amount() >= amount)) {
        printf("safemath assert failed");
        return false;
    }
    authorize_info.set_amount(authorize_info.amount() - amount);
    if (authorize_info.amount() - amount == 0) {
        if (!_authorize_info.del(authorize_info)) {
            printf(
                "Withdraw_authorization failed, del authorize_info error \n");
            return false;
        }
    } else {
        if (!_authorize_info.update(authorize_info)) {
            printf(
                "Withdraw_authorization failed, store authorize_info error \n");
            return false;
        }
    }
    printf("Withdraw_authorization true \n");
    return true;
}

bool XRC01::transfer(const std::string& to, uint64_t token_id,
                     uint64_t amount) {
    XRC01::token token;
    if (!_token.find({{"id", std::to_string(token_id)}}, &token)) {
        printf("Token with token_id=%llu haven't issued. \n", token_id);
        return false;
    }

    xchain::Account account = _ctx->sender();
    const std::string caller = account.get_name();
    XRC01::asset_info asset_info_caller;
    XRC01::asset_info asset_info_to;

    if (caller == to || amount < 1) {
        printf(
            "transfer_from failed, at least transfer 1 token and from can not "
            "equal with to. \n");
        return false;
    }

    std::string asset_info_caller_id = make_asset_info_id(caller, token_id);
    if (!_asset_info.find({{"id", asset_info_caller_id}}, &asset_info_caller)) {
        printf("Transfer failed, you do not have token, token_id=%llu \n",
               token_id);
        return false;
    }

    if (!xchain::safe_assert(asset_info_caller.amount() >=
                             asset_info_caller.authorized())) {
        printf("safemath assert failed");
        return false;
    }
    if (asset_info_caller.amount() - asset_info_caller.authorized() < amount) {
        printf(
            "Transfer failed, you do not have enough token, token_id=%llu \n",
            token_id);
        return false;
    }

    if (!xchain::safe_assert(asset_info_caller.amount() >= amount)) {
        printf("safemath assert failed");
        return false;
    }
    asset_info_caller.set_amount(asset_info_caller.amount() - amount);

    std::string asset_info_to_id = make_asset_info_id(to, token_id);
    if (!_asset_info.find({{"id", asset_info_to_id}}, &asset_info_to)) {
        asset_info_to.set_id(asset_info_to_id);
        asset_info_to.set_account(to);
        asset_info_to.set_token_id(token_id);
        asset_info_to.set_authorized(0);
        asset_info_to.set_amount(amount);
        if (!_asset_info.put(asset_info_to)) {
            printf("Transfer failed, put asset_info_to failed! \n");
            return false;
        }
    } else {
        if (!xchain::safe_assert(asset_info_to.amount() + amount >= amount)) {
            printf("safemath assert failed");
            return false;
        }
        asset_info_to.set_amount(amount + asset_info_to.amount());
        if (!_asset_info.update(asset_info_to)) {
            printf("Transfer failed, update asset_info_to failed! \n");
            return false;
        }
    }

    if (asset_info_caller.amount() == 0) {
        if (!_asset_info.del(asset_info_caller)) {
            printf("Transfer failed, del asset_info_caller failed! \n");
            return false;
        }
    } else {
        if (!_asset_info.update(asset_info_caller)) {
            printf("Transfer failed, store asset_info_caller failed! \n");
            return false;
        }
    }
    return true;
}

bool XRC01::transfer_from(const std::string& from, const std::string& to,
                          uint64_t token_id, uint64_t amount) {
    XRC01::token token;
    if (!_token.find({{"id", std::to_string(token_id)}}, &token)) {
        printf("Token with token_id=%llu haven't issued. \n", token_id);
        return false;
    }

    if (from == to || amount < 1) {
        printf(
            "transfer_from failed, at least transfer 1 token and from can not "
            "equal with to. \n");
        return false;
    }

    xchain::Account account = _ctx->sender();
    const std::string caller = account.get_name();

    XRC01::asset_info asset_info_from;
    XRC01::asset_info asset_info_to;
    XRC01::authorize_info authorize_info;

    std::string authorize_info_id = make_authorize_info_id(from, caller, token_id);
    if (!_authorize_info.find({{"id", authorize_info_id}}, &authorize_info)) {
        printf(
            "Transfer_from failed, from=%s to=%s, caller=%s can not be "
            "authorized. \n",
            from.c_str(), to.c_str(), caller.c_str());
        return false;
    }

    if (authorize_info.amount() < amount) {
        printf("Transfer_from failed, the amount of authorized not enough. \n");
        return false;
    }

    if (!xchain::safe_assert(authorize_info.amount() >= amount)) {
        printf("safemath assert failed");
        return false;
    }
    authorize_info.set_amount(authorize_info.amount() - amount);

    std::string asset_info_from_id = make_asset_info_id(from, token_id);
    if (!_asset_info.find({{"id", asset_info_from_id}}, &asset_info_from)) {
        printf("Transfer_from failed, query asset_info_from error. \n");
        return false;
    }

    if (!xchain::safe_assert(asset_info_from.amount() >= amount)) {
        printf("safemath assert failed");
        return false;
    }
    asset_info_from.set_amount(asset_info_from.amount() - amount);

    if (!xchain::safe_assert(asset_info_from.authorized() >= amount)) {
        printf("safemath assert failed");
        return false;
    }
    asset_info_from.set_authorized(asset_info_from.authorized() - amount);

    std::string asset_info_to_id = make_asset_info_id(to, token_id);
    if (!_asset_info.find({{"id", asset_info_to_id}}, &asset_info_to)) {
        asset_info_to.set_id(asset_info_to_id);
        asset_info_to.set_account(to);
        asset_info_to.set_token_id(token_id);
        asset_info_to.set_amount(amount);
        asset_info_to.set_authorized(0);
        if (!_asset_info.put(asset_info_to)) {
            printf("transfer_from failed, store asset_info_to failed! \n");
            return false;
        }
    } else {
        if (!xchain::safe_assert(asset_info_to.amount() + amount >= amount)) {
            printf("safemath assert failed");
            return false;
        }
        asset_info_to.set_amount(asset_info_to.amount() + amount);
        if (!_asset_info.update(asset_info_to)) {
            printf("transfer_from failed, store asset_info_to failed! \n");
            return false;
        }
    }

    if (asset_info_from.amount() == 0) {
        if (!_asset_info.del(asset_info_from)) {
            printf("transfer_from failed, del asset_info_from failed! \n");
            return false;
        }
    } else {
        if (!_asset_info.update(asset_info_from)) {
            printf("transfer_from failed, store asset_info_from failed! \n");
            return false;
        }
    }

    if (authorize_info.amount() == 0) {
        if (!_authorize_info.del(authorize_info)) {
            printf("transfer_from failed, del authorize_info failed! \n");
            return false;
        }
    } else {
        if (!_authorize_info.update(authorize_info)) {
            printf("transfer_from failed, del authorize_info failed! \n");
            return false;
        }
    }
    return true;
}

bool XRC01::get_balance(const std::string& account, uint64_t token_id,
                        uint64_t* balance) {
    XRC01::token token;
    if (!_token.find({{"id", std::to_string(token_id)}}, &token)) {
        printf("Token with token_id=%llu haven't issued. \n", token_id);
        return false;
    }

    XRC01::asset_info asset_info;
    std::string asset_info_id = make_asset_info_id(account, token_id);
    if (!_asset_info.find({{"id", asset_info_id}}, &asset_info)) {
        printf("Get_balance failed, query asset_info error. \n");
        return false;
    }

    *balance = asset_info.amount();
    return true;
}

bool XRC01::get_authorized(const std::string& account, uint64_t token_id,
                           uint64_t* authorized) {
    XRC01::token token;
    if (!_token.find({{"id", std::to_string(token_id)}}, &token)) {
        printf("Token with token_id=%llu haven't issued. \n", token_id);
        return false;
    }

    XRC01::asset_info asset_info;
    std::string asset_info_id = make_asset_info_id(account, token_id);
    if (!_asset_info.find({{"id", asset_info_id}}, &asset_info)) {
        printf("Get_authorized failed, query asset_info error. \n");
        return false;
    }
    *authorized = asset_info.authorized();
    return true;
}

bool XRC01::owner_of(const uint64_t token_id, std::string* owner) {
    XRC01::token token;
    if (!_token.find({{"id", std::to_string(token_id)}}, &token)) {
        printf("Token with token_id=%llu haven't issued. \n", token_id);
        return false;
    }

    if (token.fungible()) {
        printf("Token with token_id=%llu is Fungible. \n", token_id);
        return false;
    }

    XRC01::asset_info asset_info;
    auto it = _asset_info.scan({{"token_id", std::to_string(token_id)}});
    while (it->next()) {
        if (!it->get(&asset_info)) {
            std::cout << "get owner_of error" << std::endl;
            return false;
        }
    }
    *owner = asset_info.account();
    return true;
}

bool XRC01::authorize_infos(
    const std::string& account, uint64_t token_id,
    std::vector<XRC01::authorize_info>& authorize_infos) {
    XRC01::token token;
    if (!_token.find({{"id", std::to_string(token_id)}}, &token)) {
        printf("Token with token_id=%llu haven't issued.\n", token_id);
        return false;
    }

    auto it = _authorize_info.scan(
        {{"from", account}, {"token_id", std::to_string(token_id)}});
    while (it->next()) {
        XRC01::authorize_info authorize_info;
        if (!it->get(&authorize_info)) {
            std::cout << "get authorize_infos error" << std::endl;
            continue;
        }
        authorize_infos.push_back(authorize_info);
    }
    return true;
}

bool XRC01::authorized_infos(
    const std::string& account, uint64_t token_id,
    std::vector<XRC01::authorize_info>& authorized_infos) {
    XRC01::token token;
    if (!_token.find({{"id", std::to_string(token_id)}}, &token)) {
        printf("Token with token_id=%llu haven't issued.\n", token_id);
        return false;
    }

    auto it = _authorize_info.scan(
        {{"to", account}, {"token_id", std::to_string(token_id)}});
    while (it->next()) {
        XRC01::authorize_info authorize_info;
        if (it->get(&authorize_info)) {
            authorized_infos.push_back(authorize_info);
        } else {
            printf("get authorized_infos error.\n");
        }
    }
    return true;
}
