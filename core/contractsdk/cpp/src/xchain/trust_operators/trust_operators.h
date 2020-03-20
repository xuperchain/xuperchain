#pragma once
#include <map>
#include <string>


class TrustOperators {
public:
    TrustOperators(const std::string&);
    bool store(const uint32_t svn, const std::string& args, std::map<std::string, std::string>* res);
   
private: 
    const std::string& _address;
};

