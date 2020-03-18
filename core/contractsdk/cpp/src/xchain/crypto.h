#ifndef XCHAIN_CRYPTO_H
#define XCHAIN_CRYPTO_H

#include <string>

namespace xchain {
namespace crypto {
// sha256 returns the sha256 sum of input as bytes
std::string sha256(const std::string& input);
// hex_encode returns the hex encoding of input
std::string hex_encode(const std::string& intput);
// hex_decode returns the hex decoding of input
// if ret false, input is an invalid hex string
bool hex_decode(const std::string& intput, std::string* output);
// ecverify verify elliptic curve signature
// pubkey is the format of json in tx and block
bool ecverify(const std::string& pubkey, const std::string& sign,
              const std::string& hash);
}  // namespace crypto
}  // namespace xchain

#endif