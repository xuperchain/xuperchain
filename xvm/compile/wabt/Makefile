BUILD_DIR ?= build
SRC_DIRS ?= src

SRCS := $(shell ls ./src/*.cc ./src/*.c)
OBJS := $(SRCS:%=$(BUILD_DIR)/%.o)
DEPS := $(OBJS:.o=.d)

INC_DIRS := . src
INC_FLAGS := $(addprefix -I,$(INC_DIRS))

CPPFLAGS ?= $(INC_FLAGS) -MMD -MP -std=c++11
CFLAGS ?= $(INC_FLAGS) -MMD -MP

.PHONY: clean all

all: $(BUILD_DIR)/wasm2c $(BUILD_DIR)/wat2wasm $(BUILD_DIR)/wast2json

clean:
	$(RM) -r $(BUILD_DIR)

$(BUILD_DIR)/libwabt.a: $(OBJS)
	ar rc $@ $^
	ranlib $@

$(BUILD_DIR)/wasm2c: $(BUILD_DIR)/src/tools/wasm2c.cc.o $(BUILD_DIR)/libwabt.a 
	$(CXX) -o $@ $^

$(BUILD_DIR)/wat2wasm: $(BUILD_DIR)/src/tools/wat2wasm.cc.o $(BUILD_DIR)/libwabt.a 
	$(CXX) -o $@ $^

$(BUILD_DIR)/wast2json: $(BUILD_DIR)/src/tools/wast2json.cc.o $(BUILD_DIR)/libwabt.a 
	$(CXX) -o $@ $^

# c source
$(BUILD_DIR)/%.c.o: %.c
	$(MKDIR_P) $(dir $@)
	$(CC) $(CFLAGS) $(CFLAGS) -c $< -o $@

# c++ source
$(BUILD_DIR)/%.cc.o: %.cc
	$(MKDIR_P) $(dir $@)
	$(CXX) $(CPPFLAGS) $(CXXFLAGS) -c $< -o $@

-include $(DEPS)

MKDIR_P ?= mkdir -p
