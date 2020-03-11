#include "xchain/builtins.h"
#include <string>

extern "C" void xvm_hash(const char* name, const char* inputptr, int inputlen,
                         char* outputptr, int outputlen, int hex_encode);

namespace xchain {
std::string sha256(const std::string& input) {
    char out[64];
    xvm_hash("sha256", &input[0], input.size(), out, sizeof(out), 1);
    return std::string(out, sizeof(out));
}

std::string sha256raw(const std::string& input) {
    char out[32];
    xvm_hash("sha256", &input[0], input.size(), out, sizeof(out), 0);
    return std::string(out, sizeof(out));
}

}  // namespace xchain