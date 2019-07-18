#include <cstdio>
#include <gtest/gtest.h>
#include "test/fake_syscall.h"
#include "example/anchor.cc"

class AnchorTest : public ::testing::Test {
protected:
    std::map<std::string, std::string> init_rwset, init_args;
protected:
    void SetUp() override {
        xchain::cdt::ctx_lock();
        init_rwset = {{"key1", "22"}};
        init_args = {
            {"id"   , "1"},
            {"name" , "Bob"},
            {"desc" , "Bob's game"},
        };

        xchain::cdt::ctx_init(init_rwset, "alice", {"ak1", "ak2"}, "initialize", init_args);
        {
            Anchor anchor;
            cxx_initialize(anchor);
        }
    }
    void TearDown() override {
        xchain::cdt::ctx_unlock();
    }
};

TEST_F(AnchorTest, MethodGet) {
    init_args["key"] ="Bob";
    xchain::cdt::ctx_init(init_rwset, "alice", {"ak1", "zzz"}, "get", init_args);
    {
        Anchor anchor;
        cxx_get(anchor);
    }
    // ctx_assert can, only can get the response when the contract object's lifecycle was finished
    ASSERT_EQ(xchain::cdt::ctx_assert(200), true);
}

TEST_F(AnchorTest, MethodSet) {
    init_args = {
        {"id"   , "2"},
        {"name" , "Bob1"},
        {"desc" , "Bob's game"},
    };
    xchain::cdt::ctx_init(init_rwset, "alice", {"ak1", "ak2"}, "set", init_args);
    {
        Anchor anchor;
        cxx_set(anchor);
    }
    ASSERT_EQ(xchain::cdt::ctx_assert(200), true);
}

TEST_F(AnchorTest, MethodScan) {
    Anchor::entity ent;
    ent.set_id(2);
    ent.set_name("Bob");
    ent.set_desc("Bob's game");
    init_args["index"] = ent.index(0);
    xchain::cdt::ctx_init(init_rwset, "alice", {"ak1", "ak2"}, "scan", init_args);
    {
        Anchor anchor;
        cxx_scan(anchor);
    }
    ASSERT_EQ(xchain::cdt::ctx_assert(200), true);
    ASSERT_EQ(xchain::cdt::ctx_assert(200, "", "1"), true);
}

TEST_F(AnchorTest, MethodBatchSetAndScan) {
    for (int i=0; i < 100; i ++) {
        std::string id = std::to_string(i);
        init_args = {
            {"id"   , id},
            {"name" , "Tom" + id},
            {"desc" , "Tom's game"},
        };
        xchain::cdt::ctx_init(init_rwset, "alice", {"ak1", "ak2"}, "set", init_args);
        {
            Anchor anchor;
            cxx_set(anchor);
        }
        ASSERT_EQ(xchain::cdt::ctx_assert(200), true);
    }

    init_args["index"] = "Tom";
    xchain::cdt::ctx_init(init_rwset, "alice", {"xxxx", "ak2"}, "scan", init_args);
    {
        Anchor anchor;
        cxx_scan(anchor);
    }
    ASSERT_EQ(xchain::cdt::ctx_assert(200), true);
    ASSERT_EQ(xchain::cdt::ctx_assert(200, "", "100"), true);
}

int main(int argc, char** argv) {
    remove(xchain::cdt::kFileName.c_str());
    ::testing::InitGoogleTest(&argc, argv);
    return RUN_ALL_TESTS();
}
