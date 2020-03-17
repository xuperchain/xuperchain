#pragma once

#include <inttypes.h>
#include <sstream>
#include <string>

template <typename T, size_t N>
constexpr size_t xchain_sizeof(T (&arr)[N]) { return N; }

template <typename... Args>
const std::string xchain_concat(Args... args) {
    std::ostringstream oss;
    std::initializer_list<int> unused{(oss << args << ",", 0)...};
    auto raw = oss.str();
    return raw.substr(0, raw.size() - 1);
}

#define DISALLOW_COPY_AND_ASSIGN(classname)                                             \
private:                                                                                \
    classname(const classname &);                                                       \
    classname &operator=(const classname &);                                            \
public:                                                                                 \
    classname() = default;


#define DEFINE_INDEX_BEGIN(N)                                                           \
public:                                                                                 \
    std::string index(int i) {                                                          \
        auto id = std::to_string(i);                                                    \
        auto f = index_[id]();                                                          \
        return index_[f]();                                                             \
    }                                                                                   \
    std::string index_str(int i)    {                                                   \
        auto id = std::to_string(i);                                                    \
        return index_[id]();                                                            \
    }                                                                                   \
    size_t index_size() {                                                               \
        return index_.size()/2;                                                         \
    }                                                                                   \
    bool has(std::string index) {                                                       \
        return index_.find(index) != index_.end();                                      \
    }                                                                                   \
private:                                                                                \
    std::map<std::string, std::function<std::string()>> index_ = {

#define DEFINE_INDEX_ADD_0(id)                                                          \
        { #id ,                                                                         \
            [&]() -> std::string { return xchain_concat(#id); } },

#define DEFINE_INDEX_ADD_1(id, _1)                                                      \
        { xchain_concat(#_1) ,                                                          \
            [&]() -> std::string { return xchain_concat(#_1, _1()); } },                \
        { #id ,                                                                         \
            [&]() -> std::string { return xchain_concat(#_1); } },

#define DEFINE_INDEX_ADD_2(id, _1, _2)                                                  \
        { xchain_concat(#_1, #_2) ,                                                     \
            [&]() -> std::string { return xchain_concat(#_1, _1(), #_2, _2()); } },     \
        { #id ,                                                                         \
            [&]() -> std::string { return xchain_concat(#_1, #_2); } },

#define DEFINE_INDEX_ADD_3(id, _1, _2, _3)                                              \
        { xchain_concat(#_1, #_2, #_3),                                                 \
            [&]() -> std::string { return xchain_concat(#_1, _1(), #_2, _2(), #_3, _3()); } },\
        { #id ,                                                                         \
            [&]() -> std::string { return xchain_concat(#_1, #_2, #_3); } },

#define DEFINE_INDEX_ADD_4(id, _1, _2, _3, _4)                                          \
        { xchain_concat(#_1, #_2, #_3, #_4),                                            \
            [&]() -> std::string { return xchain_concat(#_1, _1(), #_2, _2(), #_3, _3(), #_4, _4()); } },\
        { #id ,                                                                         \
            [&]() -> std::string { return xchain_concat(#_1, #_2, #_3, #_4); } },

#define OVERLOAD_MACRO(_1, _2, _3, _4, _5, NAME, ...) NAME
#define DEFINE_INDEX_ADD(...) OVERLOAD_MACRO(__VA_ARGS__, DEFINE_INDEX_ADD_4,           \
        DEFINE_INDEX_ADD_3, DEFINE_INDEX_ADD_2, DEFINE_INDEX_ADD_1,                     \
        DEFINE_INDEX_ADD_0, ...)(__VA_ARGS__)

#define DEFINE_INDEX_END()   };                                                         \
  public:                                                                               \
    std::string to_str() {                                                              \
        std::string value;                                                              \
        SerializeToString(&value);                                                      \
        return value;                                                                   \
    }


#define DEFINE_ROWKEY(...)                                                              \
public:                                                                                 \
    std::string rowkey() {                                                              \
        std::string id = "0";                                                           \
        auto f = rowkey_[id]();                                                         \
        return rowkey_[f]();                                                            \
    }                                                                                   \
    std::string rowkey_str() {                                                          \
        std::string id = "0";                                                           \
        return rowkey_[id]();                                                           \
    }                                                                                   \
private:                                                                                \
    std::map<std::string, std::function<std::string()>> rowkey_ = {                     \
        DEFINE_INDEX_ADD(0, __VA_ARGS__)                                                \
    }
