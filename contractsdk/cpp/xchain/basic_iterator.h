#pragma once

#include "xchain/error.h"
#include "xchain/xchain.h"

namespace xchain {

using ElemType = std::pair<std::string, std::string>;
const size_t ITERATOR_BATCH_SIZE = 100;

class Iterator {
public:
    bool next();
    bool get(ElemType* t);
    xchain::Error error;

    Iterator(const std::string& s,
            const std::string& e, size_t l);

private:
    Iterator(const Iterator&);
    Iterator &operator=(const Iterator&);
    bool load();
    bool end();
    bool range_query(const std::string& s, const std::string& e,
            const size_t limit, std::vector<ElemType>* res);

private:
    size_t _it;
    ElemType* _cur_elem;
    ElemType _last_one;
    std::string _start, _limit;
    size_t _cap;
    std::vector<ElemType> _buf;
};

}
