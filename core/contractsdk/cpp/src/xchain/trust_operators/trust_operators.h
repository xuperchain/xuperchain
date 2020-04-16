#pragma once
#include <map>
#include <string>
#include "xchain/xchain.h"


class TrustOperators {
public:
    TrustOperators(const std::string&);
//    bool store(xchain::Context* ctx, const uint32_t svn, const std::string& args);
    std::string debug(xchain::Context* ctx, const uint32_t svn, const std::string& args);
    std::string add(xchain::Context* ctx, const uint32_t svn, const std::string& args);
    std::string sub(xchain::Context* ctx, const uint32_t svn, const std::string& args);
    std::string mul(xchain::Context* ctx, const uint32_t svn, const std::string& args);

    static std::string MapToString(std::map<std::string, std::string> strMap);
   
private: 
    const std::string& _address;
//    xchain::Context* _ctx;
//   const uint32_t _svn;
};

