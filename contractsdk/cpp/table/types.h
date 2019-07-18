#pragma once

#include <inttypes.h>
#include <sstream>
#include <string>
#include <cxxabi.h>
#include <typeinfo>

template <typename T, size_t N>
constexpr size_t xchain_sizeof(T (&arr)[N]) { return N; }

template <typename... Args>
std::string xchain_concat(Args... args) {
    std::ostringstream oss;
    std::initializer_list<int> unused{(oss << args, 0)...};
    return oss.str();
}

#define DISALLOW_COPY_AND_ASSIGN(classname)                         \
private:                                                            \
    classname(const classname &);                                   \
    classname &operator=(const classname &);                        \
public:                                                             \
    classname() = default;

#define DEFINE_ROWKEY(...)                                          \
private:                                                            \
    std::function<std::string()> rowkey_ =                          \
        [&]() -> std::string {return xchain_concat(__VA_ARGS__);};  \
public:                                                             \
    const std::string rowkey() const {                              \
        return rowkey_();                                           \
    }

#define DEFINE_INDEX_BEGIN(N)                                       \
public:                                                             \
    const std::string index(int i) const {                          \
        return index_[i]();                                         \
    }                                                               \
    const size_t index_size() const {                               \
        return xchain_sizeof(index_);                               \
    }                                                               \
private:                                                            \
    std::function<std::string()> index_[N] = {

#define DEFINE_INDEX_ADD(id, ...)                                   \
        [&]()->std::string { return xchain_concat(__VA_ARGS__); },

#define DEFINE_INDEX_END()   };                                     \
  public:                                                           \
    std::string to_str() {                                          \
        std::string value;                                          \
        SerializeToString(&value);                                  \
        return value;                                               \
    }
