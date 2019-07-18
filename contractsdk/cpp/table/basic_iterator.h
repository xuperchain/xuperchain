#pragma once

#include "xchain/contract.pb.h"
#include "xchain/xchain.h"

namespace xchain { namespace cdt {

using ElemType = std::pair<std::string, std::string>;

class Iterator {
public:
    bool next();
    bool get(ElemType* t);
    bool end();

    Iterator(xchain::Context* ctx, const std::string s, const std::string e, size_t l);

private:
    bool load();
    size_t _it;
    std::string _start, _limit;
    size_t _cap;
    xchain::Context *_ctx;
    std::vector<ElemType> _buf;
};

}}
