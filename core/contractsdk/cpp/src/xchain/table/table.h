#pragma once

#include <memory>
#include "xchain/xchain.h"
#include "xchain/table/types.h"
#include "xchain/basic_iterator.h"

namespace xchain { namespace cdt {

const std::string PREFIX_ROWKEY = "K";
const std::string PREFIX_INDEX  = "I";
const std::string PREFIX_INDEX_DEL  = "D";
const std::string PREFIX_META  = "M";

template<typename T>
class TableIterator;

typedef std::pair<std::string, std::string> PairType;


// Table implements table storage, with pb::Message as a record, supporting rowkey which should be unique,
// and multiple indeies.
// Layout on KV:
//      M{table name}            -> max index id                           (1)
//      K{rowkey}                -> row                                    (2)
//      I{index i}{max index id} -> {rowkey}                               (3)
//      D{rowkey}{i}              ->   I{index i}{max index id}            (4)
// (1) generates unique index id for (3) and (4)
// (2) store the serialized object
// (3) store the index, with i as suffix, you can range_quey {index i} to get all the rows indexed by {index i}
// (4) {rowkey}{i} is also unique, being used to find the index when deleting
template <typename T>
struct Table {
public:
    Table(xchain::Context* ctx, const std::string&);
    virtual ~Table();
    // find a row by key, which is rowkey, if return true, t will bring back the deserialized object.
    bool find(std::initializer_list<PairType>, T* t);
    // scan a index, return a iterator
    std::unique_ptr<TableIterator<T>> scan(std::initializer_list<PairType>);
    // del a record by key
    bool del(T t);
    // put insert a new row.
    bool put(T t);
    // update a row.
    bool update(T t);    

private:
    std::string _table_name;
    xchain::Context* _ctx;

    bool write_index(const std::string& rowkey, const std::string& index_str, int i);
    bool delete_index(const std::string &rowkey, const std::string& index_str, int i);
    std::string make_key(const std::string& prefix, const std::string& value);
    std::string make_index(std::initializer_list<PairType> input, bool key,
        const std::vector<std::string>& filter = std::vector<std::string>());
    int64_t get_index_no();
    bool inc_index_no(int64_t no);
    std::vector<std::string> split(const std::string& input, const std::string& regex);
};

// TableIterator is a iterator to access the retrieved rows.
template <typename T>
class TableIterator : public Iterator {
public:
    TableIterator(xchain::Context* ctx, std::string idx);

    bool get(T* res) {
        // indexName_no -> key
        ElemType et;
        if (!Iterator::get(&et)) {
            return false;
        }

        // key -> row binary
        std::string value;
        if (!_ctx->get_object(et.second, &value)) {
            return false;
        }

        // row bianry -> object
        if (!res->ParseFromString(value)) {
            return false;
        }
        return true;
    }

private:
    xchain::Context* _ctx;
    std::string _index;
};

}} //end of cdt
