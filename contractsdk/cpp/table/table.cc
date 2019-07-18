#include <inttypes.h>
#include "table/table.h"

namespace xchain { namespace cdt {

const size_t ITERATOR_BATCH_SIZE = 100;

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
bool Table<T>::find(std::string key, T* t) {
    std::string value;
    if (_ctx->get_object(PREFIX_ROWKEY + key, &value)) {
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
    auto no = get_index_no();
    if (no < 0) {
        return false;
    }
    std::string idx = index_str + std::to_string(no);
    if (!_ctx->put_object(idx, rowkey)) {
        return false;
    }
    std::string idx_key = PREFIX_INDEX + rowkey + std::to_string(idx_no);
    if (!_ctx->put_object(idx_key, idx)) {
        return false;
    }

    return inc_index_no(no + 1);
}

template <typename T>
bool Table<T>::delete_index(const std::string &rowkey, const std::string& index_str, int idx_no) {
    std::string idx = PREFIX_INDEX + rowkey + std::to_string(idx_no);
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
TableIterator<T> Table<T>::scan(std::string idx) {
    return TableIterator<T>(_ctx, idx);
}

template <typename T>
bool Table<T>::del(T t) {
    const std::string& key = t->rowkey();

    if (!_ctx->delete_object(PREFIX_ROWKEY + key)) {
        return false;
    }
    for (int i = 0; i < t->index_size(); i ++) {
        if(!delete_index(key, t->index(i), i)) {
            return false;
        }
    }
    return true;
}

template <typename T>
bool Table<T>::put(T t) {
    const std::string& key = t.rowkey();
    std::string value;
    if (_ctx->get_object(PREFIX_ROWKEY + key, &value)) {
        //重复插入 TODO 改成错误码
        std::cout << "duplicated put: " << key << std::endl;
        return false;
    }
    value.clear();
    if (!t.SerializeToString(&value)) {
        return false;
    }
    if (!_ctx->put_object(PREFIX_ROWKEY + key, value)) {
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
    :Iterator(ctx, idx, idx + KEY_END, ITERATOR_BATCH_SIZE),  _ctx(ctx), _index(idx) {
}

}};
