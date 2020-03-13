#ifndef XCHAIN_CRYPTO_H
#define XCHAIN_CRYPTO_H

#include <string>

namespace xchain {
namespace crypto {
// sha256 returns the sha256 sum of input as hex string
std::string sha256(const std::string& input);
// hex_encode returns the hex encoding of input
std::string hex_encode(const std::string& intput);
// hex_decode returns the hex decoding of input
// if ret < 0, error happens, otherwise returns the length written
bool hex_decode(const std::string& intput, std::string* output);
}  // namespace crypto
}  // namespace xchain

#endif