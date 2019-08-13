#include "xchain/xchain.h"
#include "xchain/basic_iterator.h"
#include "xchain/syscall.h"

namespace pb = xchain::contract::sdk;

namespace xchain {

Iterator::Iterator(const std::string& start, const std::string& limit, size_t cap) {
    _cap = cap;
    _start = start;
    _limit = limit;

    if (!load()) {
        _it = -1;
    } else {
        _it = 0;
    }
}

bool Iterator::load() {
    _buf.clear();
    if (_start == _limit) {
        return false;
    }
    //每次多取一个作为下一次的主键
    bool ok = range_query(_start, _limit, _cap + 1, &_buf);
    if (!ok) {
        error = Error(xchain::ErrorType::kErrIteratorLoad);
        return false;
    }
    if (_buf.size() != 0) {
        if (_buf.size() == _cap + 1) {
            _start = (*(_buf.rbegin())).first;
            _buf.pop_back();
        } else {
            _start = _limit;
            //这个时候就已经结束了. 但是留到下次取的时候返回false
        }
    }
    return true;
}

bool Iterator::next() {
    bool ret = end();
    if (ret) {
        //只有最后一批结束的时候，才会走到这里
        return false;
    }
    _cur_elem = &_buf[_it];
    ++_it;
    if (end()) {
        _last_one = *_cur_elem;
        _cur_elem = &_last_one;
        if (!load()) {
            _it = -1;
        } else {
            _it = 0;
        }
    }
    return !ret;
}

bool Iterator::get(ElemType* t) {
    *t = *_cur_elem;
    return true;
}

bool Iterator::end() {
    return _it >= _buf.size() || _it < 0;
}

bool Iterator::range_query(const std::string& s, const std::string& e,
        const size_t limit, std::vector<std::pair<std::string, std::string>>* res) {
    pb::IteratorRequest req;
    pb::IteratorResponse resp;
    req.set_limit(e);
    req.set_start(s);
    req.set_cap(limit);
    bool ok = syscall("NewIterator", req, &resp);
    if (!ok) {
        return false;
    }
    for (auto item : resp.items()) {
        res->emplace_back(item.key(), item.value());
    }
    return true;
}

}
