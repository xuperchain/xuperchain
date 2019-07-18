#include "table/basic_iterator.h"
#include "xchain/xchain.h"

namespace xchain { namespace cdt {

Iterator::Iterator(xchain::Context* ctx, const std::string start, const std::string limit, size_t cap) {
    _ctx = ctx;
    _cap = cap;
    _start = start;
    _limit = limit;

    if (!load()) {
        _it = _buf.size();
    } else {
        _it = 0;
    }
}

bool Iterator::load() {
    assert(_ctx);
    _buf.clear();
    //每次多取一个作为下一次的主键
    bool ok = _ctx->range_query(_start, _limit, _cap + 1, &_buf);
    if (!ok) {
        return false;
    }
    if (_buf.size() != 0) {
        if (_buf.size() == _cap + 1) {
            _start = (*(_buf.rbegin())).first;
            _buf.pop_back();
        } else {
            _start = _limit;
        }
    }
    return true;
}

bool Iterator::next() {
    ++_it;
    if (end()) {
        if (!load()) {
            _it = -1;
            return false;
        }
        _it = 0;
    }
    return _it >= 0;
}

bool Iterator::get(ElemType* t) {
    if(end()) {
        return false;
    }
    *t = _buf[_it];
    return true;
}

bool Iterator::end() {
    return _it >= _buf.size() || _it < 0;
}
}}
