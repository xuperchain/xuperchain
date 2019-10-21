CXXFLAGS ?= -std=c++14 -O0 -I/usr/local/include -Isrc -Werror=vla -I/Users/zhengqi/Documents/work/blockchain/xuperunion/contractsdk/cpp
LDFLAGS ?= -L/Users/zhengqi/Documents/work/blockchain/xuperunion/contractsdk/cpp/build -lxchain -Oz -s ERROR_ON_UNDEFINED_SYMBOLS=0 -s DETERMINISTIC=1 -L/usr/local/lib -lprotobuf-lite -lpthread

.PHONY: all clean

all: xrc01.wasm

clean: 
	$(RM) -r build

build/%.cc.o: %.cc
	@mkdir -p $(dir $@)
	@echo CC $<
	@$(CXX) $(CXXFLAGS) -MMD -MP -c $< -o $@

build/libmain.a: build/src/main/xrc01.cc.o build/src/main/xrc01_e1.cc.o
	@$(AR) -rc $@ $^
	@$(RANLIB) $@

build/libpb.a: build/src/pb/xrc01.pb.cc.o
	@$(AR) -rc $@ $^
	@$(RANLIB) $@

build/libtable.a: 
	@$(AR) -rc $@ $^
	@$(RANLIB) $@

build/xrc01.wasm: build/libmain.a build/libpb.a build/libtable.a
	@echo LD xrc01.wasm
	@$(CXX) -o $@ build/src/main/xrc01.cc.o build/src/main/xrc01_e1.cc.o build/src/pb/xrc01.pb.cc.o $(LDFLAGS)

xrc01.wasm: build/xrc01.wasm


-include build/src/main/xrc01.cc.d build/src/main/xrc01_e1.cc.d build/src/pb/xrc01.pb.cc.d
