#include <cstdio>
#include <sstream>
#include "util.h"

namespace xchain {

std::string hex2string(std::string const &s)
{
    std::string ret;
    std::istringstream iss(s);
    for (std::string buf; std::getline(iss, buf, ' ');)
    {
        unsigned int value;
        sscanf(buf.c_str(), "%x", &value);
        ret += ((char)value);
    }
    return ret;
}

} // namespace xchain
