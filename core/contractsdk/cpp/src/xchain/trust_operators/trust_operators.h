#pragma once
#include "xchain/xchain.h"
#include <map>
#include <string>

class TrustOperators {
public:
  TrustOperators(xchain::Context *, const uint32_t);

  struct operand {
    std::string cipher;
    std::string commitment;
  };

  struct auth_info {
    std::string data;
    std::string to;
    std::string pubkey;
    std::string signature;
    std::string kind;
  };

  bool add(const operand &left_op, const operand &right_op,
           std::string *result);
  bool sub(const operand &left_op, const operand &right_op,
           std::string *result);
  bool mul(const operand &left_op, const operand &right_op,
           std::string *result);

  bool authorize(const auth_info &auth, std::string *result);

private:
  xchain::Context *_ctx;
  const uint32_t _svn;

  bool binary_ops(const std::string op, const operand &left_op,
                  const operand &right_op, std::string *result);
};
