#pragma once
#include <map>
#include <string>
#include "xchain/xchain.h"


class TrustOperators {
public:
    TrustOperators(const std::string&);
    bool store(xchain::Context* ctx, const uint32_t svn, const std::string& args);	
   
private: 
    const std::string& _address;
};

