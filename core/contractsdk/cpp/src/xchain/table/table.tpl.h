#include <inttypes.h>
#include <iostream>
#include "xchain/table/table.h"

namespace xchain { namespace cdt {

template <typename T>
Table<T>::Table(xchain::Context* ctx, const std::string& name)
    : _table_name(name), _ctx(ctx) {
    //初始化index
    auto no = get_index_no();
    if (no < 0) {
        inc_index_no(1);
    }
}

template <typename T>
Table<T>::~Table(){
}

template <typename T>
std::string Table<T>::make_key(const std::string& prefix, const std::string& value) {
    return _table_name + prefix + value;
}

template <typename T>
bool Table<T>::find(std::initializer_list<PairType> input, T* t) {
    T tmp;
    const std::string& key = make_index(input, false, split(tmp.rowkey_str(), ","));

    std::string value;
    std::string rk = make_key(PREFIX_ROWKEY, key);
    if (_ctx->get_object(rk, &value)) {
        return t->ParseFromString(value);
    }
    return false;
}

template <typename T>
int64_t Table<T>::get_index_no() {
    std::string value;
    std::string key = PREFIX_META + _table_name;
    if (!_ctx->get_object(key,&value)) {
        return -1;
    }
    return std::stoll(value);
}

template <typename T>
bool Table<T>::inc_index_no(int64_t no) {
    std::string value = std::to_string(no);
    std::string key = PREFIX_META + _table_name;
    return _ctx->put_object(key, value);
}

template <typename T>
bool Table<T>::write_index(const std::string& rowkey, const std::string& index_str, int idx_no) {
    const std::string& all_key = make_key(PREFIX_ROWKEY,rowkey);
    auto no = get_index_no();
    if (no < 0) {
        return false;
    }
    
    std::string idx = make_key(PREFIX_INDEX, index_str + std::to_string(no));
    if (!_ctx->put_object(idx, all_key)) {
        return false;
    }
    std::string idx_key = make_key(PREFIX_INDEX_DEL, rowkey + std::to_string(idx_no));
    if (!_ctx->put_object(idx_key, idx)) {
        return false;
    }
    
    return inc_index_no(no + 1);
}

template <typename T>
bool Table<T>::delete_index(const std::string &rowkey, const std::string& index_str, int idx_no) {
    std::string idx = make_key(PREFIX_INDEX_DEL, rowkey + std::to_string(idx_no));
    std::string value;
    if (_ctx->get_object(idx, &value)) {
        // 删除
        if(!_ctx->delete_object(idx)) {
            return false;
        }
        if(!_ctx->delete_object(value)) {
            return false;
        }
    }
    return true;
}

const char KEY_0XFF[2] = {static_cast<char>(0xff), 0};
const std::string KEY_END(KEY_0XFF);

template <typename T>
std::vector<std::string> Table<T>::split(const std::string& input, const std::string& delims) {
    std::vector<std::string> v;
    std::size_t current, previous = 0;
    current = input.find_first_of(delims);
    while (current != std::string::npos) {
        v.push_back(input.substr(previous, current - previous));
        previous = current + 1;
        current = input.find_first_of(delims, previous);
    }
    v.push_back(input.substr(previous, current - previous));
    return v;
}

template <typename T>
std::string Table<T>::make_index(std::initializer_list<PairType> input, bool key,
        const std::vector<std::string>& filter) {

    auto find = [&filter](std::string in) -> bool {
        for (auto &v : filter) {
            if (v == in) {
                return true;
            }
        }
        return false;
    };

    std::ostringstream oss;
    for (auto p = input.begin(); p != input.end(); p ++) {
        //默认传进来的是空的话，就不过滤了
        if (filter.size() > 0) {
            if (!find((*p).first)) {
                continue;
            }
        }
        if (key) {
            oss << (*p).first << ",";
        } else {
            oss << (*p).first << "," << (*p).second << ",";
        }
    }
    auto raw = oss.str();
    return raw.substr(0, raw.size() - 1);
}

template <typename T>
std::unique_ptr<TableIterator<T>> Table<T>::scan(std::initializer_list<PairType> input){
    auto idx = make_index(input, true);
    T t;
    if (!t.has(idx)) {
        auto it = std::unique_ptr<TableIterator<T>>(new TableIterator<T>(_ctx, ""));
        it->error = Error(ErrorType::kErrTableIndexInvalid);
        return std::move(it);
    }
    idx = make_key(PREFIX_INDEX, make_index(input, false));
    return std::unique_ptr<TableIterator<T>>(new TableIterator<T>(_ctx, idx));
}

template <typename T>
bool Table<T>::del(T t) {
    auto key = t.rowkey();
    if (!_ctx->delete_object(make_key(PREFIX_ROWKEY, key))) {
        return false;
    }
    for (int i = 0; i < t.index_size(); i ++) {
        std::string idx;
        if(!delete_index(key, t.index(i), i)) {
            return false;
        }
    }
    return true;
}

template <typename T>
bool Table<T>::put(T t) {
    const std::string& key = t.rowkey();
    const std::string& all_key = make_key(PREFIX_ROWKEY,key);
    std::string value;
    if (_ctx->get_object(all_key, &value)) {
        //重复插入 TODO 改成错误码
        std::cout << "duplicated put: " << key << std::endl;
        return false;
    }
    value.clear();
    if (!t.SerializeToString(&value)) {
        return false;
    }
    if (!_ctx->put_object(all_key, value)) {
        return false;
    }
    for (int i = 0; i < t.index_size(); i ++) {
        if (!write_index(key, t.index(i), i)) {
            return false;
        }
    }
    return true;
}

template <typename T>
bool Table<T>::update(T t) {
    const std::string& key = t.rowkey();
    const std::string& all_key = make_key(PREFIX_ROWKEY,key);
    std::string value;
    if (!_ctx->get_object(all_key, &value)) {
        return false;
    }

    value.clear();
    if (!_ctx->delete_object(all_key)) {
        return false;
    }
    for (int i = 0; i < t.index_size(); i ++) {
        std::string idx;
        if(!delete_index(key, t.index(i), i)) {
            return false;
        }
    }

    if (!t.SerializeToString(&value)) {
        return false;
    }
    if (!_ctx->put_object(all_key, value)) {
        return false;
    }
    for (int i = 0; i < t.index_size(); i ++) {
        if (!write_index(key, t.index(i), i)) {
            return false;
        }
    }
    return true;
}



template <typename T>
TableIterator<T>::TableIterator(xchain::Context* ctx, std::string idx)
    :Iterator(idx, idx + KEY_END, ITERATOR_BATCH_SIZE),  _ctx(ctx), _index(idx) {
}

}};
