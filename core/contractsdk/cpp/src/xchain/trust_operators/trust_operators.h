#pragma once
#include "xchain/xchain.h"
#include <map>
#include <string>

class TrustOperators {
public:
  TrustOperators(xchain::Context *, const uint32_t);

  struct Operand {
    std::string cipher;
    std::string commitment;
  };

  struct AuthInfo {
    std::string data;
    std::string to;
    std::string pubkey;
    std::string signature;
    std::string kind;
  };

  bool add(const Operand &left_op, const Operand &right_op,
           std::map<std::string, std::string> *result);
  bool sub(const Operand &left_op, const Operand &right_op,
           std::map<std::string, std::string> *result);
  bool mul(const Operand &left_op, const Operand &right_op,
           std::map<std::string, std::string> *result);

  bool authorize(const AuthInfo &auth,
                 std::map<std::string, std::string> *result);

private:
  xchain::Context *_ctx;
  const uint32_t _svn;

  bool binary_ops(const std::string op, const Operand &left_op,
                  const Operand &right_op,
                  std::map<std::string, std::string> *result);
};
