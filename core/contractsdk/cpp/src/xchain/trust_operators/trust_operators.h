#pragma once
#include "xchain/xchain.h"
#include <map>
#include <string>

class TrustOperators {
public:
  TrustOperators(xchain::Context *, const uint32_t);

  bool add(const std::string &left_value, const std::string &right_value,
           const std::string &output_key);
  bool sub(const std::string &left_value, const std::string &right_value,
           const std::string &output_key);
  bool mul(const std::string &left_value, const std::string &right_value,
           const std::string &output_key);

private:
  xchain::Context *_ctx;
  const uint32_t _svn;

  bool binary_ops(const std::string op, const std::string &left_value,
                  const std::string &right_value,
                  const std::string &output_key);
};
