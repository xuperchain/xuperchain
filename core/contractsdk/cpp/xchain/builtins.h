#ifndef XCHAIN_BUILTINS_H
#define XCHAIN_BUILTINS_H

#include <string>

namespace xchain {
// sha256 returns the sha256 sum of input as hex string
std::string sha256(const std::string& input);

// sha256raw returns the sha256 sum of input as raw bytes
std::string sha256raw(const std::string& input);

}  // namespace xchain

#endif