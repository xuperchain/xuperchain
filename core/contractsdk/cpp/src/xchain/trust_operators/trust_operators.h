#pragma once
#include "xchain/xchain.h"
#include <map>
#include <string>

class TrustOperators {
public:
  TrustOperators(xchain::Context *, const uint32_t);

  bool add(const std::string &left_value, const std::string &right_value,
           const std::string &output_key, const std::string &commitment,
           const std::string &commitment2, std::string *result);
  bool sub(const std::string &left_value, const std::string &right_value,
           const std::string &output_key, const std::string &commitment,
           const std::string &commitment2, std::string *result);
  bool mul(const std::string &left_value, const std::string &right_value,
           const std::string &output_key, const std::string &commitment,
           const std::string &commitment2, std::string *result);

  bool authorize(const std::string &data, const std::string &address,
                 const std::string &pubkey, const std::string &signature,
                 const std::string &kind, std::string *result);

private:
  xchain::Context *_ctx;
  const uint32_t _svn;

  bool binary_ops(const std::string op, const std::string &left_value,
                  const std::string &right_value, const std::string &output_key,
                  const std::string &commitment, const std::string &commitment2,
                  std::string *result);
};
