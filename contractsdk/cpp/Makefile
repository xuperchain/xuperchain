BUILD_DIR ?= build

XCHAIN_SRCS := $(shell ls ./xchain/*.cc)
XCHAIN_OBJS := $(XCHAIN_SRCS:%=$(BUILD_DIR)/%.o)
XCHAIN_DEPS := $(XCHAIN_OBJS:.o=.d)

CONTRACT_SRCS := $(shell ls ./example/*.cc)
CONTRACT_BIN := $(CONTRACT_SRCS:./example/%.cc=$(BUILD_DIR)/%.wasm)

INC_DIRS := . /usr/local/include
INC_FLAGS := $(addprefix -I,$(INC_DIRS))

CPPFLAGS ?= $(INC_FLAGS) -MMD -MP -Oz -std=c++11

.PHONY: clean all

all: $(BUILD_DIR)/libxchain.a $(CONTRACT_BIN)

clean:
	$(RM) -r $(BUILD_DIR)

$(BUILD_DIR)/libxchain.a: $(XCHAIN_OBJS)
	$(AR) rc $@ $^
	$(RANLIB) $@

# wasm target
$(BUILD_DIR)/%.wasm: $(BUILD_DIR)/example/%.cc.o $(BUILD_DIR)/libxchain.a
	$(CXX) $(CPPFLAGS) $(CXXFLAGS) -o $@ $^ -Oz -s ERROR_ON_UNDEFINED_SYMBOLS=0 -s DETERMINISTIC=1 -L/usr/local/lib -lprotobuf-lite -lpthread

# c++ source
$(BUILD_DIR)/%.cc.o: %.cc
	$(MKDIR_P) $(dir $@)
	$(CXX) $(CPPFLAGS) $(CXXFLAGS) -c $< -o $@

-include $(XCHAIN_DEPS)

MKDIR_P ?= mkdir -p
