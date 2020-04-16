CXXFLAGS ?= -std=c++11 -Os -I/usr/local/include -Isrc -Werror=vla -I/Users/wanghongyan01/go/src/github.com/test/xuperchain/core/contractsdk/cpp/src
LDFLAGS ?= -Oz -s TOTAL_STACK=256KB -s TOTAL_MEMORY=1MB -s DETERMINISTIC=1 -s EXTRA_EXPORTED_RUNTIME_METHODS=["stackAlloc"] -L/usr/local/lib -lprotobuf-lite -lpthread --js-library /Users/wanghongyan01/go/src/github.com/test/xuperchain/core/contractsdk/cpp/src/xchain/exports.js

.PHONY: all build clean

all: build

clean: 
	$(RM) -r build

/Users/wanghongyan01/.xdev-cache/88/889dde1846787a00.o: /Users/wanghongyan01/go/src/github.com/test/xuperchain/core/contractsdk/cpp/example/cross_query_demo/src/main.cc
	@mkdir -p $(dir $@)
	@echo CC $(notdir $<)
	@$(CXX) $(CXXFLAGS) -MMD -MP -c $< -o $@

/Users/wanghongyan01/.xdev-cache/fb/fb0ff4c62cacdbbd.o: /Users/wanghongyan01/go/src/github.com/test/xuperchain/core/contractsdk/cpp/src/xchain/account.cc
	@mkdir -p $(dir $@)
	@echo CC $(notdir $<)
	@$(CXX) $(CXXFLAGS) -MMD -MP -c $< -o $@

/Users/wanghongyan01/.xdev-cache/d7/d7d71b5014a3183f.o: /Users/wanghongyan01/go/src/github.com/test/xuperchain/core/contractsdk/cpp/src/xchain/basic_iterator.cc
	@mkdir -p $(dir $@)
	@echo CC $(notdir $<)
	@$(CXX) $(CXXFLAGS) -MMD -MP -c $< -o $@

/Users/wanghongyan01/.xdev-cache/26/26fa28f5e44beadf.o: /Users/wanghongyan01/go/src/github.com/test/xuperchain/core/contractsdk/cpp/src/xchain/block.cc
	@mkdir -p $(dir $@)
	@echo CC $(notdir $<)
	@$(CXX) $(CXXFLAGS) -MMD -MP -c $< -o $@

/Users/wanghongyan01/.xdev-cache/67/673404f0965445be.o: /Users/wanghongyan01/go/src/github.com/test/xuperchain/core/contractsdk/cpp/src/xchain/context_impl.cc
	@mkdir -p $(dir $@)
	@echo CC $(notdir $<)
	@$(CXX) $(CXXFLAGS) -MMD -MP -c $< -o $@

/Users/wanghongyan01/.xdev-cache/1c/1c2efd740acc918e.o: /Users/wanghongyan01/go/src/github.com/test/xuperchain/core/contractsdk/cpp/src/xchain/contract.cc
	@mkdir -p $(dir $@)
	@echo CC $(notdir $<)
	@$(CXX) $(CXXFLAGS) -MMD -MP -c $< -o $@

/Users/wanghongyan01/.xdev-cache/34/3494b6338b355d16.o: /Users/wanghongyan01/go/src/github.com/test/xuperchain/core/contractsdk/cpp/src/xchain/contract.pb.cc
	@mkdir -p $(dir $@)
	@echo CC $(notdir $<)
	@$(CXX) $(CXXFLAGS) -MMD -MP -c $< -o $@

/Users/wanghongyan01/.xdev-cache/8b/8bccc64c49e4c417.o: /Users/wanghongyan01/go/src/github.com/test/xuperchain/core/contractsdk/cpp/src/xchain/crypto.cc
	@mkdir -p $(dir $@)
	@echo CC $(notdir $<)
	@$(CXX) $(CXXFLAGS) -MMD -MP -c $< -o $@

/Users/wanghongyan01/.xdev-cache/d1/d18173ca425b7f6b.o: /Users/wanghongyan01/go/src/github.com/test/xuperchain/core/contractsdk/cpp/src/xchain/syscall.cc
	@mkdir -p $(dir $@)
	@echo CC $(notdir $<)
	@$(CXX) $(CXXFLAGS) -MMD -MP -c $< -o $@

/Users/wanghongyan01/.xdev-cache/f5/f551d3748645365c.o: /Users/wanghongyan01/go/src/github.com/test/xuperchain/core/contractsdk/cpp/src/xchain/transaction.cc
	@mkdir -p $(dir $@)
	@echo CC $(notdir $<)
	@$(CXX) $(CXXFLAGS) -MMD -MP -c $< -o $@

/Users/wanghongyan01/go/src/github.com/test/xuperchain/core/contractsdk/cpp/example/cross_query_demo/cross_query_demo.wasm: /Users/wanghongyan01/.xdev-cache/88/889dde1846787a00.o /Users/wanghongyan01/.xdev-cache/fb/fb0ff4c62cacdbbd.o /Users/wanghongyan01/.xdev-cache/d7/d7d71b5014a3183f.o /Users/wanghongyan01/.xdev-cache/26/26fa28f5e44beadf.o /Users/wanghongyan01/.xdev-cache/67/673404f0965445be.o /Users/wanghongyan01/.xdev-cache/1c/1c2efd740acc918e.o /Users/wanghongyan01/.xdev-cache/34/3494b6338b355d16.o /Users/wanghongyan01/.xdev-cache/8b/8bccc64c49e4c417.o /Users/wanghongyan01/.xdev-cache/d1/d18173ca425b7f6b.o /Users/wanghongyan01/.xdev-cache/f5/f551d3748645365c.o
	@echo LD wasm
	@$(CXX) -o $@ $^ $(LDFLAGS)

build: /Users/wanghongyan01/go/src/github.com/test/xuperchain/core/contractsdk/cpp/example/cross_query_demo/cross_query_demo.wasm


-include /Users/wanghongyan01/.xdev-cache/88/889dde1846787a00.d /Users/wanghongyan01/.xdev-cache/fb/fb0ff4c62cacdbbd.d /Users/wanghongyan01/.xdev-cache/d7/d7d71b5014a3183f.d /Users/wanghongyan01/.xdev-cache/26/26fa28f5e44beadf.d /Users/wanghongyan01/.xdev-cache/67/673404f0965445be.d /Users/wanghongyan01/.xdev-cache/1c/1c2efd740acc918e.d /Users/wanghongyan01/.xdev-cache/34/3494b6338b355d16.d /Users/wanghongyan01/.xdev-cache/8b/8bccc64c49e4c417.d /Users/wanghongyan01/.xdev-cache/d1/d18173ca425b7f6b.d /Users/wanghongyan01/.xdev-cache/f5/f551d3748645365c.d
