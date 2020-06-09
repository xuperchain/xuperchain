#pragma once
#include "xchain/xchain.h"
#include <map>
#include <string>

class TrustOperators {
public:
  TrustOperators(xchain::Context *, const uint32_t);

  struct Operand {
    std::string cipher;     // ciphertext
    std::string commitment; // user's proof of right to use data
  };

  // information for authorization request
  struct AuthInfo {
    std::string data;      // data content
    std::string to;        // user address to be authorized
    std::string pubkey;    // public key of owner
    std::string signature; // owner's signature for this authorization request
    std::string kind; //"commitment" for data usage, "ownership" for data ownership
  };

  bool add(const Operand &left_op, const Operand &right_op,
           std::map<std::string, std::string> *result);
  bool sub(const Operand &left_op, const Operand &right_op,
           std::map<std::string, std::string> *result);
  bool mul(const Operand &left_op, const Operand &right_op,
           std::map<std::string, std::string> *result);

  bool authorize(const AuthInfo &auth,
                 std::map<std::string, std::string> *result);

  // paillier homomorphic addition, returns encryption of addtion of two
  // plaintext
  bool paillier_add(const Operand &left_op, const Operand &right_op,
                    const std::string pubkey,
                    std::map<std::string, std::string> *result);
  // paillier partially homomorphic multiplication, returns encryption of
  // multiplication of plaintext and a number
  bool paillier_partial_mul(const Operand &left_op, const std::string scalar,
                            const std::string pubkey,
                            std::map<std::string, std::string> *result);

private:
  xchain::Context *_ctx;
  const uint32_t _svn;

  bool binary_ops(const std::string op, const Operand &left_op,
                  const Operand &right_op,
                  std::map<std::string, std::string> *result);
};
