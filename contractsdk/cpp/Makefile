BUILD_DIR ?= build

XCHAIN_SRCS := $(shell ls ./xchain/*.cc)
XCHAIN_OBJS := $(XCHAIN_SRCS:%=$(BUILD_DIR)/%.o)
XCHAIN_DEPS := $(XCHAIN_OBJS:.o=.d)

CONTRACT_SRCS := $(shell ls ./example/*.cc)
CONTRACT_BIN := $(CONTRACT_SRCS:./example/%.cc=$(BUILD_DIR)/%.wasm)

CONTRACT_TB_SRCS := $(shell ls ./table/*.cc)
CONTRACT_TB_OBJS := $(CONTRACT_TB_SRCS:./table/%.cc=$(BUILD_DIR)/table/%.cc.o)
CONTRACT_PB_SRCS := $(shell ls ./pb/*.cc)
CONTRACT_PB_OBJS := $(CONTRACT_PB_SRCS:./pb/%.cc=$(BUILD_DIR)/pb/%.cc.o)
XCHAIN_CDT_LIBS := $(CONTRACT_PB_OBJS) $(CONTRACT_TB_OBJS)

MOCK_CLASS_SRCS := $(shell ls ./test/fake_*.cc)
MOCK_CLASS_OBJS := $(MOCK_CLASS_SRCS:./test/%.cc=$(BUILD_DIR)/test/%.cc.o)

TEST_SRCS := $(shell ls ./test/test_*.cc)
TEST_OBJS := $(TEST_SRCS:./test/%.cc=$(BUILD_DIR)/test/%.cc.o)
TEST_BIN := $(TEST_SRCS:./test/%.cc=$(BUILD_DIR)/test/%.out)

INC_DIRS := . /usr/local/include
INC_FLAGS := $(addprefix -I,$(INC_DIRS))

OPT_LEVEL := -Oz
ifeq ($(CXX),g++)
	OPT_LEVEL := -Os
endif

CPPFLAGS ?= $(INC_FLAGS) -MMD -MP $(OPT_LEVEL) -std=c++11

.PHONY: clean all

all: $(BUILD_DIR)/libxchain.a $(CONTRACT_BIN)

clean:
	$(RM) -r $(BUILD_DIR)

$(BUILD_DIR)/libxchain.a: $(XCHAIN_OBJS)
	$(AR) -rc $@ $^
	$(RANLIB) $@

test: $(TEST_BIN)

$(BUILD_DIR)/test/%.out: $(BUILD_DIR)/test/%.cc.o $(XCHAIN_CDT_LIBS) $(MOCK_CLASS_OBJS) $(BUILD_DIR)/libxchain.a
	$(CXX) $(CPPFLAGS) $(CXXFLAGS) -o $@ $^ -O3 -L/usr/local/lib -lprotobuf-lite -lpthread -lgtest
	$@

# wasm target
$(BUILD_DIR)/%.wasm: $(BUILD_DIR)/example/%.cc.o $(XCHAIN_CDT_LIBS) $(BUILD_DIR)/libxchain.a
	$(CXX) $(CPPFLAGS) $(CXXFLAGS) -o $@ $^ -Oz \
								-s DETERMINISTIC=1 \
								-s TOTAL_STACK=256KB \
								-s TOTAL_MEMORY=1MB \
								-s EXTRA_EXPORTED_RUNTIME_METHODS='["stackAlloc"]' \
								--js-library xchain/exports.js \
								-L/usr/local/lib -lprotobuf-lite -lpthread

# c++ source
$(BUILD_DIR)/%.cc.o: %.cc
	$(MKDIR_P) $(dir $@)
	$(CXX) $(CPPFLAGS) $(CXXFLAGS) -c $< -o $@

-include $(XCHAIN_DEPS)

MKDIR_P ?= mkdir -p
RANLIB ?= ranlib
