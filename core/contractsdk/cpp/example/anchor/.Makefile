CXXFLAGS ?= -std=c++11 -Os -I/usr/local/include -Isrc -Werror=vla -I/Users/wanghongyan01/go/src/github.com/hongyanwang/xuperchain/core/contractsdk/cpp/src
LDFLAGS ?= -Oz -s TOTAL_STACK=256KB -s TOTAL_MEMORY=1MB -s DETERMINISTIC=1 -s EXTRA_EXPORTED_RUNTIME_METHODS=["stackAlloc"] -L/usr/local/lib -lprotobuf-lite -lpthread --js-library /Users/wanghongyan01/go/src/github.com/hongyanwang/xuperchain/core/contractsdk/cpp/src/xchain/exports.js

.PHONY: all build clean

all: build

clean: 
	$(RM) -r build

/Users/wanghongyan01/.xdev-cache/be/be14f170a1b67ef7.o: /Users/wanghongyan01/go/src/github.com/hongyanwang/xuperchain/core/contractsdk/cpp/example/anchor/src/anchor.cc
	@mkdir -p $(dir $@)
	@echo CC $(notdir $<)
	@$(CXX) $(CXXFLAGS) -MMD -MP -c $< -o $@

/Users/wanghongyan01/.xdev-cache/02/02f36467ce728e55.o: /Users/wanghongyan01/go/src/github.com/hongyanwang/xuperchain/core/contractsdk/cpp/example/anchor/src/anchor.pb.cc
	@mkdir -p $(dir $@)
	@echo CC $(notdir $<)
	@$(CXX) $(CXXFLAGS) -MMD -MP -c $< -o $@

/Users/wanghongyan01/.xdev-cache/8e/8e62b7b7201a8fa2.o: /Users/wanghongyan01/go/src/github.com/hongyanwang/xuperchain/core/contractsdk/cpp/src/xchain/account.cc
	@mkdir -p $(dir $@)
	@echo CC $(notdir $<)
	@$(CXX) $(CXXFLAGS) -MMD -MP -c $< -o $@

/Users/wanghongyan01/.xdev-cache/60/60e1be4113e4a52a.o: /Users/wanghongyan01/go/src/github.com/hongyanwang/xuperchain/core/contractsdk/cpp/src/xchain/basic_iterator.cc
	@mkdir -p $(dir $@)
	@echo CC $(notdir $<)
	@$(CXX) $(CXXFLAGS) -MMD -MP -c $< -o $@

/Users/wanghongyan01/.xdev-cache/d9/d9bf63b2d01d0ba0.o: /Users/wanghongyan01/go/src/github.com/hongyanwang/xuperchain/core/contractsdk/cpp/src/xchain/block.cc
	@mkdir -p $(dir $@)
	@echo CC $(notdir $<)
	@$(CXX) $(CXXFLAGS) -MMD -MP -c $< -o $@

/Users/wanghongyan01/.xdev-cache/0d/0dd484f0a7177b6f.o: /Users/wanghongyan01/go/src/github.com/hongyanwang/xuperchain/core/contractsdk/cpp/src/xchain/context_impl.cc
	@mkdir -p $(dir $@)
	@echo CC $(notdir $<)
	@$(CXX) $(CXXFLAGS) -MMD -MP -c $< -o $@

/Users/wanghongyan01/.xdev-cache/6a/6a41febc3c2171eb.o: /Users/wanghongyan01/go/src/github.com/hongyanwang/xuperchain/core/contractsdk/cpp/src/xchain/contract.cc
	@mkdir -p $(dir $@)
	@echo CC $(notdir $<)
	@$(CXX) $(CXXFLAGS) -MMD -MP -c $< -o $@

/Users/wanghongyan01/.xdev-cache/99/993596d029c6fc99.o: /Users/wanghongyan01/go/src/github.com/hongyanwang/xuperchain/core/contractsdk/cpp/src/xchain/contract.pb.cc
	@mkdir -p $(dir $@)
	@echo CC $(notdir $<)
	@$(CXX) $(CXXFLAGS) -MMD -MP -c $< -o $@

/Users/wanghongyan01/.xdev-cache/f8/f83b6fb1ed4a3776.o: /Users/wanghongyan01/go/src/github.com/hongyanwang/xuperchain/core/contractsdk/cpp/src/xchain/crypto.cc
	@mkdir -p $(dir $@)
	@echo CC $(notdir $<)
	@$(CXX) $(CXXFLAGS) -MMD -MP -c $< -o $@

/Users/wanghongyan01/.xdev-cache/7a/7a52010d0d7081fc.o: /Users/wanghongyan01/go/src/github.com/hongyanwang/xuperchain/core/contractsdk/cpp/src/xchain/syscall.cc
	@mkdir -p $(dir $@)
	@echo CC $(notdir $<)
	@$(CXX) $(CXXFLAGS) -MMD -MP -c $< -o $@

/Users/wanghongyan01/.xdev-cache/6c/6cf4e2c45fffa443.o: /Users/wanghongyan01/go/src/github.com/hongyanwang/xuperchain/core/contractsdk/cpp/src/xchain/transaction.cc
	@mkdir -p $(dir $@)
	@echo CC $(notdir $<)
	@$(CXX) $(CXXFLAGS) -MMD -MP -c $< -o $@

/Users/wanghongyan01/go/src/github.com/hongyanwang/xuperchain/core/contractsdk/cpp/example/anchor/anchor.wasm: /Users/wanghongyan01/.xdev-cache/be/be14f170a1b67ef7.o /Users/wanghongyan01/.xdev-cache/02/02f36467ce728e55.o /Users/wanghongyan01/.xdev-cache/8e/8e62b7b7201a8fa2.o /Users/wanghongyan01/.xdev-cache/60/60e1be4113e4a52a.o /Users/wanghongyan01/.xdev-cache/d9/d9bf63b2d01d0ba0.o /Users/wanghongyan01/.xdev-cache/0d/0dd484f0a7177b6f.o /Users/wanghongyan01/.xdev-cache/6a/6a41febc3c2171eb.o /Users/wanghongyan01/.xdev-cache/99/993596d029c6fc99.o /Users/wanghongyan01/.xdev-cache/f8/f83b6fb1ed4a3776.o /Users/wanghongyan01/.xdev-cache/7a/7a52010d0d7081fc.o /Users/wanghongyan01/.xdev-cache/6c/6cf4e2c45fffa443.o
	@echo LD wasm
	@$(CXX) -o $@ $^ $(LDFLAGS)

build: /Users/wanghongyan01/go/src/github.com/hongyanwang/xuperchain/core/contractsdk/cpp/example/anchor/anchor.wasm


-include /Users/wanghongyan01/.xdev-cache/be/be14f170a1b67ef7.d /Users/wanghongyan01/.xdev-cache/02/02f36467ce728e55.d /Users/wanghongyan01/.xdev-cache/8e/8e62b7b7201a8fa2.d /Users/wanghongyan01/.xdev-cache/60/60e1be4113e4a52a.d /Users/wanghongyan01/.xdev-cache/d9/d9bf63b2d01d0ba0.d /Users/wanghongyan01/.xdev-cache/0d/0dd484f0a7177b6f.d /Users/wanghongyan01/.xdev-cache/6a/6a41febc3c2171eb.d /Users/wanghongyan01/.xdev-cache/99/993596d029c6fc99.d /Users/wanghongyan01/.xdev-cache/f8/f83b6fb1ed4a3776.d /Users/wanghongyan01/.xdev-cache/7a/7a52010d0d7081fc.d /Users/wanghongyan01/.xdev-cache/6c/6cf4e2c45fffa443.d
