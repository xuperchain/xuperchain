#pragma once
#include <map>
#include <string>
#include "xchain/xchain.h"


class TrustOperators {
public:
    TrustOperators(xchain::Context*, const uint32_t);

    std::string add(const std::string left_value, const std::string right_value, const std::string output_key);
    std::string sub(const std::string left_value, const std::string right_value, const std::string output_key);
    std::string mul(const std::string left_value, const std::string right_value, const std::string output_key);
   
private:
    xchain::Context* _ctx;
    const uint32_t _svn;
};

