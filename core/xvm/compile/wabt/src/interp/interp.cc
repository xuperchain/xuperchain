/*
 * Copyright 2016 WebAssembly Community Group participants
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

#include "src/interp/interp.h"

#include <algorithm>
#include <cassert>
#include <cinttypes>
#include <cmath>
#include <limits>
#include <type_traits>
#include <vector>

#include "src/interp/interp-internal.h"

#include "src/cast.h"
#include "src/stream.h"

namespace wabt {
namespace interp {

// Differs from the normal CHECK_RESULT because this one is meant to return the
// interp Result type.
#undef CHECK_RESULT
#define CHECK_RESULT(expr)   \
  do {                       \
    if (WABT_FAILED(expr)) { \
      return Result::Error;  \
    }                        \
  } while (0)

// Differs from CHECK_RESULT since it can return different traps, not just
// Error. Also uses __VA_ARGS__ so templates can be passed without surrounding
// parentheses.
#define CHECK_TRAP(...)            \
  do {                             \
    Result result = (__VA_ARGS__); \
    if (result != Result::Ok) {    \
      return result;               \
    }                              \
  } while (0)

std::string TypedValueToString(const TypedValue& tv) {
  switch (tv.type) {
    case Type::I32:
      return StringPrintf("i32:%u", tv.get_i32());

    case Type::I64:
      return StringPrintf("i64:%" PRIu64, tv.get_i64());

    case Type::F32: {
      return StringPrintf("f32:%f", tv.get_f32());
    }

    case Type::F64: {
      return StringPrintf("f64:%f", tv.get_f64());
    }

    case Type::V128:
      return StringPrintf("v128 i32x4:0x%08x 0x%08x 0x%08x 0x%08x",
                          tv.value.v128_bits.v[0], tv.value.v128_bits.v[1],
                          tv.value.v128_bits.v[2], tv.value.v128_bits.v[3]);

    default:
      WABT_UNREACHABLE;
  }
}

void WriteTypedValue(Stream* stream, const TypedValue& tv) {
  std::string s = TypedValueToString(tv);
  stream->WriteData(s.data(), s.size());
}

void WriteTypedValues(Stream* stream, const TypedValues& values) {
  for (size_t i = 0; i < values.size(); ++i) {
    WriteTypedValue(stream, values[i]);
    if (i != values.size() - 1) {
      stream->Writef(", ");
    }
  }
}

#define V(name, str) str,
static const char* s_trap_strings[] = {FOREACH_INTERP_RESULT(V)};
#undef V

const char* ResultToString(Result result) {
  return s_trap_strings[static_cast<size_t>(result)];
}

void WriteResult(Stream* stream, const char* desc, Result result) {
  stream->Writef("%s: %s\n", desc, ResultToString(result));
}

void WriteCall(Stream* stream,
               string_view module_name,
               string_view func_name,
               const TypedValues& args,
               const TypedValues& results,
               Result result) {
  if (!module_name.empty()) {
    stream->Writef(PRIstringview ".", WABT_PRINTF_STRING_VIEW_ARG(module_name));
  }
  stream->Writef(PRIstringview "(", WABT_PRINTF_STRING_VIEW_ARG(func_name));
  WriteTypedValues(stream, args);
  stream->Writef(") =>");
  if (result == Result::Ok) {
    if (results.size() > 0) {
      stream->Writef(" ");
      WriteTypedValues(stream, results);
    }
    stream->Writef("\n");
  } else {
    WriteResult(stream, " error", result);
  }
}

Environment::Environment() : istream_(new OutputBuffer()) {}

Index Environment::FindModuleIndex(string_view name) const {
  auto iter = module_bindings_.find(name.to_string());
  if (iter == module_bindings_.end()) {
    return kInvalidIndex;
  }
  return iter->second.index;
}

Module* Environment::FindModule(string_view name) {
  Index index = FindModuleIndex(name);
  return index == kInvalidIndex ? nullptr : modules_[index].get();
}

Module* Environment::FindRegisteredModule(string_view name) {
  auto iter = registered_module_bindings_.find(name.to_string());
  if (iter == registered_module_bindings_.end()) {
    return nullptr;
  }
  return modules_[iter->second.index].get();
}

Thread::Options::Options(uint32_t value_stack_size, uint32_t call_stack_size)
    : value_stack_size(value_stack_size), call_stack_size(call_stack_size) {}

Thread::Thread(Environment* env, const Options& options)
    : env_(env),
      value_stack_(options.value_stack_size),
      call_stack_(options.call_stack_size) {}

FuncSignature::FuncSignature(std::vector<Type> param_types,
                             std::vector<Type> result_types)
    : param_types(param_types), result_types(result_types) {}

FuncSignature::FuncSignature(Index param_count,
                             Type* param_types,
                             Index result_count,
                             Type* result_types)
    : param_types(param_types, param_types + param_count),
      result_types(result_types, result_types + result_count) {}

Module::Module(bool is_host)
    : memory_index(kInvalidIndex),
      table_index(kInvalidIndex),
      is_host(is_host) {}

Module::Module(string_view name, bool is_host)
    : name(name.to_string()),
      memory_index(kInvalidIndex),
      table_index(kInvalidIndex),
      is_host(is_host) {}

Export* Module::GetFuncExport(Environment* env,
                              string_view name,
                              Index sig_index) {
  auto range = export_bindings.equal_range(name.to_string());
  for (auto iter = range.first; iter != range.second; ++iter) {
    const Binding& binding = iter->second;
    Export* export_ = &exports[binding.index];
    if (export_->kind == ExternalKind::Func) {
      const Func* func = env->GetFunc(export_->index);
      if (env->FuncSignaturesAreEqual(sig_index, func->sig_index)) {
        return export_;
      }
    }
  }

  // No match; check whether the module wants to spontaneously create a
  // function of this name and signature.
  Index index = OnUnknownFuncExport(name, sig_index);
  if (index != kInvalidIndex) {
    Export* export_ = &exports[index];
    assert(export_->kind == ExternalKind::Func);
    const Func* func = env->GetFunc(export_->index);
    WABT_USE(func);
    assert(env->FuncSignaturesAreEqual(sig_index, func->sig_index));
    return export_;
  }

  return nullptr;
}

Export* Module::GetExport(string_view name) {
  int field_index = export_bindings.FindIndex(name);
  if (field_index < 0) {
    return nullptr;
  }
  return &exports[field_index];
}

Index Module::AppendExport(ExternalKind kind,
                           Index item_index,
                           string_view name) {
  exports.emplace_back(name, kind, item_index);
  Export* export_ = &exports.back();
  export_bindings.emplace(export_->name, Binding(exports.size() - 1));
  return exports.size() - 1;
}

DefinedModule::DefinedModule()
    : Module(false),
      start_func_index(kInvalidIndex),
      istream_start(kInvalidIstreamOffset),
      istream_end(kInvalidIstreamOffset) {}

HostModule::HostModule(Environment* env, string_view name)
    : Module(name, true), env_(env) {}

Index HostModule::OnUnknownFuncExport(string_view name, Index sig_index) {
  if (on_unknown_func_export) {
    return on_unknown_func_export(env_, this, name, sig_index);
  }
  return kInvalidIndex;
}

std::pair<HostFunc*, Index> HostModule::AppendFuncExport(
    string_view name,
    const FuncSignature& sig,
    HostFunc::Callback callback) {
  // TODO(binji): dedupe signature?
  env_->EmplaceBackFuncSignature(sig);
  Index sig_index = env_->GetFuncSignatureCount() - 1;
  return AppendFuncExport(name, sig_index, callback);
}

std::pair<HostFunc*, Index> HostModule::AppendFuncExport(
    string_view name,
    Index sig_index,
    HostFunc::Callback callback) {
  auto* host_func = new HostFunc(this->name, name, sig_index, callback);
  env_->EmplaceBackFunc(host_func);
  Index func_env_index = env_->GetFuncCount() - 1;
  Index export_index = AppendExport(ExternalKind::Func, func_env_index, name);
  return {host_func, export_index};
}

std::pair<Table*, Index> HostModule::AppendTableExport(string_view name,
                                                       Type elem_type,
                                                       const Limits& limits) {
  Table* table = env_->EmplaceBackTable(elem_type, limits);
  Index table_env_index = env_->GetTableCount() - 1;
  Index export_index = AppendExport(ExternalKind::Table, table_env_index, name);
  return {table, export_index};
}

std::pair<Memory*, Index> HostModule::AppendMemoryExport(string_view name,
                                                         const Limits& limits) {
  Memory* memory = env_->EmplaceBackMemory(limits);
  Index memory_env_index = env_->GetMemoryCount() - 1;
  Index export_index =
      AppendExport(ExternalKind::Memory, memory_env_index, name);
  return {memory, export_index};
}

std::pair<Global*, Index> HostModule::AppendGlobalExport(string_view name,
                                                         Type type,
                                                         bool mutable_) {
  Global* global = env_->EmplaceBackGlobal(TypedValue(type), mutable_);
  Index global_env_index = env_->GetGlobalCount() - 1;
  Index export_index =
      AppendExport(ExternalKind::Global, global_env_index, name);
  return {global, export_index};
}

std::pair<Global*, Index> HostModule::AppendGlobalExport(string_view name,
                                             bool mutable_,
                                             uint32_t value) {
  std::pair<Global*, Index> pair =
      AppendGlobalExport(name, Type::I32, mutable_);
  pair.first->typed_value.set_i32(value);
  return pair;
}

std::pair<Global*, Index> HostModule::AppendGlobalExport(string_view name,
                                             bool mutable_,
                                             uint64_t value) {
  std::pair<Global*, Index> pair =
      AppendGlobalExport(name, Type::I64, mutable_);
  pair.first->typed_value.set_i64(value);
  return pair;
}

std::pair<Global*, Index> HostModule::AppendGlobalExport(string_view name,
                                             bool mutable_,
                                             float value) {
  std::pair<Global*, Index> pair =
      AppendGlobalExport(name, Type::F32, mutable_);
  pair.first->typed_value.set_f32(value);
  return pair;
}

std::pair<Global*, Index> HostModule::AppendGlobalExport(string_view name,
                                             bool mutable_,
                                             double value) {
  std::pair<Global*, Index> pair =
      AppendGlobalExport(name, Type::F64, mutable_);
  pair.first->typed_value.set_f64(value);
  return pair;
}

Environment::MarkPoint Environment::Mark() {
  MarkPoint mark;
  mark.modules_size = modules_.size();
  mark.sigs_size = sigs_.size();
  mark.funcs_size = funcs_.size();
  mark.memories_size = memories_.size();
  mark.tables_size = tables_.size();
  mark.globals_size = globals_.size();
  mark.istream_size = istream_->data.size();
  return mark;
}

void Environment::ResetToMarkPoint(const MarkPoint& mark) {
  // Destroy entries in the binding hash.
  for (size_t i = mark.modules_size; i < modules_.size(); ++i) {
    std::string name = modules_[i]->name;
    if (!name.empty()) {
      module_bindings_.erase(name);
    }
  }

  // registered_module_bindings_ maps from an arbitrary name to a module index,
  // so we have to iterate through the entire table to find entries to remove.
  auto iter = registered_module_bindings_.begin();
  while (iter != registered_module_bindings_.end()) {
    if (iter->second.index >= mark.modules_size) {
      iter = registered_module_bindings_.erase(iter);
    } else {
      ++iter;
    }
  }

  modules_.erase(modules_.begin() + mark.modules_size, modules_.end());
  sigs_.erase(sigs_.begin() + mark.sigs_size, sigs_.end());
  funcs_.erase(funcs_.begin() + mark.funcs_size, funcs_.end());
  memories_.erase(memories_.begin() + mark.memories_size, memories_.end());
  tables_.erase(tables_.begin() + mark.tables_size, tables_.end());
  globals_.erase(globals_.begin() + mark.globals_size, globals_.end());
  istream_->data.resize(mark.istream_size);
}

HostModule* Environment::AppendHostModule(string_view name) {
  HostModule* module = new HostModule(this, name);
  modules_.emplace_back(module);
  registered_module_bindings_.emplace(name.to_string(),
                                      Binding(modules_.size() - 1));
  return module;
}

uint32_t ToRep(bool x) { return x ? 1 : 0; }
uint32_t ToRep(uint32_t x) { return x; }
uint64_t ToRep(uint64_t x) { return x; }
uint32_t ToRep(int32_t x) { return Bitcast<uint32_t>(x); }
uint64_t ToRep(int64_t x) { return Bitcast<uint64_t>(x); }
uint32_t ToRep(float x) { return Bitcast<uint32_t>(x); }
uint64_t ToRep(double x) { return Bitcast<uint64_t>(x); }
v128     ToRep(v128 x) { return Bitcast<v128>(x); }

template <typename Dst, typename Src>
Dst FromRep(Src x);

template <>
uint32_t FromRep<uint32_t>(uint32_t x) { return x; }
template <>
uint64_t FromRep<uint64_t>(uint64_t x) { return x; }
template <>
int32_t FromRep<int32_t>(uint32_t x) { return Bitcast<int32_t>(x); }
template <>
int64_t FromRep<int64_t>(uint64_t x) { return Bitcast<int64_t>(x); }
template <>
float FromRep<float>(uint32_t x) { return Bitcast<float>(x); }
template <>
double FromRep<double>(uint64_t x) { return Bitcast<double>(x); }
template <>
v128 FromRep<v128>(v128 x) { return Bitcast<v128>(x); }

template <typename T>
struct FloatTraits;

template <typename R, typename T>
bool IsConversionInRange(ValueTypeRep<T> bits);

/* 3 32222222 222...00
 * 1 09876543 210...10
 * -------------------
 * 0 00000000 000...00 => 0x00000000 => 0
 * 0 10011101 111...11 => 0x4effffff => 2147483520                  (~INT32_MAX)
 * 0 10011110 000...00 => 0x4f000000 => 2147483648
 * 0 10011110 111...11 => 0x4f7fffff => 4294967040                 (~UINT32_MAX)
 * 0 10111110 111...11 => 0x5effffff => 9223371487098961920         (~INT64_MAX)
 * 0 10111110 000...00 => 0x5f000000 => 9223372036854775808
 * 0 10111111 111...11 => 0x5f7fffff => 18446742974197923840       (~UINT64_MAX)
 * 0 10111111 000...00 => 0x5f800000 => 18446744073709551616
 * 0 11111111 000...00 => 0x7f800000 => inf
 * 0 11111111 000...01 => 0x7f800001 => nan(0x1)
 * 0 11111111 111...11 => 0x7fffffff => nan(0x7fffff)
 * 1 00000000 000...00 => 0x80000000 => -0
 * 1 01111110 111...11 => 0xbf7fffff => -1 + ulp      (~UINT32_MIN, ~UINT64_MIN)
 * 1 01111111 000...00 => 0xbf800000 => -1
 * 1 10011110 000...00 => 0xcf000000 => -2147483648                  (INT32_MIN)
 * 1 10111110 000...00 => 0xdf000000 => -9223372036854775808         (INT64_MIN)
 * 1 11111111 000...00 => 0xff800000 => -inf
 * 1 11111111 000...01 => 0xff800001 => -nan(0x1)
 * 1 11111111 111...11 => 0xffffffff => -nan(0x7fffff)
 */

template <>
struct FloatTraits<float> {
  static const uint32_t kMax = 0x7f7fffffU;
  static const uint32_t kInf = 0x7f800000U;
  static const uint32_t kNegMax = 0xff7fffffU;
  static const uint32_t kNegInf = 0xff800000U;
  static const uint32_t kNegOne = 0xbf800000U;
  static const uint32_t kNegZero = 0x80000000U;
  static const uint32_t kQuietNan = 0x7fc00000U;
  static const uint32_t kQuietNegNan = 0xffc00000U;
  static const int kSigBits = 23;
  static const uint32_t kSigMask = 0x7fffff;
  static const uint32_t kSignMask = 0x80000000U;

  static bool IsNan(uint32_t bits) {
    return (bits > kInf && bits < kNegZero) || (bits > kNegInf);
  }

  static bool IsZero(uint32_t bits) { return bits == 0 || bits == kNegZero; }

  static bool IsCanonicalNan(uint32_t bits) {
    return bits == kQuietNan || bits == kQuietNegNan;
  }

  static bool IsArithmeticNan(uint32_t bits) {
    return (bits & kQuietNan) == kQuietNan;
  }

  static uint32_t CanonicalizeNan(uint32_t bits) {
    return WABT_UNLIKELY(IsNan(bits)) ? kQuietNan : bits;
  }
};

bool IsCanonicalNan(uint32_t bits) {
  return FloatTraits<float>::IsCanonicalNan(bits);
}

bool IsArithmeticNan(uint32_t bits) {
  return FloatTraits<float>::IsArithmeticNan(bits);
}

template <>
bool IsConversionInRange<int32_t, float>(uint32_t bits) {
  return (bits < 0x4f000000U) ||
         (bits >= FloatTraits<float>::kNegZero && bits <= 0xcf000000U);
}

template <>
bool IsConversionInRange<int64_t, float>(uint32_t bits) {
  return (bits < 0x5f000000U) ||
         (bits >= FloatTraits<float>::kNegZero && bits <= 0xdf000000U);
}

template <>
bool IsConversionInRange<uint32_t, float>(uint32_t bits) {
  return (bits < 0x4f800000U) || (bits >= FloatTraits<float>::kNegZero &&
                                  bits < FloatTraits<float>::kNegOne);
}

template <>
bool IsConversionInRange<uint64_t, float>(uint32_t bits) {
  return (bits < 0x5f800000U) || (bits >= FloatTraits<float>::kNegZero &&
                                  bits < FloatTraits<float>::kNegOne);
}

/*
 * 6 66655555555 5544..2..222221...000
 * 3 21098765432 1098..9..432109...210
 * -----------------------------------
 * 0 00000000000 0000..0..000000...000 0x0000000000000000 => 0
 * 0 10000011101 1111..1..111000...000 0x41dfffffffc00000 => 2147483647           (INT32_MAX)
 * 0 10000011110 1111..1..111100...000 0x41efffffffe00000 => 4294967295           (UINT32_MAX)
 * 0 10000111101 1111..1..111111...111 0x43dfffffffffffff => 9223372036854774784  (~INT64_MAX)
 * 0 10000111110 0000..0..000000...000 0x43e0000000000000 => 9223372036854775808
 * 0 10000111110 1111..1..111111...111 0x43efffffffffffff => 18446744073709549568 (~UINT64_MAX)
 * 0 10000111111 0000..0..000000...000 0x43f0000000000000 => 18446744073709551616
 * 0 10001111110 1111..1..000000...000 0x47efffffe0000000 => 3.402823e+38         (FLT_MAX)
 * 0 11111111111 0000..0..000000...000 0x7ff0000000000000 => inf
 * 0 11111111111 0000..0..000000...001 0x7ff0000000000001 => nan(0x1)
 * 0 11111111111 1111..1..111111...111 0x7fffffffffffffff => nan(0xfff...)
 * 1 00000000000 0000..0..000000...000 0x8000000000000000 => -0
 * 1 01111111110 1111..1..111111...111 0xbfefffffffffffff => -1 + ulp             (~UINT32_MIN, ~UINT64_MIN)
 * 1 01111111111 0000..0..000000...000 0xbff0000000000000 => -1
 * 1 10000011110 0000..0..000000...000 0xc1e0000000000000 => -2147483648          (INT32_MIN)
 * 1 10000111110 0000..0..000000...000 0xc3e0000000000000 => -9223372036854775808 (INT64_MIN)
 * 1 10001111110 1111..1..000000...000 0xc7efffffe0000000 => -3.402823e+38        (-FLT_MAX)
 * 1 11111111111 0000..0..000000...000 0xfff0000000000000 => -inf
 * 1 11111111111 0000..0..000000...001 0xfff0000000000001 => -nan(0x1)
 * 1 11111111111 1111..1..111111...111 0xffffffffffffffff => -nan(0xfff...)
 */

template <>
struct FloatTraits<double> {
  static const uint64_t kInf = 0x7ff0000000000000ULL;
  static const uint64_t kNegInf = 0xfff0000000000000ULL;
  static const uint64_t kNegOne = 0xbff0000000000000ULL;
  static const uint64_t kNegZero = 0x8000000000000000ULL;
  static const uint64_t kQuietNan = 0x7ff8000000000000ULL;
  static const uint64_t kQuietNegNan = 0xfff8000000000000ULL;
  static const int kSigBits = 52;
  static const uint64_t kSigMask = 0xfffffffffffffULL;
  static const uint64_t kSignMask = 0x8000000000000000ULL;

  static bool IsNan(uint64_t bits) {
    return (bits > kInf && bits < kNegZero) || (bits > kNegInf);
  }

  static bool IsZero(uint64_t bits) { return bits == 0 || bits == kNegZero; }

  static bool IsCanonicalNan(uint64_t bits) {
    return bits == kQuietNan || bits == kQuietNegNan;
  }

  static bool IsArithmeticNan(uint64_t bits) {
    return (bits & kQuietNan) == kQuietNan;
  }

  static uint64_t CanonicalizeNan(uint64_t bits) {
    return WABT_UNLIKELY(IsNan(bits)) ? kQuietNan : bits;
  }
};

bool IsCanonicalNan(uint64_t bits) {
  return FloatTraits<double>::IsCanonicalNan(bits);
}

bool IsArithmeticNan(uint64_t bits) {
  return FloatTraits<double>::IsArithmeticNan(bits);
}

template <>
bool IsConversionInRange<int32_t, double>(uint64_t bits) {
  return (bits <= 0x41dfffffffc00000ULL) ||
         (bits >= FloatTraits<double>::kNegZero &&
          bits <= 0xc1e0000000000000ULL);
}

template <>
bool IsConversionInRange<int64_t, double>(uint64_t bits) {
  return (bits < 0x43e0000000000000ULL) ||
         (bits >= FloatTraits<double>::kNegZero &&
          bits <= 0xc3e0000000000000ULL);
}

template <>
bool IsConversionInRange<uint32_t, double>(uint64_t bits) {
  return (bits <= 0x41efffffffe00000ULL) ||
         (bits >= FloatTraits<double>::kNegZero &&
          bits < FloatTraits<double>::kNegOne);
}

template <>
bool IsConversionInRange<uint64_t, double>(uint64_t bits) {
  return (bits < 0x43f0000000000000ULL) ||
         (bits >= FloatTraits<double>::kNegZero &&
          bits < FloatTraits<double>::kNegOne);
}

template <>
bool IsConversionInRange<float, double>(uint64_t bits) {
  return (bits <= 0x47efffffe0000000ULL) ||
         (bits >= FloatTraits<double>::kNegZero &&
          bits <= 0xc7efffffe0000000ULL);
}

// The WebAssembly rounding mode means that these values (which are > F32_MAX)
// should be rounded to F32_MAX and not set to infinity. Unfortunately, UBSAN
// complains that the value is not representable as a float, so we'll special
// case them.
bool IsInRangeF64DemoteF32RoundToF32Max(uint64_t bits) {
  return bits > 0x47efffffe0000000ULL && bits < 0x47effffff0000000ULL;
}

bool IsInRangeF64DemoteF32RoundToNegF32Max(uint64_t bits) {
  return bits > 0xc7efffffe0000000ULL && bits < 0xc7effffff0000000ULL;
}

template <typename T, typename MemType> struct ExtendMemType;
template<> struct ExtendMemType<uint32_t, uint8_t> { typedef uint32_t type; };
template<> struct ExtendMemType<uint32_t, int8_t> { typedef int32_t type; };
template<> struct ExtendMemType<uint32_t, uint16_t> { typedef uint32_t type; };
template<> struct ExtendMemType<uint32_t, int16_t> { typedef int32_t type; };
template<> struct ExtendMemType<uint32_t, uint32_t> { typedef uint32_t type; };
template<> struct ExtendMemType<uint32_t, int32_t> { typedef int32_t type; };
template<> struct ExtendMemType<uint64_t, uint8_t> { typedef uint64_t type; };
template<> struct ExtendMemType<uint64_t, int8_t> { typedef int64_t type; };
template<> struct ExtendMemType<uint64_t, uint16_t> { typedef uint64_t type; };
template<> struct ExtendMemType<uint64_t, int16_t> { typedef int64_t type; };
template<> struct ExtendMemType<uint64_t, uint32_t> { typedef uint64_t type; };
template<> struct ExtendMemType<uint64_t, int32_t> { typedef int64_t type; };
template<> struct ExtendMemType<uint64_t, uint64_t> { typedef uint64_t type; };
template<> struct ExtendMemType<uint64_t, int64_t> { typedef int64_t type; };
template<> struct ExtendMemType<float, float> { typedef float type; };
template<> struct ExtendMemType<double, double> { typedef double type; };
template<> struct ExtendMemType<v128, v128> { typedef v128 type; };

template <typename T, typename MemType> struct WrapMemType;
template<> struct WrapMemType<uint32_t, uint8_t> { typedef uint8_t type; };
template<> struct WrapMemType<uint32_t, uint16_t> { typedef uint16_t type; };
template<> struct WrapMemType<uint32_t, uint32_t> { typedef uint32_t type; };
template<> struct WrapMemType<uint64_t, uint8_t> { typedef uint8_t type; };
template<> struct WrapMemType<uint64_t, uint16_t> { typedef uint16_t type; };
template<> struct WrapMemType<uint64_t, uint32_t> { typedef uint32_t type; };
template<> struct WrapMemType<uint64_t, uint64_t> { typedef uint64_t type; };
template<> struct WrapMemType<float, float> { typedef uint32_t type; };
template<> struct WrapMemType<double, double> { typedef uint64_t type; };
template<> struct WrapMemType<v128, v128> { typedef v128 type; };

template <typename T>
Value MakeValue(ValueTypeRep<T>);

template <>
Value MakeValue<uint32_t>(uint32_t v) {
  Value result;
  result.i32 = v;
  return result;
}

template <>
Value MakeValue<int32_t>(uint32_t v) {
  Value result;
  result.i32 = v;
  return result;
}

template <>
Value MakeValue<uint64_t>(uint64_t v) {
  Value result;
  result.i64 = v;
  return result;
}

template <>
Value MakeValue<int64_t>(uint64_t v) {
  Value result;
  result.i64 = v;
  return result;
}

template <>
Value MakeValue<float>(uint32_t v) {
  Value result;
  result.f32_bits = v;
  return result;
}

template <>
Value MakeValue<double>(uint64_t v) {
  Value result;
  result.f64_bits = v;
  return result;
}

template <>
Value MakeValue<v128>(v128 v) {
  Value result;
  result.v128_bits = v;
  return result;
}

template <typename T> ValueTypeRep<T> GetValue(Value);
template<> uint32_t GetValue<int32_t>(Value v) { return v.i32; }
template<> uint32_t GetValue<uint32_t>(Value v) { return v.i32; }
template<> uint64_t GetValue<int64_t>(Value v) { return v.i64; }
template<> uint64_t GetValue<uint64_t>(Value v) { return v.i64; }
template<> uint32_t GetValue<float>(Value v) { return v.f32_bits; }
template<> uint64_t GetValue<double>(Value v) { return v.f64_bits; }
template<> v128 GetValue<v128>(Value v) { return v.v128_bits; }

template <typename T>
ValueTypeRep<T> CanonicalizeNan(ValueTypeRep<T> rep) {
  return rep;
}

template <>
ValueTypeRep<float> CanonicalizeNan<float>(ValueTypeRep<float> rep) {
  return FloatTraits<float>::CanonicalizeNan(rep);
}

template <>
ValueTypeRep<double> CanonicalizeNan<double>(ValueTypeRep<double> rep) {
  return FloatTraits<double>::CanonicalizeNan(rep);
}

#define TRAP(type) return Result::Trap##type
#define TRAP_UNLESS(cond, type) TRAP_IF(!(cond), type)
#define TRAP_IF(cond, type)    \
  do {                         \
    if (WABT_UNLIKELY(cond)) { \
      TRAP(type);              \
    }                          \
  } while (0)

#define CHECK_STACK() \
  TRAP_IF(value_stack_top_ >= value_stack_.size(), ValueStackExhausted)

#define PUSH_NEG_1_AND_BREAK_IF(cond) \
  if (WABT_UNLIKELY(cond)) {          \
    CHECK_TRAP(Push<int32_t>(-1));    \
    break;                            \
  }

#define GOTO(offset) pc = &istream[offset]

Memory* Thread::ReadMemory(const uint8_t** pc) {
  Index memory_index = ReadU32(pc);
  return &env_->memories_[memory_index];
}

template <typename MemType>
Result Thread::GetAccessAddress(const uint8_t** pc, void** out_address) {
  Memory* memory = ReadMemory(pc);
  uint64_t addr = static_cast<uint64_t>(Pop<uint32_t>()) + ReadU32(pc);
  TRAP_IF(addr + sizeof(MemType) > memory->data.size(),
          MemoryAccessOutOfBounds);
  *out_address = memory->data.data() + static_cast<IstreamOffset>(addr);
  return Result::Ok;
}

template <typename MemType>
Result Thread::GetAtomicAccessAddress(const uint8_t** pc, void** out_address) {
  Memory* memory = ReadMemory(pc);
  uint64_t addr = static_cast<uint64_t>(Pop<uint32_t>()) + ReadU32(pc);
  TRAP_IF(addr + sizeof(MemType) > memory->data.size(),
          MemoryAccessOutOfBounds);
  TRAP_IF((addr & (sizeof(MemType) - 1)) != 0, AtomicMemoryAccessUnaligned);
  *out_address = memory->data.data() + static_cast<IstreamOffset>(addr);
  return Result::Ok;
}

Value& Thread::Top() {
  return Pick(1);
}

Value& Thread::Pick(Index depth) {
  return value_stack_[value_stack_top_ - depth];
}

void Thread::Reset() {
  pc_ = 0;
  value_stack_top_ = 0;
  call_stack_top_ = 0;
}

Result Thread::Push(Value value) {
  CHECK_STACK();
  value_stack_[value_stack_top_++] = value;
  return Result::Ok;
}

Value Thread::Pop() {
  return value_stack_[--value_stack_top_];
}

Value Thread::ValueAt(Index at) const {
  assert(at < value_stack_top_);
  return value_stack_[at];
}

template <typename T>
Result Thread::Push(T value) {
  return PushRep<T>(ToRep(value));
}

template <typename T>
T Thread::Pop() {
  return FromRep<T>(PopRep<T>());
}

template <typename T>
Result Thread::PushRep(ValueTypeRep<T> value) {
  return Push(MakeValue<T>(value));
}

template <typename T>
ValueTypeRep<T> Thread::PopRep() {
  return GetValue<T>(Pop());
}

void Thread::DropKeep(uint32_t drop_count, uint32_t keep_count) {
  // Copy backward to avoid clobbering when the regions overlap.
  for (uint32_t i = keep_count; i > 0; --i) {
    Pick(drop_count + i) = Pick(i);
  }
  value_stack_top_ -= drop_count;
}

Result Thread::PushCall(const uint8_t* pc) {
  TRAP_IF(call_stack_top_ >= call_stack_.size(), CallStackExhausted);
  call_stack_[call_stack_top_++] = pc - GetIstream();
  return Result::Ok;
}

IstreamOffset Thread::PopCall() {
  return call_stack_[--call_stack_top_];
}

template <typename T>
void LoadFromMemory(T* dst, const void* src) {
  memcpy(dst, src, sizeof(T));
}

template <typename T>
void StoreToMemory(void* dst, T value) {
  memcpy(dst, &value, sizeof(T));
}

template <typename MemType, typename ResultType>
Result Thread::Load(const uint8_t** pc) {
  typedef typename ExtendMemType<ResultType, MemType>::type ExtendedType;
  static_assert(std::is_floating_point<MemType>::value ==
                    std::is_floating_point<ExtendedType>::value,
                "Extended type should be float iff MemType is float");

  void* src;
  CHECK_TRAP(GetAccessAddress<MemType>(pc, &src));
  MemType value;
  LoadFromMemory<MemType>(&value, src);
  return Push<ResultType>(static_cast<ExtendedType>(value));
}

template <typename MemType, typename ResultType>
Result Thread::Store(const uint8_t** pc) {
  typedef typename WrapMemType<ResultType, MemType>::type WrappedType;
  WrappedType value = PopRep<ResultType>();
  void* dst;
  CHECK_TRAP(GetAccessAddress<MemType>(pc, &dst));
  StoreToMemory<WrappedType>(dst, value);
  return Result::Ok;
}

template <typename MemType, typename ResultType>
Result Thread::AtomicLoad(const uint8_t** pc) {
  typedef typename ExtendMemType<ResultType, MemType>::type ExtendedType;
  static_assert(!std::is_floating_point<MemType>::value,
                "AtomicLoad type can't be float");
  void* src;
  CHECK_TRAP(GetAtomicAccessAddress<MemType>(pc, &src));
  MemType value;
  LoadFromMemory<MemType>(&value, src);
  return Push<ResultType>(static_cast<ExtendedType>(value));
}

template <typename MemType, typename ResultType>
Result Thread::AtomicStore(const uint8_t** pc) {
  typedef typename WrapMemType<ResultType, MemType>::type WrappedType;
  WrappedType value = PopRep<ResultType>();
  void* dst;
  CHECK_TRAP(GetAtomicAccessAddress<MemType>(pc, &dst));
  StoreToMemory<WrappedType>(dst, value);
  return Result::Ok;
}

template <typename MemType, typename ResultType>
Result Thread::AtomicRmw(BinopFunc<ResultType, ResultType> func,
                         const uint8_t** pc) {
  typedef typename ExtendMemType<ResultType, MemType>::type ExtendedType;
  MemType rhs = PopRep<ResultType>();
  void* addr;
  CHECK_TRAP(GetAtomicAccessAddress<MemType>(pc, &addr));
  MemType read;
  LoadFromMemory<MemType>(&read, addr);
  StoreToMemory<MemType>(addr, func(read, rhs));
  return Push<ResultType>(static_cast<ExtendedType>(read));
}

template <typename MemType, typename ResultType>
Result Thread::AtomicRmwCmpxchg(const uint8_t** pc) {
  typedef typename ExtendMemType<ResultType, MemType>::type ExtendedType;
  MemType replace = PopRep<ResultType>();
  MemType expect = PopRep<ResultType>();
  void* addr;
  CHECK_TRAP(GetAtomicAccessAddress<MemType>(pc, &addr));
  MemType read;
  LoadFromMemory<MemType>(&read, addr);
  if (read == expect) {
    StoreToMemory<MemType>(addr, replace);
  }
  return Push<ResultType>(static_cast<ExtendedType>(read));
}

template <typename R, typename T>
Result Thread::Unop(UnopFunc<R, T> func) {
  auto value = PopRep<T>();
  return PushRep<R>(func(value));
}

// {i8, i16, 132, i64}{16, 8, 4, 2}.(neg)
template <typename T, typename L, typename R, typename P>
Result Thread::SimdUnop(UnopFunc<R, P> func) {
  auto value = PopRep<T>();

  // Calculate how many Lanes according to input lane data type.
  constexpr int32_t lanes = sizeof(T) / sizeof(L);

  // Define SIMD data array for Simd add by Lanes.
  L simd_data_ret[lanes];
  L simd_data_0[lanes];

  // Convert intput SIMD data to array.
  memcpy(simd_data_0, &value, sizeof(T));

  // Constuct the Simd value by Lane data and Lane nums.
  for (int32_t i = 0; i < lanes; i++) {
    simd_data_ret[i] = static_cast<L>(func(simd_data_0[i]));
  }

  return PushRep<T>(Bitcast<T>(simd_data_ret));
}

template <typename R, typename T>
Result Thread::UnopTrap(UnopTrapFunc<R, T> func) {
  auto value = PopRep<T>();
  ValueTypeRep<R> result_value;
  CHECK_TRAP(func(value, &result_value));
  return PushRep<R>(result_value);
}

template <typename R, typename T>
Result Thread::Binop(BinopFunc<R, T> func) {
  auto rhs_rep = PopRep<T>();
  auto lhs_rep = PopRep<T>();
  return PushRep<R>(func(lhs_rep, rhs_rep));
}

// {i8, i16, 132, i64}{16, 8, 4, 2}.(add/sub/mul)
template <typename T, typename L, typename R, typename P>
Result Thread::SimdBinop(BinopFunc<R, P> func) {
  auto rhs_rep = PopRep<T>();
  auto lhs_rep = PopRep<T>();

  // Calculate how many Lanes according to input lane data type.
  constexpr int32_t lanes = sizeof(T) / sizeof(L);

  // Define SIMD data array for Simd add by Lanes.
  L simd_data_ret[lanes];
  L simd_data_0[lanes];
  L simd_data_1[lanes];

  // Convert intput SIMD data to array.
  memcpy(simd_data_0, &lhs_rep, sizeof(T));
  memcpy(simd_data_1, &rhs_rep, sizeof(T));

  // Constuct the Simd value by Lane data and Lane nums.
  for (int32_t i = 0; i < lanes; i++) {
    simd_data_ret[i] = static_cast<L>(func(simd_data_0[i], simd_data_1[i]));
  }

  return PushRep<T>(Bitcast<T>(simd_data_ret));
}

// {i8, i16, 132, i64, f32, f64}{16, 8, 4, 2}.(eq/ne/lt/le/gt/ge)
template <typename T, typename L, typename R, typename P>
Result Thread::SimdRelBinop(BinopFunc<R, P> func) {
  auto rhs_rep = PopRep<T>();
  auto lhs_rep = PopRep<T>();

  // Calculate how many Lanes according to input lane data type.
  constexpr int32_t lanes = sizeof(T) / sizeof(L);

  // Define SIMD data array for Simd add by Lanes.
  L simd_data_ret[lanes];
  L simd_data_0[lanes];
  L simd_data_1[lanes];

  // Convert intput SIMD data to array.
  memcpy(simd_data_0, &lhs_rep, sizeof(T));
  memcpy(simd_data_1, &rhs_rep, sizeof(T));

  // Constuct the Simd value by Lane data and Lane nums.
  for (int32_t i = 0; i < lanes; i++) {
    simd_data_ret[i] = static_cast<L>(
      func(simd_data_0[i], simd_data_1[i]) == 0? 0 : -1
    );
  }

  return PushRep<T>(Bitcast<T>(simd_data_ret));
}

template <typename R, typename T>
Result Thread::BinopTrap(BinopTrapFunc<R, T> func) {
  auto rhs_rep = PopRep<T>();
  auto lhs_rep = PopRep<T>();
  ValueTypeRep<R> result_value;
  CHECK_TRAP(func(lhs_rep, rhs_rep, &result_value));
  return PushRep<R>(result_value);
}

// {i,f}{32,64}.add
template <typename T>
ValueTypeRep<T> Add(ValueTypeRep<T> lhs_rep, ValueTypeRep<T> rhs_rep) {
  return CanonicalizeNan<T>(ToRep(FromRep<T>(lhs_rep) + FromRep<T>(rhs_rep)));
}

template <typename T, typename R>
ValueTypeRep<T> AddSaturate(ValueTypeRep<T> lhs_rep, ValueTypeRep<T> rhs_rep) {
  T max = std::numeric_limits<R>::max();
  T min = std::numeric_limits<R>::min();
  T result = static_cast<T>(lhs_rep) + static_cast<T>(rhs_rep);

  if (result < min) {
    return ToRep(min);
  } else if (result > max) {
    return ToRep(max);
  } else {
    return ToRep(result);
  }
}

template <typename T, typename R>
ValueTypeRep<T> SubSaturate(ValueTypeRep<T> lhs_rep, ValueTypeRep<T> rhs_rep) {
  T max = std::numeric_limits<R>::max();
  T min = std::numeric_limits<R>::min();
  T result = static_cast<T>(lhs_rep) - static_cast<T>(rhs_rep);

  if (result < min) {
    return ToRep(min);
  } else if (result > max) {
    return ToRep(max);
  } else {
    return ToRep(result);
  }
}

template <typename T, typename L>
int32_t SimdIsLaneTrue(ValueTypeRep<T> value, int32_t true_cond) {
  int true_count = 0;

  // Calculate how many Lanes according to input lane data type.
  constexpr int32_t lanes = sizeof(T) / sizeof(L);

  // Define SIMD data array for Simd Lanes.
  L simd_data_0[lanes];

  // Convert intput SIMD data to array.
  memcpy(simd_data_0, &value, sizeof(T));

  // Constuct the Simd value by Lane data and Lane nums.
  for (int32_t i = 0; i < lanes; i++) {
    if (simd_data_0[i] != 0)
      true_count++;
  }

  return (true_count >= true_cond) ? 1 : 0;
}

// {i,f}{32,64}.sub
template <typename T>
ValueTypeRep<T> Sub(ValueTypeRep<T> lhs_rep, ValueTypeRep<T> rhs_rep) {
  return CanonicalizeNan<T>(ToRep(FromRep<T>(lhs_rep) - FromRep<T>(rhs_rep)));
}

// {i,f}{32,64}.mul
template <typename T>
ValueTypeRep<T> Mul(ValueTypeRep<T> lhs_rep, ValueTypeRep<T> rhs_rep) {
  return CanonicalizeNan<T>(ToRep(FromRep<T>(lhs_rep) * FromRep<T>(rhs_rep)));
}

// i{32,64}.{div,rem}_s are special-cased because they trap when dividing the
// max signed value by -1. The modulo operation on x86 uses the same
// instruction to generate the quotient and the remainder.
template <typename T>
bool IsNormalDivRemS(T lhs, T rhs) {
  static_assert(std::is_signed<T>::value, "T should be a signed type.");
  return !(lhs == std::numeric_limits<T>::min() && rhs == -1);
}

// i{32,64}.div_s
template <typename T>
Result IntDivS(ValueTypeRep<T> lhs_rep,
               ValueTypeRep<T> rhs_rep,
               ValueTypeRep<T>* out_result) {
  auto lhs = FromRep<T>(lhs_rep);
  auto rhs = FromRep<T>(rhs_rep);
  TRAP_IF(rhs == 0, IntegerDivideByZero);
  TRAP_UNLESS(IsNormalDivRemS(lhs, rhs), IntegerOverflow);
  *out_result = ToRep(lhs / rhs);
  return Result::Ok;
}

// i{32,64}.rem_s
template <typename T>
Result IntRemS(ValueTypeRep<T> lhs_rep,
               ValueTypeRep<T> rhs_rep,
               ValueTypeRep<T>* out_result) {
  auto lhs = FromRep<T>(lhs_rep);
  auto rhs = FromRep<T>(rhs_rep);
  TRAP_IF(rhs == 0, IntegerDivideByZero);
  if (WABT_LIKELY(IsNormalDivRemS(lhs, rhs))) {
    *out_result = ToRep(lhs % rhs);
  } else {
    *out_result = 0;
  }
  return Result::Ok;
}

// i{32,64}.div_u
template <typename T>
Result IntDivU(ValueTypeRep<T> lhs_rep,
               ValueTypeRep<T> rhs_rep,
               ValueTypeRep<T>* out_result) {
  auto lhs = FromRep<T>(lhs_rep);
  auto rhs = FromRep<T>(rhs_rep);
  TRAP_IF(rhs == 0, IntegerDivideByZero);
  *out_result = ToRep(lhs / rhs);
  return Result::Ok;
}

// i{32,64}.rem_u
template <typename T>
Result IntRemU(ValueTypeRep<T> lhs_rep,
               ValueTypeRep<T> rhs_rep,
               ValueTypeRep<T>* out_result) {
  auto lhs = FromRep<T>(lhs_rep);
  auto rhs = FromRep<T>(rhs_rep);
  TRAP_IF(rhs == 0, IntegerDivideByZero);
  *out_result = ToRep(lhs % rhs);
  return Result::Ok;
}

// f{32,64}.div
template <typename T>
ValueTypeRep<T> FloatDiv(ValueTypeRep<T> lhs_rep, ValueTypeRep<T> rhs_rep) {
  typedef FloatTraits<T> Traits;
  ValueTypeRep<T> result;
  if (WABT_UNLIKELY(Traits::IsZero(rhs_rep))) {
    if (Traits::IsNan(lhs_rep) || Traits::IsZero(lhs_rep)) {
      result = Traits::kQuietNan;
    } else {
      auto sign = (lhs_rep & Traits::kSignMask) ^ (rhs_rep & Traits::kSignMask);
      result = sign | Traits::kInf;
    }
  } else {
    result =
        CanonicalizeNan<T>(ToRep(FromRep<T>(lhs_rep) / FromRep<T>(rhs_rep)));
  }
  return result;
}

// i{32,64}.and
template <typename T>
ValueTypeRep<T> IntAnd(ValueTypeRep<T> lhs_rep, ValueTypeRep<T> rhs_rep) {
  return ToRep(FromRep<T>(lhs_rep) & FromRep<T>(rhs_rep));
}

// i{32,64}.or
template <typename T>
ValueTypeRep<T> IntOr(ValueTypeRep<T> lhs_rep, ValueTypeRep<T> rhs_rep) {
  return ToRep(FromRep<T>(lhs_rep) | FromRep<T>(rhs_rep));
}

// i{32,64}.xor
template <typename T>
ValueTypeRep<T> IntXor(ValueTypeRep<T> lhs_rep, ValueTypeRep<T> rhs_rep) {
  return ToRep(FromRep<T>(lhs_rep) ^ FromRep<T>(rhs_rep));
}

// i{32,64}.shl
template <typename T>
ValueTypeRep<T> IntShl(ValueTypeRep<T> lhs_rep, ValueTypeRep<T> rhs_rep) {
  const int mask = sizeof(T) * 8 - 1;
  return ToRep(FromRep<T>(lhs_rep) << (FromRep<T>(rhs_rep) & mask));
}

// i{32,64}.shr_{s,u}
template <typename T>
ValueTypeRep<T> IntShr(ValueTypeRep<T> lhs_rep, ValueTypeRep<T> rhs_rep) {
  const int mask = sizeof(T) * 8 - 1;
  return ToRep(FromRep<T>(lhs_rep) >> (FromRep<T>(rhs_rep) & mask));
}

// i{32,64}.rotl
template <typename T>
ValueTypeRep<T> IntRotl(ValueTypeRep<T> lhs_rep, ValueTypeRep<T> rhs_rep) {
  const int mask = sizeof(T) * 8 - 1;
  int amount = FromRep<T>(rhs_rep) & mask;
  auto lhs = FromRep<T>(lhs_rep);
  if (amount == 0) {
    return ToRep(lhs);
  } else {
    return ToRep((lhs << amount) | (lhs >> (mask + 1 - amount)));
  }
}

// i{32,64}.rotr
template <typename T>
ValueTypeRep<T> IntRotr(ValueTypeRep<T> lhs_rep, ValueTypeRep<T> rhs_rep) {
  const int mask = sizeof(T) * 8 - 1;
  int amount = FromRep<T>(rhs_rep) & mask;
  auto lhs = FromRep<T>(lhs_rep);
  if (amount == 0) {
    return ToRep(lhs);
  } else {
    return ToRep((lhs >> amount) | (lhs << (mask + 1 - amount)));
  }
}

// i{32,64}.eqz
template <typename R, typename T>
ValueTypeRep<R> IntEqz(ValueTypeRep<T> v_rep) {
  return ToRep(v_rep == 0);
}

template <typename T>
ValueTypeRep<T> IntNeg(ValueTypeRep<T> v_rep) {
  T tmp = static_cast<T>(v_rep);
  return ToRep(-tmp);
}

template <typename T>
ValueTypeRep<T> IntNot(ValueTypeRep<T> v_rep) {
  T tmp = static_cast<T>(v_rep);
  return ToRep(~tmp);
}

// f{32,64}.abs
template <typename T>
ValueTypeRep<T> FloatAbs(ValueTypeRep<T> v_rep) {
  return v_rep & ~FloatTraits<T>::kSignMask;
}

// f{32,64}.neg
template <typename T>
ValueTypeRep<T> FloatNeg(ValueTypeRep<T> v_rep) {
  return v_rep ^ FloatTraits<T>::kSignMask;
}

// f{32,64}.ceil
template <typename T>
ValueTypeRep<T> FloatCeil(ValueTypeRep<T> v_rep) {
  return CanonicalizeNan<T>(ToRep(std::ceil(FromRep<T>(v_rep))));
}

// f{32,64}.floor
template <typename T>
ValueTypeRep<T> FloatFloor(ValueTypeRep<T> v_rep) {
  return CanonicalizeNan<T>(ToRep(std::floor(FromRep<T>(v_rep))));
}

// f{32,64}.trunc
template <typename T>
ValueTypeRep<T> FloatTrunc(ValueTypeRep<T> v_rep) {
  return CanonicalizeNan<T>(ToRep(std::trunc(FromRep<T>(v_rep))));
}

// f{32,64}.nearest
template <typename T>
ValueTypeRep<T> FloatNearest(ValueTypeRep<T> v_rep) {
  return CanonicalizeNan<T>(ToRep(std::nearbyint(FromRep<T>(v_rep))));
}

// f{32,64}.sqrt
template <typename T>
ValueTypeRep<T> FloatSqrt(ValueTypeRep<T> v_rep) {
  return CanonicalizeNan<T>(ToRep(std::sqrt(FromRep<T>(v_rep))));
}

// f{32,64}.min
template <typename T>
ValueTypeRep<T> FloatMin(ValueTypeRep<T> lhs_rep, ValueTypeRep<T> rhs_rep) {
  typedef FloatTraits<T> Traits;

  if (WABT_UNLIKELY(Traits::IsNan(lhs_rep) || Traits::IsNan(rhs_rep))) {
    return Traits::kQuietNan;
  } else if (WABT_UNLIKELY(Traits::IsZero(lhs_rep) &&
                           Traits::IsZero(rhs_rep))) {
    // min(0.0, -0.0) == -0.0, but std::min won't produce the correct result.
    // We can instead compare using the unsigned integer representation, but
    // just max instead (since the sign bit makes the value larger).
    return std::max(lhs_rep, rhs_rep);
  } else {
    return ToRep(std::min(FromRep<T>(lhs_rep), FromRep<T>(rhs_rep)));
  }
}

// f{32,64}.max
template <typename T>
ValueTypeRep<T> FloatMax(ValueTypeRep<T> lhs_rep, ValueTypeRep<T> rhs_rep) {
  typedef FloatTraits<T> Traits;

  if (WABT_UNLIKELY(Traits::IsNan(lhs_rep) || Traits::IsNan(rhs_rep))) {
    return Traits::kQuietNan;
  } else if (WABT_UNLIKELY(Traits::IsZero(lhs_rep) &&
                           Traits::IsZero(rhs_rep))) {
    // min(0.0, -0.0) == -0.0, but std::min won't produce the correct result.
    // We can instead compare using the unsigned integer representation, but
    // just max instead (since the sign bit makes the value larger).
    return std::min(lhs_rep, rhs_rep);
  } else {
    return ToRep(std::max(FromRep<T>(lhs_rep), FromRep<T>(rhs_rep)));
  }
}

// f{32,64}.copysign
template <typename T>
ValueTypeRep<T> FloatCopySign(ValueTypeRep<T> lhs_rep,
                              ValueTypeRep<T> rhs_rep) {
  typedef FloatTraits<T> Traits;
  return (lhs_rep & ~Traits::kSignMask) | (rhs_rep & Traits::kSignMask);
}

// {i,f}{32,64}.eq
template <typename T>
uint32_t Eq(ValueTypeRep<T> lhs_rep, ValueTypeRep<T> rhs_rep) {
  return ToRep(FromRep<T>(lhs_rep) == FromRep<T>(rhs_rep));
}

// {i,f}{32,64}.ne
template <typename T>
uint32_t Ne(ValueTypeRep<T> lhs_rep, ValueTypeRep<T> rhs_rep) {
  return ToRep(FromRep<T>(lhs_rep) != FromRep<T>(rhs_rep));
}

// f{32,64}.lt | i{32,64}.lt_{s,u}
template <typename T>
uint32_t Lt(ValueTypeRep<T> lhs_rep, ValueTypeRep<T> rhs_rep) {
  return ToRep(FromRep<T>(lhs_rep) < FromRep<T>(rhs_rep));
}

// f{32,64}.le | i{32,64}.le_{s,u}
template <typename T>
uint32_t Le(ValueTypeRep<T> lhs_rep, ValueTypeRep<T> rhs_rep) {
  return ToRep(FromRep<T>(lhs_rep) <= FromRep<T>(rhs_rep));
}

// f{32,64}.gt | i{32,64}.gt_{s,u}
template <typename T>
uint32_t Gt(ValueTypeRep<T> lhs_rep, ValueTypeRep<T> rhs_rep) {
  return ToRep(FromRep<T>(lhs_rep) > FromRep<T>(rhs_rep));
}

// f{32,64}.ge | i{32,64}.ge_{s,u}
template <typename T>
uint32_t Ge(ValueTypeRep<T> lhs_rep, ValueTypeRep<T> rhs_rep) {
  return ToRep(FromRep<T>(lhs_rep) >= FromRep<T>(rhs_rep));
}

// f32x4.convert_{s,u}/i32x4 and f64x2.convert_s/i64x2.
template <typename R, typename T>
ValueTypeRep<R> SimdConvert(ValueTypeRep<T> v_rep) {
  return ToRep(static_cast<R>(static_cast<T>(v_rep)));
}

// f64x2.convert_u/i64x2 use this instance due to MSVC issue.
template <>
ValueTypeRep<double> SimdConvert<double, uint64_t>(
    ValueTypeRep<uint64_t> v_rep) {
  return ToRep(wabt_convert_uint64_to_double(v_rep));
}

// i{32,64}.trunc_{s,u}/f{32,64}
template <typename R, typename T>
Result IntTrunc(ValueTypeRep<T> v_rep, ValueTypeRep<R>* out_result) {
  TRAP_IF(FloatTraits<T>::IsNan(v_rep), InvalidConversionToInteger);
  TRAP_UNLESS((IsConversionInRange<R, T>(v_rep)), IntegerOverflow);
  *out_result = ToRep(static_cast<R>(FromRep<T>(v_rep)));
  return Result::Ok;
}

// i{32,64}.trunc_{s,u}:sat/f{32,64}
template <typename R, typename T>
ValueTypeRep<R> IntTruncSat(ValueTypeRep<T> v_rep) {
  typedef FloatTraits<T> Traits;
  if (WABT_UNLIKELY(Traits::IsNan(v_rep))) {
    return 0;
  } else if (WABT_UNLIKELY((!IsConversionInRange<R, T>(v_rep)))) {
    if (v_rep & Traits::kSignMask) {
      return ToRep(std::numeric_limits<R>::min());
    } else {
      return ToRep(std::numeric_limits<R>::max());
    }
  } else {
    return ToRep(static_cast<R>(FromRep<T>(v_rep)));
  }
}

// i{32,64}.extend{8,16,32}_s
template <typename T, typename E>
ValueTypeRep<T> IntExtendS(ValueTypeRep<T> v_rep) {
  // To avoid undefined/implementation-defined behavior, convert from unsigned
  // type (T), to an unsigned value of the smaller size (EU), then bitcast from
  // unsigned to signed, then cast from the smaller signed type to the larger
  // signed type (TS) to sign extend. ToRep then will bitcast back from signed
  // to unsigned.
  static_assert(std::is_unsigned<ValueTypeRep<T>>::value, "T must be unsigned");
  static_assert(std::is_signed<E>::value, "E must be signed");
  typedef typename std::make_unsigned<E>::type EU;
  typedef typename std::make_signed<T>::type TS;
  return ToRep(static_cast<TS>(Bitcast<E>(static_cast<EU>(v_rep))));
}

// i{32,64}.atomic.rmw(8,16,32}_u.xchg
template <typename T>
ValueTypeRep<T> Xchg(ValueTypeRep<T> lhs_rep, ValueTypeRep<T> rhs_rep) {
  return rhs_rep;
}

// i(8,16,32,64) f(32,64) X (2,4,8,16) splat ==> v128
template <typename T, typename V>
ValueTypeRep<T> SimdSplat(V lane_data) {
  // Calculate how many Lanes according to input lane data type.
  int32_t lanes = sizeof(T) / sizeof(V);

  // Define SIMD data array by Lanes.
  V simd_data[sizeof(T) / sizeof(V)];

  // Constuct the Simd value by Land data and Lane nums.
  for (int32_t i = 0; i < lanes; i++) {
    simd_data[i] = lane_data;
  }

  return ToRep(Bitcast<T>(simd_data));
}

// Simd instructions of Lane extract.
// value: input v128 value.
// typename T: lane data type.
template <typename R, typename V, typename T>
ValueTypeRep<R> SimdExtractLane(V value, uint32_t laneidx) {
  // Calculate how many Lanes according to input lane data type.
  constexpr int32_t lanes = sizeof(V) / sizeof(T);

  // Define SIMD data array for Simd add by Lanes.
  T simd_data_0[lanes];

  // Convert intput SIMD data to array.
  memcpy(simd_data_0, &value, sizeof(V));

  return ToRep(static_cast<R>(simd_data_0[laneidx]));
}

// Simd instructions of Lane replace.
// value: input v128 value.  lane_val: input lane data.
// typename T: lane data type.
template <typename R, typename V, typename T>
ValueTypeRep<R> SimdReplaceLane(V value, uint32_t lane_idx, T lane_val) {
  // Calculate how many Lanes according to input lane data type.
  constexpr int32_t lanes = sizeof(V) / sizeof(T);

  // Define SIMD data array for Simd add by Lanes.
  T simd_data_0[lanes];

  // Convert intput SIMD data to array.
  memcpy(simd_data_0, &value, sizeof(V));

  // Replace the indicated lane.
  simd_data_0[lane_idx] = lane_val;

  return ToRep(Bitcast<R>(simd_data_0));
}

bool Environment::FuncSignaturesAreEqual(Index sig_index_0,
                                         Index sig_index_1) const {
  if (sig_index_0 == sig_index_1) {
    return true;
  }
  const FuncSignature* sig_0 = &sigs_[sig_index_0];
  const FuncSignature* sig_1 = &sigs_[sig_index_1];
  return sig_0->param_types == sig_1->param_types &&
         sig_0->result_types == sig_1->result_types;
}

Result Thread::CallHost(HostFunc* func) {
  FuncSignature* sig = &env_->sigs_[func->sig_index];

  size_t num_params = sig->param_types.size();
  size_t num_results = sig->result_types.size();
  TypedValues params(num_params);
  TypedValues results(num_results);

  for (size_t i = num_params; i > 0; --i) {
    params[i - 1].value = Pop();
    params[i - 1].type = sig->param_types[i - 1];
  }

  for (size_t i = 0; i < num_results; ++i) {
    results[i].type = sig->result_types[i];
    results[i].SetZero();
  }

  Result call_result = func->callback(func, sig, params, results);
  TRAP_IF(call_result != Result::Ok, HostTrapped);

  TRAP_IF(results.size() != num_results, HostResultTypeMismatch);
  for (size_t i = 0; i < num_results; ++i) {
    TRAP_IF(results[i].type != sig->result_types[i], HostResultTypeMismatch);
    CHECK_TRAP(Push(results[i].value));
  }

  return Result::Ok;
}

Result Thread::Run(int num_instructions) {
  Result result = Result::Ok;

  const uint8_t* istream = GetIstream();
  const uint8_t* pc = &istream[pc_];
  for (int i = 0; i < num_instructions; ++i) {
    Opcode opcode = ReadOpcode(&pc);
    assert(!opcode.IsInvalid());
    switch (opcode) {
      case Opcode::Select: {
        uint32_t cond = Pop<uint32_t>();
        Value false_ = Pop();
        Value true_ = Pop();
        CHECK_TRAP(Push(cond ? true_ : false_));
        break;
      }

      case Opcode::Br:
        GOTO(ReadU32(&pc));
        break;

      case Opcode::BrIf: {
        IstreamOffset new_pc = ReadU32(&pc);
        if (Pop<uint32_t>()) {
          GOTO(new_pc);
        }
        break;
      }

      case Opcode::BrTable: {
        Index num_targets = ReadU32(&pc);
        IstreamOffset table_offset = ReadU32(&pc);
        uint32_t key = Pop<uint32_t>();
        IstreamOffset key_offset =
            (key >= num_targets ? num_targets : key) * WABT_TABLE_ENTRY_SIZE;
        const uint8_t* entry = istream + table_offset + key_offset;
        IstreamOffset new_pc;
        uint32_t drop_count;
        uint32_t keep_count;
        ReadTableEntryAt(entry, &new_pc, &drop_count, &keep_count);
        DropKeep(drop_count, keep_count);
        GOTO(new_pc);
        break;
      }

      case Opcode::Return:
        if (call_stack_top_ == 0) {
          result = Result::Returned;
          goto exit_loop;
        }
        GOTO(PopCall());
        break;

      case Opcode::Unreachable:
        TRAP(Unreachable);
        break;

      case Opcode::I32Const:
        CHECK_TRAP(Push<uint32_t>(ReadU32(&pc)));
        break;

      case Opcode::I64Const:
        CHECK_TRAP(Push<uint64_t>(ReadU64(&pc)));
        break;

      case Opcode::F32Const:
        CHECK_TRAP(PushRep<float>(ReadU32(&pc)));
        break;

      case Opcode::F64Const:
        CHECK_TRAP(PushRep<double>(ReadU64(&pc)));
        break;

      case Opcode::GlobalGet: {
        Index index = ReadU32(&pc);
        assert(index < env_->globals_.size());
        CHECK_TRAP(Push(env_->globals_[index].typed_value.value));
        break;
      }

      case Opcode::GlobalSet: {
        Index index = ReadU32(&pc);
        assert(index < env_->globals_.size());
        env_->globals_[index].typed_value.value = Pop();
        break;
      }

      case Opcode::LocalGet: {
        Value value = Pick(ReadU32(&pc));
        CHECK_TRAP(Push(value));
        break;
      }

      case Opcode::LocalSet: {
        Value value = Pop();
        Pick(ReadU32(&pc)) = value;
        break;
      }

      case Opcode::LocalTee:
        Pick(ReadU32(&pc)) = Top();
        break;

      case Opcode::Call: {
        IstreamOffset offset = ReadU32(&pc);
        CHECK_TRAP(PushCall(pc));
        GOTO(offset);
        break;
      }

      case Opcode::CallIndirect: {
        Index table_index = ReadU32(&pc);
        Table* table = &env_->tables_[table_index];
        Index sig_index = ReadU32(&pc);
        Index entry_index = Pop<uint32_t>();
        TRAP_IF(entry_index >= table->func_indexes.size(), UndefinedTableIndex);
        Index func_index = table->func_indexes[entry_index];
        TRAP_IF(func_index == kInvalidIndex, UninitializedTableElement);
        Func* func = env_->funcs_[func_index].get();
        TRAP_UNLESS(env_->FuncSignaturesAreEqual(func->sig_index, sig_index),
                    IndirectCallSignatureMismatch);
        if (func->is_host) {
          CHECK_TRAP(CallHost(cast<HostFunc>(func)));
        } else {
          CHECK_TRAP(PushCall(pc));
          GOTO(cast<DefinedFunc>(func)->offset);
        }
        break;
      }

      case Opcode::InterpCallHost: {
        Index func_index = ReadU32(&pc);
        CHECK_TRAP(CallHost(cast<HostFunc>(env_->funcs_[func_index].get())));
        break;
      }

      case Opcode::ReturnCall: {
        IstreamOffset offset = ReadU32(&pc);
        GOTO(offset);

        break;
      }

      case Opcode::ReturnCallIndirect:{
        Index table_index = ReadU32(&pc);
        Table* table = &env_->tables_[table_index];
        Index sig_index = ReadU32(&pc);
        Index entry_index = Pop<uint32_t>();
        TRAP_IF(entry_index >= table->func_indexes.size(), UndefinedTableIndex);
        Index func_index = table->func_indexes[entry_index];
        TRAP_IF(func_index == kInvalidIndex, UninitializedTableElement);
        Func* func = env_->funcs_[func_index].get();
        TRAP_UNLESS(env_->FuncSignaturesAreEqual(func->sig_index, sig_index),
                    IndirectCallSignatureMismatch);
        if (func->is_host) { // Emulate a call/return for imported functions
          CHECK_TRAP(CallHost(cast<HostFunc>(func)));
          if (call_stack_top_ == 0) {
            result = Result::Returned;
            goto exit_loop;
          }
          GOTO(PopCall());
        } else {
          GOTO(cast<DefinedFunc>(func)->offset);
        }
        break;
      }

      case Opcode::I32Load8S:
        CHECK_TRAP(Load<int8_t, uint32_t>(&pc));
        break;

      case Opcode::I32Load8U:
        CHECK_TRAP(Load<uint8_t, uint32_t>(&pc));
        break;

      case Opcode::I32Load16S:
        CHECK_TRAP(Load<int16_t, uint32_t>(&pc));
        break;

      case Opcode::I32Load16U:
        CHECK_TRAP(Load<uint16_t, uint32_t>(&pc));
        break;

      case Opcode::I64Load8S:
        CHECK_TRAP(Load<int8_t, uint64_t>(&pc));
        break;

      case Opcode::I64Load8U:
        CHECK_TRAP(Load<uint8_t, uint64_t>(&pc));
        break;

      case Opcode::I64Load16S:
        CHECK_TRAP(Load<int16_t, uint64_t>(&pc));
        break;

      case Opcode::I64Load16U:
        CHECK_TRAP(Load<uint16_t, uint64_t>(&pc));
        break;

      case Opcode::I64Load32S:
        CHECK_TRAP(Load<int32_t, uint64_t>(&pc));
        break;

      case Opcode::I64Load32U:
        CHECK_TRAP(Load<uint32_t, uint64_t>(&pc));
        break;

      case Opcode::I32Load:
        CHECK_TRAP(Load<uint32_t>(&pc));
        break;

      case Opcode::I64Load:
        CHECK_TRAP(Load<uint64_t>(&pc));
        break;

      case Opcode::F32Load:
        CHECK_TRAP(Load<float>(&pc));
        break;

      case Opcode::F64Load:
        CHECK_TRAP(Load<double>(&pc));
        break;

      case Opcode::I32Store8:
        CHECK_TRAP(Store<uint8_t, uint32_t>(&pc));
        break;

      case Opcode::I32Store16:
        CHECK_TRAP(Store<uint16_t, uint32_t>(&pc));
        break;

      case Opcode::I64Store8:
        CHECK_TRAP(Store<uint8_t, uint64_t>(&pc));
        break;

      case Opcode::I64Store16:
        CHECK_TRAP(Store<uint16_t, uint64_t>(&pc));
        break;

      case Opcode::I64Store32:
        CHECK_TRAP(Store<uint32_t, uint64_t>(&pc));
        break;

      case Opcode::I32Store:
        CHECK_TRAP(Store<uint32_t>(&pc));
        break;

      case Opcode::I64Store:
        CHECK_TRAP(Store<uint64_t>(&pc));
        break;

      case Opcode::F32Store:
        CHECK_TRAP(Store<float>(&pc));
        break;

      case Opcode::F64Store:
        CHECK_TRAP(Store<double>(&pc));
        break;

      case Opcode::I32AtomicLoad8U:
        CHECK_TRAP(AtomicLoad<uint8_t, uint32_t>(&pc));
        break;

      case Opcode::I32AtomicLoad16U:
        CHECK_TRAP(AtomicLoad<uint16_t, uint32_t>(&pc));
        break;

      case Opcode::I64AtomicLoad8U:
        CHECK_TRAP(AtomicLoad<uint8_t, uint64_t>(&pc));
        break;

      case Opcode::I64AtomicLoad16U:
        CHECK_TRAP(AtomicLoad<uint16_t, uint64_t>(&pc));
        break;

      case Opcode::I64AtomicLoad32U:
        CHECK_TRAP(AtomicLoad<uint32_t, uint64_t>(&pc));
        break;

      case Opcode::I32AtomicLoad:
        CHECK_TRAP(AtomicLoad<uint32_t>(&pc));
        break;

      case Opcode::I64AtomicLoad:
        CHECK_TRAP(AtomicLoad<uint64_t>(&pc));
        break;

      case Opcode::I32AtomicStore8:
        CHECK_TRAP(AtomicStore<uint8_t, uint32_t>(&pc));
        break;

      case Opcode::I32AtomicStore16:
        CHECK_TRAP(AtomicStore<uint16_t, uint32_t>(&pc));
        break;

      case Opcode::I64AtomicStore8:
        CHECK_TRAP(AtomicStore<uint8_t, uint64_t>(&pc));
        break;

      case Opcode::I64AtomicStore16:
        CHECK_TRAP(AtomicStore<uint16_t, uint64_t>(&pc));
        break;

      case Opcode::I64AtomicStore32:
        CHECK_TRAP(AtomicStore<uint32_t, uint64_t>(&pc));
        break;

      case Opcode::I32AtomicStore:
        CHECK_TRAP(AtomicStore<uint32_t>(&pc));
        break;

      case Opcode::I64AtomicStore:
        CHECK_TRAP(AtomicStore<uint64_t>(&pc));
        break;

#define ATOMIC_RMW(rmwop, func)                                     \
  case Opcode::I32AtomicRmw##rmwop:                                 \
    CHECK_TRAP(AtomicRmw<uint32_t, uint32_t>(func<uint32_t>, &pc)); \
    break;                                                          \
  case Opcode::I64AtomicRmw##rmwop:                                 \
    CHECK_TRAP(AtomicRmw<uint64_t, uint64_t>(func<uint64_t>, &pc)); \
    break;                                                          \
  case Opcode::I32AtomicRmw8##rmwop##U:                             \
    CHECK_TRAP(AtomicRmw<uint8_t, uint32_t>(func<uint32_t>, &pc));  \
    break;                                                          \
  case Opcode::I32AtomicRmw16##rmwop##U:                            \
    CHECK_TRAP(AtomicRmw<uint16_t, uint32_t>(func<uint32_t>, &pc)); \
    break;                                                          \
  case Opcode::I64AtomicRmw8##rmwop##U:                             \
    CHECK_TRAP(AtomicRmw<uint8_t, uint64_t>(func<uint64_t>, &pc));  \
    break;                                                          \
  case Opcode::I64AtomicRmw16##rmwop##U:                            \
    CHECK_TRAP(AtomicRmw<uint16_t, uint64_t>(func<uint64_t>, &pc)); \
    break;                                                          \
  case Opcode::I64AtomicRmw32##rmwop##U:                            \
    CHECK_TRAP(AtomicRmw<uint32_t, uint64_t>(func<uint64_t>, &pc)); \
    break /* no semicolon */

        ATOMIC_RMW(Add, Add);
        ATOMIC_RMW(Sub, Sub);
        ATOMIC_RMW(And, IntAnd);
        ATOMIC_RMW(Or, IntOr);
        ATOMIC_RMW(Xor, IntXor);
        ATOMIC_RMW(Xchg, Xchg);

#undef ATOMIC_RMW

      case Opcode::I32AtomicRmwCmpxchg:
        CHECK_TRAP(AtomicRmwCmpxchg<uint32_t, uint32_t>(&pc));
        break;

      case Opcode::I64AtomicRmwCmpxchg:
        CHECK_TRAP(AtomicRmwCmpxchg<uint64_t, uint64_t>(&pc));
        break;

      case Opcode::I32AtomicRmw8CmpxchgU:
        CHECK_TRAP(AtomicRmwCmpxchg<uint8_t, uint32_t>(&pc));
        break;

      case Opcode::I32AtomicRmw16CmpxchgU:
        CHECK_TRAP(AtomicRmwCmpxchg<uint16_t, uint32_t>(&pc));
        break;

      case Opcode::I64AtomicRmw8CmpxchgU:
        CHECK_TRAP(AtomicRmwCmpxchg<uint8_t, uint64_t>(&pc));
        break;

      case Opcode::I64AtomicRmw16CmpxchgU:
        CHECK_TRAP(AtomicRmwCmpxchg<uint16_t, uint64_t>(&pc));
        break;

      case Opcode::I64AtomicRmw32CmpxchgU:
        CHECK_TRAP(AtomicRmwCmpxchg<uint32_t, uint64_t>(&pc));
        break;

      case Opcode::MemorySize:
        CHECK_TRAP(Push<uint32_t>(ReadMemory(&pc)->page_limits.initial));
        break;

      case Opcode::MemoryGrow: {
        Memory* memory = ReadMemory(&pc);
        uint32_t old_page_size = memory->page_limits.initial;
        uint32_t grow_pages = Pop<uint32_t>();
        uint32_t new_page_size = old_page_size + grow_pages;
        uint32_t max_page_size = memory->page_limits.has_max
                                     ? memory->page_limits.max
                                     : WABT_MAX_PAGES;
        PUSH_NEG_1_AND_BREAK_IF(new_page_size > max_page_size);
        PUSH_NEG_1_AND_BREAK_IF(
            static_cast<uint64_t>(new_page_size) * WABT_PAGE_SIZE > UINT32_MAX);
        memory->data.resize(new_page_size * WABT_PAGE_SIZE);
        memory->page_limits.initial = new_page_size;
        CHECK_TRAP(Push<uint32_t>(old_page_size));
        break;
      }

      case Opcode::I32Add:
        CHECK_TRAP(Binop(Add<uint32_t>));
        break;

      case Opcode::I32Sub:
        CHECK_TRAP(Binop(Sub<uint32_t>));
        break;

      case Opcode::I32Mul:
        CHECK_TRAP(Binop(Mul<uint32_t>));
        break;

      case Opcode::I32DivS:
        CHECK_TRAP(BinopTrap(IntDivS<int32_t>));
        break;

      case Opcode::I32DivU:
        CHECK_TRAP(BinopTrap(IntDivU<uint32_t>));
        break;

      case Opcode::I32RemS:
        CHECK_TRAP(BinopTrap(IntRemS<int32_t>));
        break;

      case Opcode::I32RemU:
        CHECK_TRAP(BinopTrap(IntRemU<uint32_t>));
        break;

      case Opcode::I32And:
        CHECK_TRAP(Binop(IntAnd<uint32_t>));
        break;

      case Opcode::I32Or:
        CHECK_TRAP(Binop(IntOr<uint32_t>));
        break;

      case Opcode::I32Xor:
        CHECK_TRAP(Binop(IntXor<uint32_t>));
        break;

      case Opcode::I32Shl:
        CHECK_TRAP(Binop(IntShl<uint32_t>));
        break;

      case Opcode::I32ShrU:
        CHECK_TRAP(Binop(IntShr<uint32_t>));
        break;

      case Opcode::I32ShrS:
        CHECK_TRAP(Binop(IntShr<int32_t>));
        break;

      case Opcode::I32Eq:
        CHECK_TRAP(Binop(Eq<uint32_t>));
        break;

      case Opcode::I32Ne:
        CHECK_TRAP(Binop(Ne<uint32_t>));
        break;

      case Opcode::I32LtS:
        CHECK_TRAP(Binop(Lt<int32_t>));
        break;

      case Opcode::I32LeS:
        CHECK_TRAP(Binop(Le<int32_t>));
        break;

      case Opcode::I32LtU:
        CHECK_TRAP(Binop(Lt<uint32_t>));
        break;

      case Opcode::I32LeU:
        CHECK_TRAP(Binop(Le<uint32_t>));
        break;

      case Opcode::I32GtS:
        CHECK_TRAP(Binop(Gt<int32_t>));
        break;

      case Opcode::I32GeS:
        CHECK_TRAP(Binop(Ge<int32_t>));
        break;

      case Opcode::I32GtU:
        CHECK_TRAP(Binop(Gt<uint32_t>));
        break;

      case Opcode::I32GeU:
        CHECK_TRAP(Binop(Ge<uint32_t>));
        break;

      case Opcode::I32Clz:
        CHECK_TRAP(Push<uint32_t>(Clz(Pop<uint32_t>())));
        break;

      case Opcode::I32Ctz:
        CHECK_TRAP(Push<uint32_t>(Ctz(Pop<uint32_t>())));
        break;

      case Opcode::I32Popcnt:
        CHECK_TRAP(Push<uint32_t>(Popcount(Pop<uint32_t>())));
        break;

      case Opcode::I32Eqz:
        CHECK_TRAP(Unop(IntEqz<uint32_t, uint32_t>));
        break;

      case Opcode::I64Add:
        CHECK_TRAP(Binop(Add<uint64_t>));
        break;

      case Opcode::I64Sub:
        CHECK_TRAP(Binop(Sub<uint64_t>));
        break;

      case Opcode::I64Mul:
        CHECK_TRAP(Binop(Mul<uint64_t>));
        break;

      case Opcode::I64DivS:
        CHECK_TRAP(BinopTrap(IntDivS<int64_t>));
        break;

      case Opcode::I64DivU:
        CHECK_TRAP(BinopTrap(IntDivU<uint64_t>));
        break;

      case Opcode::I64RemS:
        CHECK_TRAP(BinopTrap(IntRemS<int64_t>));
        break;

      case Opcode::I64RemU:
        CHECK_TRAP(BinopTrap(IntRemU<uint64_t>));
        break;

      case Opcode::I64And:
        CHECK_TRAP(Binop(IntAnd<uint64_t>));
        break;

      case Opcode::I64Or:
        CHECK_TRAP(Binop(IntOr<uint64_t>));
        break;

      case Opcode::I64Xor:
        CHECK_TRAP(Binop(IntXor<uint64_t>));
        break;

      case Opcode::I64Shl:
        CHECK_TRAP(Binop(IntShl<uint64_t>));
        break;

      case Opcode::I64ShrU:
        CHECK_TRAP(Binop(IntShr<uint64_t>));
        break;

      case Opcode::I64ShrS:
        CHECK_TRAP(Binop(IntShr<int64_t>));
        break;

      case Opcode::I64Eq:
        CHECK_TRAP(Binop(Eq<uint64_t>));
        break;

      case Opcode::I64Ne:
        CHECK_TRAP(Binop(Ne<uint64_t>));
        break;

      case Opcode::I64LtS:
        CHECK_TRAP(Binop(Lt<int64_t>));
        break;

      case Opcode::I64LeS:
        CHECK_TRAP(Binop(Le<int64_t>));
        break;

      case Opcode::I64LtU:
        CHECK_TRAP(Binop(Lt<uint64_t>));
        break;

      case Opcode::I64LeU:
        CHECK_TRAP(Binop(Le<uint64_t>));
        break;

      case Opcode::I64GtS:
        CHECK_TRAP(Binop(Gt<int64_t>));
        break;

      case Opcode::I64GeS:
        CHECK_TRAP(Binop(Ge<int64_t>));
        break;

      case Opcode::I64GtU:
        CHECK_TRAP(Binop(Gt<uint64_t>));
        break;

      case Opcode::I64GeU:
        CHECK_TRAP(Binop(Ge<uint64_t>));
        break;

      case Opcode::I64Clz:
        CHECK_TRAP(Push<uint64_t>(Clz(Pop<uint64_t>())));
        break;

      case Opcode::I64Ctz:
        CHECK_TRAP(Push<uint64_t>(Ctz(Pop<uint64_t>())));
        break;

      case Opcode::I64Popcnt:
        CHECK_TRAP(Push<uint64_t>(Popcount(Pop<uint64_t>())));
        break;

      case Opcode::F32Add:
        CHECK_TRAP(Binop(Add<float>));
        break;

      case Opcode::F32Sub:
        CHECK_TRAP(Binop(Sub<float>));
        break;

      case Opcode::F32Mul:
        CHECK_TRAP(Binop(Mul<float>));
        break;

      case Opcode::F32Div:
        CHECK_TRAP(Binop(FloatDiv<float>));
        break;

      case Opcode::F32Min:
        CHECK_TRAP(Binop(FloatMin<float>));
        break;

      case Opcode::F32Max:
        CHECK_TRAP(Binop(FloatMax<float>));
        break;

      case Opcode::F32Abs:
        CHECK_TRAP(Unop(FloatAbs<float>));
        break;

      case Opcode::F32Neg:
        CHECK_TRAP(Unop(FloatNeg<float>));
        break;

      case Opcode::F32Copysign:
        CHECK_TRAP(Binop(FloatCopySign<float>));
        break;

      case Opcode::F32Ceil:
        CHECK_TRAP(Unop(FloatCeil<float>));
        break;

      case Opcode::F32Floor:
        CHECK_TRAP(Unop(FloatFloor<float>));
        break;

      case Opcode::F32Trunc:
        CHECK_TRAP(Unop(FloatTrunc<float>));
        break;

      case Opcode::F32Nearest:
        CHECK_TRAP(Unop(FloatNearest<float>));
        break;

      case Opcode::F32Sqrt:
        CHECK_TRAP(Unop(FloatSqrt<float>));
        break;

      case Opcode::F32Eq:
        CHECK_TRAP(Binop(Eq<float>));
        break;

      case Opcode::F32Ne:
        CHECK_TRAP(Binop(Ne<float>));
        break;

      case Opcode::F32Lt:
        CHECK_TRAP(Binop(Lt<float>));
        break;

      case Opcode::F32Le:
        CHECK_TRAP(Binop(Le<float>));
        break;

      case Opcode::F32Gt:
        CHECK_TRAP(Binop(Gt<float>));
        break;

      case Opcode::F32Ge:
        CHECK_TRAP(Binop(Ge<float>));
        break;

      case Opcode::F64Add:
        CHECK_TRAP(Binop(Add<double>));
        break;

      case Opcode::F64Sub:
        CHECK_TRAP(Binop(Sub<double>));
        break;

      case Opcode::F64Mul:
        CHECK_TRAP(Binop(Mul<double>));
        break;

      case Opcode::F64Div:
        CHECK_TRAP(Binop(FloatDiv<double>));
        break;

      case Opcode::F64Min:
        CHECK_TRAP(Binop(FloatMin<double>));
        break;

      case Opcode::F64Max:
        CHECK_TRAP(Binop(FloatMax<double>));
        break;

      case Opcode::F64Abs:
        CHECK_TRAP(Unop(FloatAbs<double>));
        break;

      case Opcode::F64Neg:
        CHECK_TRAP(Unop(FloatNeg<double>));
        break;

      case Opcode::F64Copysign:
        CHECK_TRAP(Binop(FloatCopySign<double>));
        break;

      case Opcode::F64Ceil:
        CHECK_TRAP(Unop(FloatCeil<double>));
        break;

      case Opcode::F64Floor:
        CHECK_TRAP(Unop(FloatFloor<double>));
        break;

      case Opcode::F64Trunc:
        CHECK_TRAP(Unop(FloatTrunc<double>));
        break;

      case Opcode::F64Nearest:
        CHECK_TRAP(Unop(FloatNearest<double>));
        break;

      case Opcode::F64Sqrt:
        CHECK_TRAP(Unop(FloatSqrt<double>));
        break;

      case Opcode::F64Eq:
        CHECK_TRAP(Binop(Eq<double>));
        break;

      case Opcode::F64Ne:
        CHECK_TRAP(Binop(Ne<double>));
        break;

      case Opcode::F64Lt:
        CHECK_TRAP(Binop(Lt<double>));
        break;

      case Opcode::F64Le:
        CHECK_TRAP(Binop(Le<double>));
        break;

      case Opcode::F64Gt:
        CHECK_TRAP(Binop(Gt<double>));
        break;

      case Opcode::F64Ge:
        CHECK_TRAP(Binop(Ge<double>));
        break;

      case Opcode::I32TruncF32S:
        CHECK_TRAP(UnopTrap(IntTrunc<int32_t, float>));
        break;

      case Opcode::I32TruncSatF32S:
        CHECK_TRAP(Unop(IntTruncSat<int32_t, float>));
        break;

      case Opcode::I32TruncF64S:
        CHECK_TRAP(UnopTrap(IntTrunc<int32_t, double>));
        break;

      case Opcode::I32TruncSatF64S:
        CHECK_TRAP(Unop(IntTruncSat<int32_t, double>));
        break;

      case Opcode::I32TruncF32U:
        CHECK_TRAP(UnopTrap(IntTrunc<uint32_t, float>));
        break;

      case Opcode::I32TruncSatF32U:
        CHECK_TRAP(Unop(IntTruncSat<uint32_t, float>));
        break;

      case Opcode::I32TruncF64U:
        CHECK_TRAP(UnopTrap(IntTrunc<uint32_t, double>));
        break;

      case Opcode::I32TruncSatF64U:
        CHECK_TRAP(Unop(IntTruncSat<uint32_t, double>));
        break;

      case Opcode::I32WrapI64:
        CHECK_TRAP(Push<uint32_t>(Pop<uint64_t>()));
        break;

      case Opcode::I64TruncF32S:
        CHECK_TRAP(UnopTrap(IntTrunc<int64_t, float>));
        break;

      case Opcode::I64TruncSatF32S:
        CHECK_TRAP(Unop(IntTruncSat<int64_t, float>));
        break;

      case Opcode::I64TruncF64S:
        CHECK_TRAP(UnopTrap(IntTrunc<int64_t, double>));
        break;

      case Opcode::I64TruncSatF64S:
        CHECK_TRAP(Unop(IntTruncSat<int64_t, double>));
        break;

      case Opcode::I64TruncF32U:
        CHECK_TRAP(UnopTrap(IntTrunc<uint64_t, float>));
        break;

      case Opcode::I64TruncSatF32U:
        CHECK_TRAP(Unop(IntTruncSat<uint64_t, float>));
        break;

      case Opcode::I64TruncF64U:
        CHECK_TRAP(UnopTrap(IntTrunc<uint64_t, double>));
        break;

      case Opcode::I64TruncSatF64U:
        CHECK_TRAP(Unop(IntTruncSat<uint64_t, double>));
        break;

      case Opcode::I64ExtendI32S:
        CHECK_TRAP(Push<uint64_t>(Pop<int32_t>()));
        break;

      case Opcode::I64ExtendI32U:
        CHECK_TRAP(Push<uint64_t>(Pop<uint32_t>()));
        break;

      case Opcode::F32ConvertI32S:
        CHECK_TRAP(Push<float>(Pop<int32_t>()));
        break;

      case Opcode::F32ConvertI32U:
        CHECK_TRAP(Push<float>(Pop<uint32_t>()));
        break;

      case Opcode::F32ConvertI64S:
        CHECK_TRAP(Push<float>(Pop<int64_t>()));
        break;

      case Opcode::F32ConvertI64U:
        CHECK_TRAP(Push<float>(wabt_convert_uint64_to_float(Pop<uint64_t>())));
        break;

      case Opcode::F32DemoteF64: {
        typedef FloatTraits<float> F32Traits;

        uint64_t value = PopRep<double>();
        if (WABT_LIKELY((IsConversionInRange<float, double>(value)))) {
          CHECK_TRAP(Push<float>(FromRep<double>(value)));
        } else if (IsInRangeF64DemoteF32RoundToF32Max(value)) {
          CHECK_TRAP(PushRep<float>(F32Traits::kMax));
        } else if (IsInRangeF64DemoteF32RoundToNegF32Max(value)) {
          CHECK_TRAP(PushRep<float>(F32Traits::kNegMax));
        } else if (FloatTraits<double>::IsNan(value)) {
          CHECK_TRAP(PushRep<float>(F32Traits::kQuietNan));
        } else {
          // Infinity.
          uint32_t sign = (value >> 32) & F32Traits::kSignMask;
          CHECK_TRAP(PushRep<float>(sign | F32Traits::kInf));
        }
        break;
      }

      case Opcode::F32ReinterpretI32:
        CHECK_TRAP(PushRep<float>(Pop<uint32_t>()));
        break;

      case Opcode::F64ConvertI32S:
        CHECK_TRAP(Push<double>(Pop<int32_t>()));
        break;

      case Opcode::F64ConvertI32U:
        CHECK_TRAP(Push<double>(Pop<uint32_t>()));
        break;

      case Opcode::F64ConvertI64S:
        CHECK_TRAP(Push<double>(Pop<int64_t>()));
        break;

      case Opcode::F64ConvertI64U:
        CHECK_TRAP(
            Push<double>(wabt_convert_uint64_to_double(Pop<uint64_t>())));
        break;

      case Opcode::F64PromoteF32: {
        uint32_t value = PopRep<float>();
        if (WABT_UNLIKELY(FloatTraits<float>::IsNan(value))) {
          CHECK_TRAP(PushRep<double>(FloatTraits<double>::kQuietNan));
        } else {
          CHECK_TRAP(Push<double>(Bitcast<float>(value)));
        }
        break;
      }

      case Opcode::F64ReinterpretI64:
        CHECK_TRAP(PushRep<double>(Pop<uint64_t>()));
        break;

      case Opcode::I32ReinterpretF32:
        CHECK_TRAP(Push<uint32_t>(PopRep<float>()));
        break;

      case Opcode::I64ReinterpretF64:
        CHECK_TRAP(Push<uint64_t>(PopRep<double>()));
        break;

      case Opcode::I32Rotr:
        CHECK_TRAP(Binop(IntRotr<uint32_t>));
        break;

      case Opcode::I32Rotl:
        CHECK_TRAP(Binop(IntRotl<uint32_t>));
        break;

      case Opcode::I64Rotr:
        CHECK_TRAP(Binop(IntRotr<uint64_t>));
        break;

      case Opcode::I64Rotl:
        CHECK_TRAP(Binop(IntRotl<uint64_t>));
        break;

      case Opcode::I64Eqz:
        CHECK_TRAP(Unop(IntEqz<uint32_t, uint64_t>));
        break;

      case Opcode::I32Extend8S:
        CHECK_TRAP(Unop(IntExtendS<uint32_t, int8_t>));
        break;

      case Opcode::I32Extend16S:
        CHECK_TRAP(Unop(IntExtendS<uint32_t, int16_t>));
        break;

      case Opcode::I64Extend8S:
        CHECK_TRAP(Unop(IntExtendS<uint64_t, int8_t>));
        break;

      case Opcode::I64Extend16S:
        CHECK_TRAP(Unop(IntExtendS<uint64_t, int16_t>));
        break;

      case Opcode::I64Extend32S:
        CHECK_TRAP(Unop(IntExtendS<uint64_t, int32_t>));
        break;

      case Opcode::InterpAlloca: {
        uint32_t old_value_stack_top = value_stack_top_;
        size_t count = ReadU32(&pc);
        value_stack_top_ += count;
        CHECK_STACK();
        memset(&value_stack_[old_value_stack_top], 0, count * sizeof(Value));
        break;
      }

      case Opcode::InterpBrUnless: {
        IstreamOffset new_pc = ReadU32(&pc);
        if (!Pop<uint32_t>()) {
          GOTO(new_pc);
        }
        break;
      }

      case Opcode::Drop:
        (void)Pop();
        break;

      case Opcode::InterpDropKeep: {
        uint32_t drop_count = ReadU32(&pc);
        uint32_t keep_count = ReadU32(&pc);
        DropKeep(drop_count, keep_count);
        break;
      }

      case Opcode::Nop:
        break;

      case Opcode::I32AtomicWait:
      case Opcode::I64AtomicWait:
      case Opcode::AtomicNotify:
        // TODO(binji): Implement.
        TRAP(Unreachable);
        break;

      case Opcode::V128Const: {
        CHECK_TRAP(PushRep<v128>(ReadV128(&pc)));
        break;
      }

      case Opcode::V128Load:
        CHECK_TRAP(Load<v128>(&pc));
        break;

      case Opcode::V128Store:
        CHECK_TRAP(Store<v128>(&pc));
        break;

      case Opcode::I8X16Splat: {
        uint8_t lane_data = Pop<uint32_t>();
        CHECK_TRAP(Push<v128>(SimdSplat<v128, uint8_t>(lane_data)));
        break;
      }

      case Opcode::I16X8Splat: {
        uint16_t lane_data = Pop<uint32_t>();
        CHECK_TRAP(Push<v128>(SimdSplat<v128, uint16_t>(lane_data)));
        break;
      }

      case Opcode::I32X4Splat: {
        uint32_t lane_data = Pop<uint32_t>();
        CHECK_TRAP(Push<v128>(SimdSplat<v128, uint32_t>(lane_data)));
        break;
      }

      case Opcode::I64X2Splat: {
        uint64_t lane_data = Pop<uint64_t>();
        CHECK_TRAP(Push<v128>(SimdSplat<v128, uint64_t>(lane_data)));
        break;
      }

      case Opcode::F32X4Splat: {
        float lane_data = Pop<float>();
        CHECK_TRAP(Push<v128>(SimdSplat<v128, float>(lane_data)));
        break;
      }

      case Opcode::F64X2Splat: {
        double lane_data = Pop<double>();
        CHECK_TRAP(Push<v128>(SimdSplat<v128, double>(lane_data)));
        break;
      }

      case Opcode::I8X16ExtractLaneS: {
        v128 lane_val = static_cast<v128>(Pop<v128>());
        uint32_t lane_idx = ReadU8(&pc);
        CHECK_TRAP(PushRep<int32_t>(
            SimdExtractLane<int32_t, v128, int8_t>(lane_val, lane_idx)));
        break;
      }

      case Opcode::I8X16ExtractLaneU: {
        v128 lane_val = static_cast<v128>(Pop<v128>());
        uint32_t lane_idx = ReadU8(&pc);
        CHECK_TRAP(PushRep<int32_t>(
            SimdExtractLane<int32_t, v128, uint8_t>(lane_val, lane_idx)));
        break;
      }

      case Opcode::I16X8ExtractLaneS: {
        v128 lane_val = static_cast<v128>(Pop<v128>());
        uint32_t lane_idx = ReadU8(&pc);
        CHECK_TRAP(PushRep<int32_t>(
            SimdExtractLane<int32_t, v128, int16_t>(lane_val, lane_idx)));
        break;
      }

      case Opcode::I16X8ExtractLaneU: {
        v128 lane_val = static_cast<v128>(Pop<v128>());
        uint32_t lane_idx = ReadU8(&pc);
        CHECK_TRAP(PushRep<int32_t>(
            SimdExtractLane<int32_t, v128, uint16_t>(lane_val, lane_idx)));
        break;
      }

      case Opcode::I32X4ExtractLane: {
        v128 lane_val = static_cast<v128>(Pop<v128>());
        uint32_t lane_idx = ReadU8(&pc);
        CHECK_TRAP(PushRep<int32_t>(
            SimdExtractLane<int32_t, v128, int32_t>(lane_val, lane_idx)));
        break;
      }

      case Opcode::I64X2ExtractLane: {
        v128 lane_val = static_cast<v128>(Pop<v128>());
        uint32_t lane_idx = ReadU8(&pc);
        CHECK_TRAP(PushRep<int64_t>(
            SimdExtractLane<int64_t, v128, int64_t>(lane_val, lane_idx)));
        break;
      }

      case Opcode::F32X4ExtractLane: {
        v128 lane_val = static_cast<v128>(Pop<v128>());
        uint32_t lane_idx = ReadU8(&pc);
        CHECK_TRAP(PushRep<float>(
            SimdExtractLane<int32_t, v128, int32_t>(lane_val, lane_idx)));
        break;
      }

      case Opcode::F64X2ExtractLane: {
        v128 lane_val = static_cast<v128>(Pop<v128>());
        uint32_t lane_idx = ReadU8(&pc);
        CHECK_TRAP(PushRep<double>(
            SimdExtractLane<int64_t, v128, int64_t>(lane_val, lane_idx)));
        break;
      }

      case Opcode::I8X16ReplaceLane: {
        int8_t lane_val = static_cast<int8_t>(Pop<int32_t>());
        v128 value = static_cast<v128>(Pop<v128>());
        uint32_t lane_idx = ReadU8(&pc);
        CHECK_TRAP(Push<v128>(
            SimdReplaceLane<v128, v128, int8_t>(value, lane_idx, lane_val)));
        break;
      }

      case Opcode::I16X8ReplaceLane: {
        int16_t lane_val = static_cast<int16_t>(Pop<int32_t>());
        v128 value = static_cast<v128>(Pop<v128>());
        uint32_t lane_idx = ReadU8(&pc);
        CHECK_TRAP(Push<v128>(
            SimdReplaceLane<v128, v128, int16_t>(value, lane_idx, lane_val)));
        break;
      }

      case Opcode::I32X4ReplaceLane: {
        int32_t lane_val = Pop<int32_t>();
        v128 value = static_cast<v128>(Pop<v128>());
        uint32_t lane_idx = ReadU8(&pc);
        CHECK_TRAP(Push<v128>(
            SimdReplaceLane<v128, v128, int32_t>(value, lane_idx, lane_val)));
        break;
      }

      case Opcode::I64X2ReplaceLane: {
        int64_t lane_val = Pop<int64_t>();
        v128 value = static_cast<v128>(Pop<v128>());
        uint32_t lane_idx = ReadU8(&pc);
        CHECK_TRAP(Push<v128>(
            SimdReplaceLane<v128, v128, int64_t>(value, lane_idx, lane_val)));
        break;
      }

      case Opcode::F32X4ReplaceLane: {
        float lane_val = Pop<float>();
        v128 value = static_cast<v128>(Pop<v128>());
        uint32_t lane_idx = ReadU8(&pc);
        CHECK_TRAP(Push<v128>(
            SimdReplaceLane<v128, v128, float>(value, lane_idx, lane_val)));
        break;
      }

      case Opcode::F64X2ReplaceLane: {
        double lane_val = Pop<double>();
        v128 value = static_cast<v128>(Pop<v128>());
        uint32_t lane_idx = ReadU8(&pc);
        CHECK_TRAP(Push<v128>(
            SimdReplaceLane<v128, v128, double>(value, lane_idx, lane_val)));
        break;
      }

      case Opcode::V8X16Shuffle: {
        const int32_t lanes = 16;
        // Define SIMD data array for Simd add by Lanes.
        int8_t simd_data_ret[lanes];
        int8_t simd_data_0[lanes];
        int8_t simd_data_1[lanes];
        int8_t simd_shuffle[lanes];

        v128 v2 = PopRep<v128>();
        v128 v1 = PopRep<v128>();
        v128 shuffle_imm = ReadV128(&pc);

        // Convert intput SIMD data to array.
        memcpy(simd_data_0, &v1, sizeof(v128));
        memcpy(simd_data_1, &v2, sizeof(v128));
        memcpy(simd_shuffle, &shuffle_imm, sizeof(v128));

        // Constuct the Simd value by Lane data and Lane nums.
        for (int32_t i = 0; i < lanes; i++) {
          int8_t lane_idx = simd_shuffle[i];
          simd_data_ret[i] = (lane_idx < lanes) ? simd_data_0[lane_idx]
                                                : simd_data_1[lane_idx - lanes];
        }

        CHECK_TRAP(PushRep<v128>(Bitcast<v128>(simd_data_ret)));
        break;
      }

      case Opcode::I8X16Add:
        CHECK_TRAP(SimdBinop<v128, uint8_t>(Add<uint32_t>));
        break;

      case Opcode::I16X8Add:
        CHECK_TRAP(SimdBinop<v128, uint16_t>(Add<uint32_t>));
        break;

      case Opcode::I32X4Add:
        CHECK_TRAP(SimdBinop<v128, uint32_t>(Add<uint32_t>));
        break;

      case Opcode::I64X2Add:
        CHECK_TRAP(SimdBinop<v128, uint64_t>(Add<uint64_t>));
        break;

      case Opcode::I8X16Sub:
        CHECK_TRAP(SimdBinop<v128, uint8_t>(Sub<uint32_t>));
        break;

      case Opcode::I16X8Sub:
        CHECK_TRAP(SimdBinop<v128, uint16_t>(Sub<uint32_t>));
        break;

      case Opcode::I32X4Sub:
        CHECK_TRAP(SimdBinop<v128, uint32_t>(Sub<uint32_t>));
        break;

      case Opcode::I64X2Sub:
        CHECK_TRAP(SimdBinop<v128, uint64_t>(Sub<uint64_t>));
        break;

      case Opcode::I8X16Mul:
        CHECK_TRAP(SimdBinop<v128, uint8_t>(Mul<uint32_t>));
        break;

      case Opcode::I16X8Mul:
        CHECK_TRAP(SimdBinop<v128, uint16_t>(Mul<uint32_t>));
        break;

      case Opcode::I32X4Mul:
        CHECK_TRAP(SimdBinop<v128, uint32_t>(Mul<uint32_t>));
        break;

      case Opcode::I8X16Neg:
        CHECK_TRAP(SimdUnop<v128, int8_t>(IntNeg<int32_t>));
        break;

      case Opcode::I16X8Neg:
        CHECK_TRAP(SimdUnop<v128, int16_t>(IntNeg<int32_t>));
        break;

      case Opcode::I32X4Neg:
        CHECK_TRAP(SimdUnop<v128, int32_t>(IntNeg<int32_t>));
        break;

      case Opcode::I64X2Neg:
        CHECK_TRAP(SimdUnop<v128, int64_t>(IntNeg<int64_t>));
        break;

      case Opcode::I8X16AddSaturateS:
        CHECK_TRAP(SimdBinop<v128, int8_t>(AddSaturate<int32_t, int8_t>));
        break;

      case Opcode::I8X16AddSaturateU:
        CHECK_TRAP(SimdBinop<v128, uint8_t>(AddSaturate<uint32_t, uint8_t>));
        break;

      case Opcode::I16X8AddSaturateS:
        CHECK_TRAP(SimdBinop<v128, int16_t>(AddSaturate<int32_t, int16_t>));
        break;

      case Opcode::I16X8AddSaturateU:
        CHECK_TRAP(SimdBinop<v128, uint16_t>(AddSaturate<uint32_t, uint16_t>));
        break;

      case Opcode::I8X16SubSaturateS:
        CHECK_TRAP(SimdBinop<v128, int8_t>(SubSaturate<int32_t, int8_t>));
        break;

      case Opcode::I8X16SubSaturateU:
        CHECK_TRAP(SimdBinop<v128, uint8_t>(SubSaturate<int32_t, uint8_t>));
        break;

      case Opcode::I16X8SubSaturateS:
        CHECK_TRAP(SimdBinop<v128, int16_t>(SubSaturate<int32_t, int16_t>));
        break;

      case Opcode::I16X8SubSaturateU:
        CHECK_TRAP(SimdBinop<v128, uint16_t>(SubSaturate<int32_t, uint16_t>));
        break;

      case Opcode::I8X16Shl: {
        uint32_t shift_count = Pop<uint32_t>();
        shift_count = shift_count % 8;
        CHECK_TRAP(Push<v128>(SimdSplat<v128, uint8_t>(shift_count)));
        CHECK_TRAP(SimdBinop<v128, uint8_t>(IntShl<uint32_t>));
        break;
      }

      case Opcode::I16X8Shl: {
        uint32_t shift_count = Pop<uint32_t>();
        shift_count = shift_count % 16;
        CHECK_TRAP(Push<v128>(SimdSplat<v128, uint16_t>(shift_count)));
        CHECK_TRAP(SimdBinop<v128, uint16_t>(IntShl<uint32_t>));
        break;
      }

      case Opcode::I32X4Shl: {
        uint32_t shift_count = Pop<uint32_t>();
        shift_count = shift_count % 32;
        CHECK_TRAP(Push<v128>(SimdSplat<v128, uint32_t>(shift_count)));
        CHECK_TRAP(SimdBinop<v128, uint32_t>(IntShl<uint32_t>));
        break;
      }

      case Opcode::I64X2Shl: {
        uint32_t shift_count = Pop<uint32_t>();
        shift_count = shift_count % 64;
        CHECK_TRAP(Push<v128>(SimdSplat<v128, uint64_t>(shift_count)));
        CHECK_TRAP(SimdBinop<v128, uint64_t>(IntShl<uint64_t>));
        break;
      }

      case Opcode::I8X16ShrS: {
        uint32_t shift_count = Pop<uint32_t>();
        shift_count = shift_count % 8;
        CHECK_TRAP(Push<v128>(SimdSplat<v128, uint8_t>(shift_count)));
        CHECK_TRAP(SimdBinop<v128, int8_t>(IntShr<int32_t>));
        break;
      }

      case Opcode::I8X16ShrU: {
        uint32_t shift_count = Pop<uint32_t>();
        shift_count = shift_count % 8;
        CHECK_TRAP(Push<v128>(SimdSplat<v128, uint8_t>(shift_count)));
        CHECK_TRAP(SimdBinop<v128, uint8_t>(IntShr<uint32_t>));
        break;
      }

      case Opcode::I16X8ShrS: {
        uint32_t shift_count = Pop<uint32_t>();
        shift_count = shift_count % 16;
        CHECK_TRAP(Push<v128>(SimdSplat<v128, uint16_t>(shift_count)));
        CHECK_TRAP(SimdBinop<v128, int16_t>(IntShr<int32_t>));
        break;
      }

      case Opcode::I16X8ShrU: {
        uint32_t shift_count = Pop<uint32_t>();
        shift_count = shift_count % 16;
        CHECK_TRAP(Push<v128>(SimdSplat<v128, uint16_t>(shift_count)));
        CHECK_TRAP(SimdBinop<v128, uint16_t>(IntShr<uint32_t>));
        break;
      }

      case Opcode::I32X4ShrS: {
        uint32_t shift_count = Pop<uint32_t>();
        shift_count = shift_count % 32;
        CHECK_TRAP(Push<v128>(SimdSplat<v128, uint32_t>(shift_count)));
        CHECK_TRAP(SimdBinop<v128, int32_t>(IntShr<int32_t>));
        break;
      }

      case Opcode::I32X4ShrU: {
        uint32_t shift_count = Pop<uint32_t>();
        shift_count = shift_count % 32;
        CHECK_TRAP(Push<v128>(SimdSplat<v128, uint32_t>(shift_count)));
        CHECK_TRAP(SimdBinop<v128, uint32_t>(IntShr<uint32_t>));
        break;
      }

      case Opcode::I64X2ShrS: {
        uint32_t shift_count = Pop<uint32_t>();
        shift_count = shift_count % 64;
        CHECK_TRAP(Push<v128>(SimdSplat<v128, uint64_t>(shift_count)));
        CHECK_TRAP(SimdBinop<v128, int64_t>(IntShr<int64_t>));
        break;
      }

      case Opcode::I64X2ShrU: {
        uint32_t shift_count = Pop<uint32_t>();
        shift_count = shift_count % 64;
        CHECK_TRAP(Push<v128>(SimdSplat<v128, uint64_t>(shift_count)));
        CHECK_TRAP(SimdBinop<v128, uint64_t>(IntShr<uint64_t>));
        break;
      }

      case Opcode::V128And:
        CHECK_TRAP(SimdBinop<v128, uint64_t>(IntAnd<uint64_t>));
        break;

      case Opcode::V128Or:
        CHECK_TRAP(SimdBinop<v128, uint64_t>(IntOr<uint64_t>));
        break;

      case Opcode::V128Xor:
        CHECK_TRAP(SimdBinop<v128, uint64_t>(IntXor<uint64_t>));
        break;

      case Opcode::V128Not:
        CHECK_TRAP(SimdUnop<v128, uint64_t>(IntNot<uint64_t>));
        break;

      case Opcode::V128BitSelect: {
        // Follow Wasm Simd spec to compute V128BitSelect:
        // v128.or(v128.and(v1, c), v128.and(v2, v128.not(c)))
        v128 c_mask = PopRep<v128>();
        v128 v2 = PopRep<v128>();
        v128 v1 = PopRep<v128>();
        // 1. v128.and(v1, c)
        CHECK_TRAP(Push<v128>(v1));
        CHECK_TRAP(Push<v128>(c_mask));
        CHECK_TRAP(SimdBinop<v128, uint64_t>(IntAnd<uint64_t>));
        // 2. v128.and(v2, v128.not(c))
        CHECK_TRAP(Push<v128>(v2));
        CHECK_TRAP(Push<v128>(c_mask));
        CHECK_TRAP(SimdUnop<v128, uint64_t>(IntNot<uint64_t>));
        CHECK_TRAP(SimdBinop<v128, uint64_t>(IntAnd<uint64_t>));
        // 3. v128.or( 1 , 2)
        CHECK_TRAP(SimdBinop<v128, uint64_t>(IntOr<uint64_t>));
        break;
      }

      case Opcode::I8X16AnyTrue: {
        v128 value = PopRep<v128>();
        CHECK_TRAP(Push<int32_t>(SimdIsLaneTrue<v128, uint8_t>(value, 1)));
        break;
      }

      case Opcode::I16X8AnyTrue: {
        v128 value = PopRep<v128>();
        CHECK_TRAP(Push<int32_t>(SimdIsLaneTrue<v128, uint16_t>(value, 1)));
        break;
      }

      case Opcode::I32X4AnyTrue: {
        v128 value = PopRep<v128>();
        CHECK_TRAP(Push<int32_t>(SimdIsLaneTrue<v128, uint32_t>(value, 1)));
        break;
      }

      case Opcode::I64X2AnyTrue: {
        v128 value = PopRep<v128>();
        CHECK_TRAP(Push<int32_t>(SimdIsLaneTrue<v128, uint64_t>(value, 1)));
        break;
      }

      case Opcode::I8X16AllTrue: {
        v128 value = PopRep<v128>();
        CHECK_TRAP(Push<int32_t>(SimdIsLaneTrue<v128, uint8_t>(value, 16)));
        break;
      }

      case Opcode::I16X8AllTrue: {
        v128 value = PopRep<v128>();
        CHECK_TRAP(Push<int32_t>(SimdIsLaneTrue<v128, uint16_t>(value, 8)));
        break;
      }

      case Opcode::I32X4AllTrue: {
        v128 value = PopRep<v128>();
        CHECK_TRAP(Push<int32_t>(SimdIsLaneTrue<v128, uint32_t>(value, 4)));
        break;
      }

      case Opcode::I64X2AllTrue: {
        v128 value = PopRep<v128>();
        CHECK_TRAP(Push<int32_t>(SimdIsLaneTrue<v128, uint64_t>(value, 2)));
        break;
      }

      case Opcode::I8X16Eq:
        CHECK_TRAP(SimdRelBinop<v128, int8_t>(Eq<int32_t>));
        break;

      case Opcode::I16X8Eq:
        CHECK_TRAP(SimdRelBinop<v128, int16_t>(Eq<int32_t>));
        break;

      case Opcode::I32X4Eq:
        CHECK_TRAP(SimdRelBinop<v128, int32_t>(Eq<int32_t>));
        break;

      case Opcode::F32X4Eq:
        CHECK_TRAP(SimdRelBinop<v128, int32_t>(Eq<float>));
        break;

      case Opcode::F64X2Eq:
        CHECK_TRAP(SimdRelBinop<v128, int64_t>(Eq<double>));
        break;

      case Opcode::I8X16Ne:
        CHECK_TRAP(SimdRelBinop<v128, int8_t>(Ne<int32_t>));
        break;

      case Opcode::I16X8Ne:
        CHECK_TRAP(SimdRelBinop<v128, int16_t>(Ne<int32_t>));
        break;

      case Opcode::I32X4Ne:
        CHECK_TRAP(SimdRelBinop<v128, int32_t>(Ne<int32_t>));
        break;

      case Opcode::F32X4Ne:
        CHECK_TRAP(SimdRelBinop<v128, int32_t>(Ne<float>));
        break;

      case Opcode::F64X2Ne:
        CHECK_TRAP(SimdRelBinop<v128, int64_t>(Ne<double>));
        break;

      case Opcode::I8X16LtS:
        CHECK_TRAP(SimdRelBinop<v128, int8_t>(Lt<int32_t>));
        break;

      case Opcode::I8X16LtU:
        CHECK_TRAP(SimdRelBinop<v128, uint8_t>(Lt<uint32_t>));
        break;

      case Opcode::I16X8LtS:
        CHECK_TRAP(SimdRelBinop<v128, int16_t>(Lt<int32_t>));
        break;

      case Opcode::I16X8LtU:
        CHECK_TRAP(SimdRelBinop<v128, uint16_t>(Lt<uint32_t>));
        break;

      case Opcode::I32X4LtS:
        CHECK_TRAP(SimdRelBinop<v128, int32_t>(Lt<int32_t>));
        break;

      case Opcode::I32X4LtU:
        CHECK_TRAP(SimdRelBinop<v128, uint32_t>(Lt<uint32_t>));
        break;

      case Opcode::F32X4Lt:
        CHECK_TRAP(SimdRelBinop<v128, int32_t>(Lt<float>));
        break;

      case Opcode::F64X2Lt:
        CHECK_TRAP(SimdRelBinop<v128, int64_t>(Lt<double>));
        break;

      case Opcode::I8X16LeS:
        CHECK_TRAP(SimdRelBinop<v128, int8_t>(Le<int32_t>));
        break;

      case Opcode::I8X16LeU:
        CHECK_TRAP(SimdRelBinop<v128, uint8_t>(Le<uint32_t>));
        break;

      case Opcode::I16X8LeS:
        CHECK_TRAP(SimdRelBinop<v128, int16_t>(Le<int32_t>));
        break;

      case Opcode::I16X8LeU:
        CHECK_TRAP(SimdRelBinop<v128, uint16_t>(Le<uint32_t>));
        break;

      case Opcode::I32X4LeS:
        CHECK_TRAP(SimdRelBinop<v128, int32_t>(Le<int32_t>));
        break;

      case Opcode::I32X4LeU:
        CHECK_TRAP(SimdRelBinop<v128, uint32_t>(Le<uint32_t>));
        break;

      case Opcode::F32X4Le:
        CHECK_TRAP(SimdRelBinop<v128, int32_t>(Le<float>));
        break;

      case Opcode::F64X2Le:
        CHECK_TRAP(SimdRelBinop<v128, int64_t>(Le<double>));
        break;

      case Opcode::I8X16GtS:
        CHECK_TRAP(SimdRelBinop<v128, int8_t>(Gt<int32_t>));
        break;

      case Opcode::I8X16GtU:
        CHECK_TRAP(SimdRelBinop<v128, uint8_t>(Gt<uint32_t>));
        break;

      case Opcode::I16X8GtS:
        CHECK_TRAP(SimdRelBinop<v128, int16_t>(Gt<int32_t>));
        break;

      case Opcode::I16X8GtU:
        CHECK_TRAP(SimdRelBinop<v128, uint16_t>(Gt<uint32_t>));
        break;

      case Opcode::I32X4GtS:
        CHECK_TRAP(SimdRelBinop<v128, int32_t>(Gt<int32_t>));
        break;

      case Opcode::I32X4GtU:
        CHECK_TRAP(SimdRelBinop<v128, uint32_t>(Gt<uint32_t>));
        break;

      case Opcode::F32X4Gt:
        CHECK_TRAP(SimdRelBinop<v128, int32_t>(Gt<float>));
        break;

      case Opcode::F64X2Gt:
        CHECK_TRAP(SimdRelBinop<v128, int64_t>(Gt<double>));
        break;

      case Opcode::I8X16GeS:
        CHECK_TRAP(SimdRelBinop<v128, int8_t>(Ge<int32_t>));
        break;

      case Opcode::I8X16GeU:
        CHECK_TRAP(SimdRelBinop<v128, uint8_t>(Ge<uint32_t>));
        break;

      case Opcode::I16X8GeS:
        CHECK_TRAP(SimdRelBinop<v128, int16_t>(Ge<int32_t>));
        break;

      case Opcode::I16X8GeU:
        CHECK_TRAP(SimdRelBinop<v128, uint16_t>(Ge<uint32_t>));
        break;

      case Opcode::I32X4GeS:
        CHECK_TRAP(SimdRelBinop<v128, int32_t>(Ge<int32_t>));
        break;

      case Opcode::I32X4GeU:
        CHECK_TRAP(SimdRelBinop<v128, uint32_t>(Ge<uint32_t>));
        break;

      case Opcode::F32X4Ge:
        CHECK_TRAP(SimdRelBinop<v128, int32_t>(Ge<float>));
        break;

      case Opcode::F64X2Ge:
        CHECK_TRAP(SimdRelBinop<v128, int64_t>(Ge<double>));
        break;

      case Opcode::F32X4Neg:
        CHECK_TRAP(SimdUnop<v128, int32_t>(FloatNeg<float>));
        break;

      case Opcode::F64X2Neg:
        CHECK_TRAP(SimdUnop<v128, int64_t>(FloatNeg<double>));
        break;

      case Opcode::F32X4Abs:
        CHECK_TRAP(SimdUnop<v128, int32_t>(FloatAbs<float>));
        break;

      case Opcode::F64X2Abs:
        CHECK_TRAP(SimdUnop<v128, int64_t>(FloatAbs<double>));
        break;

      case Opcode::F32X4Min:
        CHECK_TRAP(SimdBinop<v128, int32_t>(FloatMin<float>));
        break;

      case Opcode::F64X2Min:
        CHECK_TRAP(SimdBinop<v128, int64_t>(FloatMin<double>));
        break;

      case Opcode::F32X4Max:
        CHECK_TRAP(SimdBinop<v128, int32_t>(FloatMax<float>));
        break;

      case Opcode::F64X2Max:
        CHECK_TRAP(SimdBinop<v128, int64_t>(FloatMax<double>));
        break;

      case Opcode::F32X4Add:
        CHECK_TRAP(SimdBinop<v128, int32_t>(Add<float>));
        break;

      case Opcode::F64X2Add:
        CHECK_TRAP(SimdBinop<v128, int64_t>(Add<double>));
        break;

      case Opcode::F32X4Sub:
        CHECK_TRAP(SimdBinop<v128, int32_t>(Sub<float>));
        break;

      case Opcode::F64X2Sub:
        CHECK_TRAP(SimdBinop<v128, int64_t>(Sub<double>));
        break;

      case Opcode::F32X4Div:
        CHECK_TRAP(SimdBinop<v128, int32_t>(FloatDiv<float>));
        break;

      case Opcode::F64X2Div:
        CHECK_TRAP(SimdBinop<v128, int64_t>(FloatDiv<double>));
        break;

      case Opcode::F32X4Mul:
        CHECK_TRAP(SimdBinop<v128, int32_t>(Mul<float>));
        break;

      case Opcode::F64X2Mul:
        CHECK_TRAP(SimdBinop<v128, int64_t>(Mul<double>));
        break;

      case Opcode::F32X4Sqrt:
        CHECK_TRAP(SimdUnop<v128, int32_t>(FloatSqrt<float>));
        break;

      case Opcode::F64X2Sqrt:
        CHECK_TRAP(SimdUnop<v128, int64_t>(FloatSqrt<double>));
        break;

      case Opcode::F32X4ConvertI32X4S:
        CHECK_TRAP(SimdUnop<v128, int32_t>(SimdConvert<float, int32_t>));
        break;

      case Opcode::F32X4ConvertI32X4U:
        CHECK_TRAP(SimdUnop<v128, uint32_t>(SimdConvert<float, uint32_t>));
        break;

      case Opcode::F64X2ConvertI64X2S:
        CHECK_TRAP(SimdUnop<v128, int64_t>(SimdConvert<double, int64_t>));
        break;

      case Opcode::F64X2ConvertI64X2U:
        CHECK_TRAP(SimdUnop<v128, uint64_t>(SimdConvert<double, uint64_t>));
        break;

      case Opcode::I32X4TruncSatF32X4S:
        CHECK_TRAP(SimdUnop<v128, int32_t>(IntTruncSat<int32_t, float>));
        break;

      case Opcode::I32X4TruncSatF32X4U:
        CHECK_TRAP(SimdUnop<v128, uint32_t>(IntTruncSat<uint32_t, float>));
        break;

      case Opcode::I64X2TruncSatF64X2S:
        CHECK_TRAP(SimdUnop<v128, int64_t>(IntTruncSat<int64_t, double>));
        break;

      case Opcode::I64X2TruncSatF64X2U:
        CHECK_TRAP(SimdUnop<v128, uint64_t>(IntTruncSat<uint64_t, double>));
        break;

      case Opcode::TableGet:
      case Opcode::TableSet:
      case Opcode::TableGrow:
      case Opcode::TableSize:
      case Opcode::RefNull:
      case Opcode::RefIsNull:
        WABT_UNREACHABLE;
        break;

      case Opcode::MemoryInit:
        WABT_UNREACHABLE;
        break;

      case Opcode::DataDrop:
        WABT_UNREACHABLE;
        break;

      case Opcode::MemoryCopy:
        WABT_UNREACHABLE;
        break;

      case Opcode::MemoryFill:
        WABT_UNREACHABLE;
        break;

      case Opcode::TableInit:
        WABT_UNREACHABLE;
        break;

      case Opcode::ElemDrop:
        WABT_UNREACHABLE;
        break;

      case Opcode::TableCopy:
        WABT_UNREACHABLE;
        break;

      // The following opcodes are either never generated or should never be
      // executed.
      case Opcode::Block:
      case Opcode::BrOnExn:
      case Opcode::Catch:
      case Opcode::Else:
      case Opcode::End:
      case Opcode::If:
      case Opcode::InterpData:
      case Opcode::Invalid:
      case Opcode::Loop:
      case Opcode::Rethrow:
      case Opcode::Throw:
      case Opcode::Try:
        WABT_UNREACHABLE;
        break;
    }
  }

exit_loop:
  pc_ = pc - istream;
  return result;
}

Executor::Executor(Environment* env,
                   Stream* trace_stream,
                   const Thread::Options& options)
    : env_(env), trace_stream_(trace_stream), thread_(env, options) {}

ExecResult Executor::RunFunction(Index func_index, const TypedValues& args) {
  ExecResult exec_result;
  Func* func = env_->GetFunc(func_index);
  FuncSignature* sig = env_->GetFuncSignature(func->sig_index);

  thread_.Reset();
  exec_result.result = PushArgs(sig, args);
  if (exec_result.result == Result::Ok) {
    exec_result.result =
        func->is_host ? thread_.CallHost(cast<HostFunc>(func))
                      : RunDefinedFunction(cast<DefinedFunc>(func)->offset);
    if (exec_result.result == Result::Ok) {
      CopyResults(sig, &exec_result.values);
    }
  }

  return exec_result;
}

ExecResult Executor::RunStartFunction(DefinedModule* module) {
  if (module->start_func_index == kInvalidIndex) {
    return ExecResult(Result::Ok);
  }

  if (trace_stream_) {
    trace_stream_->Writef(">>> running start function:\n");
  }

  TypedValues args;
  ExecResult exec_result = RunFunction(module->start_func_index, args);
  assert(exec_result.values.size() == 0);
  return exec_result;
}

ExecResult Executor::RunExport(const Export* export_, const TypedValues& args) {
  if (trace_stream_) {
    trace_stream_->Writef(">>> running export \"" PRIstringview "\":\n",
                          WABT_PRINTF_STRING_VIEW_ARG(export_->name));
  }

  assert(export_->kind == ExternalKind::Func);
  return RunFunction(export_->index, args);
}

ExecResult Executor::RunExportByName(Module* module,
                                     string_view name,
                                     const TypedValues& args) {
  Export* export_ = module->GetExport(name);
  if (!export_) {
    return ExecResult(Result::UnknownExport);
  }
  if (export_->kind != ExternalKind::Func) {
    return ExecResult(Result::ExportKindMismatch);
  }
  return RunExport(export_, args);
}

Result Executor::RunDefinedFunction(IstreamOffset function_offset) {
  Result result = Result::Ok;
  thread_.set_pc(function_offset);
  if (trace_stream_) {
    const int kNumInstructions = 1;
    while (result == Result::Ok) {
      thread_.Trace(trace_stream_);
      result = thread_.Run(kNumInstructions);
    }
  } else {
    const int kNumInstructions = 1000;
    while (result == Result::Ok) {
      result = thread_.Run(kNumInstructions);
    }
  }
  if (result != Result::Returned) {
    return result;
  }
  // Use OK instead of RETURNED for consistency.
  return Result::Ok;
}

Result Executor::PushArgs(const FuncSignature* sig, const TypedValues& args) {
  if (sig->param_types.size() != args.size()) {
    return Result::ArgumentTypeMismatch;
  }

  for (size_t i = 0; i < sig->param_types.size(); ++i) {
    if (sig->param_types[i] != args[i].type) {
      return Result::ArgumentTypeMismatch;
    }

    Result result = thread_.Push(args[i].value);
    if (result != Result::Ok) {
      return result;
    }
  }
  return Result::Ok;
}

void Executor::CopyResults(const FuncSignature* sig, TypedValues* out_results) {
  size_t expected_results = sig->result_types.size();
  assert(expected_results == thread_.NumValues());

  out_results->clear();
  for (size_t i = 0; i < expected_results; ++i)
    out_results->emplace_back(sig->result_types[i], thread_.ValueAt(i));
}

}  // namespace interp
}  // namespace wabt
