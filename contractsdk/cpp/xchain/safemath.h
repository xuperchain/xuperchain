#ifndef XCHAIN_SAFEMATH_H
#define XCHAIN_SAFEMATH_H

// TODO: this just call the system assert, will implement in near future.
namespace xchain {
    bool safe_assert(bool value) {
        assert(value);
        return true;
    };
}
#endif