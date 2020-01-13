#include <cstdio>
#include <gtest/gtest.h>
#include "test/fake_syscall.h"
#include "example/builtin_types.cc"

class BuiltinTypesTest : public ::testing::Test {
protected:
    std::map<std::string, std::string> init_rwset, init_args;
protected:
    void SetUp() override {
        xchain::cdt::ctx_lock();
        init_rwset = {{"key1", "22"}};
        init_args = {};

        xchain::cdt::ctx_init(init_rwset, "alice", {"ak1", "ak2"}, "initialize", init_args);
        {
            BuiltinTypes builtinTypes;
            cxx_initialize(builtinTypes);
        }
    }
    void TearDown() override {
        xchain::cdt::ctx_unlock();
    }
};

TEST_F(BuiltinTypesTest, MethodGetTx) {
    init_args["txid"] ="c9d3390118c509b094c6e2cf4b369d849ce2dd50f2254a54e9a9b5626d7d9422";
    xchain::cdt::ctx_init(init_rwset, "alice", {"ak1", "zzz"}, "gettx", init_args);
    {
        BuiltinTypes builtinTypes;
        cxx_gettx(builtinTypes);
    }
    // ctx_assert can, only can get the response when the contract object's lifecycle was finished
    ASSERT_EQ(xchain::cdt::ctx_assert(200), true);
}

TEST_F(BuiltinTypesTest, MethodGetBlock) {
    init_args["blockid"] ="5a9266b17608dce11f84bddd9a3eae37cf36d3a4f33fd95b53e25077e6e16757";
    xchain::cdt::ctx_init(init_rwset, "alice", {"ak1", "ak2"}, "getblock", init_args);
    {
        BuiltinTypes builtinTypes;
        cxx_getblock(builtinTypes);
    }
    ASSERT_EQ(xchain::cdt::ctx_assert(200), true);
}

int main(int argc, char** argv) {
    remove(xchain::cdt::kFileName.c_str());
    ::testing::InitGoogleTest(&argc, argv);
    return RUN_ALL_TESTS();
}
