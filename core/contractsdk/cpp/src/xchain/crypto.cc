#include "xchain/crypto.h"
#include <string>

extern "C" void xvm_hash(const char* name, const char* inputptr, int inputlen,
                         char* outputptr, int outputlen);
extern "C" void xvm_encode(const char* name, const char* inputptr, int inputlen,
                           char** outputpptr, int* outlen);
extern "C" int xvm_decode(const char* name, const char* inputptr, int inputlen,
                          char** outputpptr, int* outlen);
extern "C" int xvm_ecverify(const char* pubptr, int publen, const char* sigptr,
                            int siglen, const char* hashptr, int hashlen);

namespace xchain {
namespace crypto {
std::string sha256(const std::string& input) {
    char out[32];
    xvm_hash("sha256", (const char*)&input[0], input.size(), out, sizeof(out));
    return std::string(out, sizeof(out));
}

std::string hex_encode(const std::string& input) {
    char* out = NULL;
    int outlen = 0;
    xvm_encode("hex", (const char*)&input[0], input.size(), &out, &outlen);
    std::string ret(out, outlen);
    free(out);
    return ret;
}

bool hex_decode(const std::string& input, std::string* output) {
    char* out = NULL;
    int outlen = 0;
    int ret = 0;
    ret =
        xvm_decode("hex", (const char*)&input[0], input.size(), &out, &outlen);
    if (ret != 0) {
        return false;
    }
    output->assign(out, outlen);
    free(out);
    return true;
}

bool ecverify(const std::string& pubkey, const std::string& sign,
              const std::string& hash) {
    const char* curveptr = NULL;
    int ret = 0;
    ret = xvm_ecverify((const char*)&pubkey[0], pubkey.size(),
                       (const char*)&sign[0], sign.size(),
                       (const char*)&hash[0], hash.size());
    if (ret != 0) {
        return false;
    }
    return true;
}

}  // namespace crypto
}  // namespace xchain