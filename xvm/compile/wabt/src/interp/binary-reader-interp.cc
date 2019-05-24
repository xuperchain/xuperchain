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

#include "src/interp/binary-reader-interp.h"

#include <cassert>
#include <cinttypes>
#include <cstdarg>
#include <cstdio>
#include <vector>

#include "src/binary-reader-nop.h"
#include "src/cast.h"
#include "src/feature.h"
#include "src/interp/interp.h"
#include "src/interp/interp-internal.h"
#include "src/stream.h"
#include "src/type-checker.h"

namespace wabt {

using namespace interp;

namespace {

typedef std::vector<Index> IndexVector;
typedef std::vector<IstreamOffset> IstreamOffsetVector;
typedef std::vector<IstreamOffsetVector> IstreamOffsetVectorVector;

struct Label {
  Label(IstreamOffset offset, IstreamOffset fixup_offset);

  IstreamOffset offset;
  IstreamOffset fixup_offset;
};

Label::Label(IstreamOffset offset, IstreamOffset fixup_offset)
    : offset(offset), fixup_offset(fixup_offset) {}

struct ElemSegmentInfo {
  ElemSegmentInfo(Index* dst, Index func_index)
      : dst(dst), func_index(func_index) {}

  Index* dst;
  Index func_index;
};

struct DataSegmentInfo {
  DataSegmentInfo(void* dst_data, const void* src_data, IstreamOffset size)
      : dst_data(dst_data), src_data(src_data), size(size) {}

  void* dst_data;        // Not owned.
  const void* src_data;  // Not owned.
  IstreamOffset size;
};

class BinaryReaderInterp : public BinaryReaderNop {
 public:
  BinaryReaderInterp(Environment* env,
                     DefinedModule* module,
                     std::unique_ptr<OutputBuffer> istream,
                     Errors* errors,
                     const Features& features);

  wabt::Result ReadBinary(DefinedModule* out_module);

  std::unique_ptr<OutputBuffer> ReleaseOutputBuffer();

  // Implement BinaryReader.
  bool OnError(const Error&) override;

  wabt::Result EndModule() override;

  wabt::Result OnTypeCount(Index count) override;
  wabt::Result OnType(Index index,
                      Index param_count,
                      Type* param_types,
                      Index result_count,
                      Type* result_types) override;

  wabt::Result OnImportFunc(Index import_index,
                            string_view module_name,
                            string_view field_name,
                            Index func_index,
                            Index sig_index) override;
  wabt::Result OnImportTable(Index import_index,
                             string_view module_name,
                             string_view field_name,
                             Index table_index,
                             Type elem_type,
                             const Limits* elem_limits) override;
  wabt::Result OnImportMemory(Index import_index,
                              string_view module_name,
                              string_view field_name,
                              Index memory_index,
                              const Limits* page_limits) override;
  wabt::Result OnImportGlobal(Index import_index,
                              string_view module_name,
                              string_view field_name,
                              Index global_index,
                              Type type,
                              bool mutable_) override;

  wabt::Result OnFunctionCount(Index count) override;
  wabt::Result OnFunction(Index index, Index sig_index) override;

  wabt::Result OnTable(Index index,
                       Type elem_type,
                       const Limits* elem_limits) override;

  wabt::Result OnMemory(Index index, const Limits* limits) override;

  wabt::Result OnGlobalCount(Index count) override;
  wabt::Result BeginGlobal(Index index, Type type, bool mutable_) override;
  wabt::Result EndGlobalInitExpr(Index index) override;

  wabt::Result OnExport(Index index,
                        ExternalKind kind,
                        Index item_index,
                        string_view name) override;

  wabt::Result OnStartFunction(Index func_index) override;

  wabt::Result BeginFunctionBody(Index index, Offset size) override;
  wabt::Result OnLocalDeclCount(Index count) override;
  wabt::Result OnLocalDecl(Index decl_index, Index count, Type type) override;

  wabt::Result OnAtomicLoadExpr(Opcode opcode,
                                uint32_t alignment_log2,
                                Address offset) override;
  wabt::Result OnAtomicStoreExpr(Opcode opcode,
                                 uint32_t alignment_log2,
                                 Address offset) override;
  wabt::Result OnAtomicRmwExpr(Opcode opcode,
                               uint32_t alignment_log2,
                               Address offset) override;
  wabt::Result OnAtomicRmwCmpxchgExpr(Opcode opcode,
                                      uint32_t alignment_log2,
                                      Address offset) override;
  wabt::Result OnAtomicWaitExpr(Opcode opcode,
                                uint32_t alignment_log2,
                                Address offset) override;
  wabt::Result OnAtomicNotifyExpr(Opcode opcode,
                                uint32_t alignment_log2,
                                Address offset) override;
  wabt::Result OnBinaryExpr(wabt::Opcode opcode) override;
  wabt::Result OnBlockExpr(Type sig_type) override;
  wabt::Result OnBrExpr(Index depth) override;
  wabt::Result OnBrIfExpr(Index depth) override;
  wabt::Result OnBrTableExpr(Index num_targets,
                             Index* target_depths,
                             Index default_target_depth) override;
  wabt::Result OnCallExpr(Index func_index) override;
  wabt::Result OnCallIndirectExpr(Index sig_index, Index table_index) override;
  wabt::Result OnReturnCallExpr(Index func_index) override;
  wabt::Result OnReturnCallIndirectExpr(Index sig_index, Index table_index) override;
  wabt::Result OnCompareExpr(wabt::Opcode opcode) override;
  wabt::Result OnConvertExpr(wabt::Opcode opcode) override;
  wabt::Result OnDropExpr() override;
  wabt::Result OnElseExpr() override;
  wabt::Result OnEndExpr() override;
  wabt::Result OnF32ConstExpr(uint32_t value_bits) override;
  wabt::Result OnF64ConstExpr(uint64_t value_bits) override;
  wabt::Result OnV128ConstExpr(v128 value_bits) override;
  wabt::Result OnGlobalGetExpr(Index global_index) override;
  wabt::Result OnGlobalSetExpr(Index global_index) override;
  wabt::Result OnI32ConstExpr(uint32_t value) override;
  wabt::Result OnI64ConstExpr(uint64_t value) override;
  wabt::Result OnIfExpr(Type sig_type) override;
  wabt::Result OnLoadExpr(wabt::Opcode opcode,
                          uint32_t alignment_log2,
                          Address offset) override;
  wabt::Result OnLocalGetExpr(Index local_index) override;
  wabt::Result OnLocalSetExpr(Index local_index) override;
  wabt::Result OnLocalTeeExpr(Index local_index) override;
  wabt::Result OnLoopExpr(Type sig_type) override;
  wabt::Result OnMemoryGrowExpr() override;
  wabt::Result OnMemorySizeExpr() override;
  wabt::Result OnNopExpr() override;
  wabt::Result OnReturnExpr() override;
  wabt::Result OnSelectExpr() override;
  wabt::Result OnStoreExpr(wabt::Opcode opcode,
                           uint32_t alignment_log2,
                           Address offset) override;
  wabt::Result OnUnaryExpr(wabt::Opcode opcode) override;
  wabt::Result OnTernaryExpr(wabt::Opcode opcode) override;
  wabt::Result OnUnreachableExpr() override;
  wabt::Result EndFunctionBody(Index index) override;
  wabt::Result OnSimdLaneOpExpr(wabt::Opcode opcode, uint64_t value) override;
  wabt::Result OnSimdShuffleOpExpr(wabt::Opcode opcode, v128 value) override;

  wabt::Result EndElemSegmentInitExpr(Index index) override;
  wabt::Result OnElemSegmentFunctionIndexCount(Index index,
                                               Index count) override;
  wabt::Result OnElemSegmentFunctionIndex(Index index,
                                          Index func_index) override;

  wabt::Result OnDataSegmentData(Index index,
                                 const void* data,
                                 Address size) override;

  wabt::Result OnInitExprF32ConstExpr(Index index, uint32_t value) override;
  wabt::Result OnInitExprF64ConstExpr(Index index, uint64_t value) override;
  wabt::Result OnInitExprV128ConstExpr(Index index, v128 value) override;
  wabt::Result OnInitExprGlobalGetExpr(Index index,
                                       Index global_index) override;
  wabt::Result OnInitExprI32ConstExpr(Index index, uint32_t value) override;
  wabt::Result OnInitExprI64ConstExpr(Index index, uint64_t value) override;

 private:
  Label* GetLabel(Index depth);
  Label* TopLabel();
  void PushLabel(IstreamOffset offset, IstreamOffset fixup_offset);
  void PopLabel();

  void PrintError(const char* format, ...);

  Index TranslateSigIndexToEnv(Index sig_index);
  void GetBlockSignature(Type sig_type,
                         TypeVector* out_param_types,
                         TypeVector* out_result_types);
  FuncSignature* GetSignatureByModuleIndex(Index sig_index);
  Index TranslateFuncIndexToEnv(Index func_index);
  Index TranslateModuleFuncIndexToDefined(Index func_index);
  Func* GetFuncByModuleIndex(Index func_index);
  Index TranslateGlobalIndexToEnv(Index global_index);
  Global* GetGlobalByModuleIndex(Index global_index);
  Type GetGlobalTypeByModuleIndex(Index global_index);
  Index TranslateLocalIndex(Index local_index);
  Type GetLocalTypeByIndex(Func* func, Index local_index);

  IstreamOffset GetIstreamOffset();

  wabt::Result EmitDataAt(IstreamOffset offset,
                          const void* data,
                          IstreamOffset size);
  wabt::Result EmitData(const void* data, IstreamOffset size);
  wabt::Result EmitOpcode(Opcode opcode);
  wabt::Result EmitI8(uint8_t value);
  wabt::Result EmitI32(uint32_t value);
  wabt::Result EmitI64(uint64_t value);
  wabt::Result EmitV128(v128 value);
  wabt::Result EmitI32At(IstreamOffset offset, uint32_t value);
  wabt::Result EmitDropKeep(uint32_t drop, uint32_t keep);
  wabt::Result AppendFixup(IstreamOffsetVectorVector* fixups_vector,
                           Index index);
  wabt::Result EmitBrOffset(Index depth, IstreamOffset offset);
  wabt::Result GetBrDropKeepCount(Index depth,
                                  Index* out_drop_count,
                                  Index* out_keep_count);
  wabt::Result GetReturnDropKeepCount(Index* out_drop_count,
                                      Index* out_keep_count);
  wabt::Result EmitBr(Index depth, Index drop_count, Index keep_count);
  wabt::Result EmitBrTableOffset(Index depth);
  wabt::Result FixupTopLabel();
  wabt::Result EmitFuncOffset(DefinedFunc* func, Index func_index);

  wabt::Result CheckLocal(Index local_index);
  wabt::Result CheckGlobal(Index global_index);
  wabt::Result CheckImportKind(Import* import, ExternalKind expected_kind);
  wabt::Result CheckImportLimits(const Limits* declared_limits,
                                 const Limits* actual_limits);
  wabt::Result CheckHasMemory(wabt::Opcode opcode);
  wabt::Result CheckAlign(uint32_t alignment_log2, Address natural_alignment);
  wabt::Result CheckAtomicAlign(uint32_t alignment_log2,
                                Address natural_alignment);

  wabt::Result AppendExport(Module* module,
                            ExternalKind kind,
                            Index item_index,
                            string_view name);
  wabt::Result FindRegisteredModule(string_view module_name,
                                    Module** out_module);
  wabt::Result GetModuleExport(Module* module,
                               string_view field_name,
                               Export** out_export);

  Features features_;
  Errors* errors_ = nullptr;
  Environment* env_ = nullptr;
  DefinedModule* module_ = nullptr;
  DefinedFunc* current_func_ = nullptr;
  TypeChecker typechecker_;
  std::vector<Label> label_stack_;
  IstreamOffsetVectorVector func_fixups_;
  IstreamOffsetVectorVector depth_fixups_;
  MemoryStream istream_;
  IstreamOffset istream_offset_ = 0;
  /* mappings from module index space to env index space; this won't just be a
   * translation, because imported values will be resolved as well */
  IndexVector sig_index_mapping_;
  IndexVector func_index_mapping_;
  IndexVector global_index_mapping_;

  Index num_func_imports_ = 0;
  Index num_global_imports_ = 0;

  // Changes to linear memory and tables should not apply if a validation error
  // occurs; these vectors cache the changes that must be applied after we know
  // that there are no validation errors.
  std::vector<ElemSegmentInfo> elem_segment_infos_;
  std::vector<DataSegmentInfo> data_segment_infos_;

  // Values cached so they can be shared between callbacks.
  TypedValue init_expr_value_;
  IstreamOffset table_offset_ = 0;
};

BinaryReaderInterp::BinaryReaderInterp(Environment* env,
                                       DefinedModule* module,
                                       std::unique_ptr<OutputBuffer> istream,
                                       Errors* errors,
                                       const Features& features)
    : features_(features),
      errors_(errors),
      env_(env),
      module_(module),
      istream_(std::move(istream)),
      istream_offset_(istream_.output_buffer().size()) {
  typechecker_.set_error_callback(
      [this](const char* msg) { PrintError("%s", msg); });
}

std::unique_ptr<OutputBuffer> BinaryReaderInterp::ReleaseOutputBuffer() {
  return istream_.ReleaseOutputBuffer();
}

Label* BinaryReaderInterp::GetLabel(Index depth) {
  assert(depth < label_stack_.size());
  return &label_stack_[label_stack_.size() - depth - 1];
}

Label* BinaryReaderInterp::TopLabel() {
  return GetLabel(0);
}

void WABT_PRINTF_FORMAT(2, 3) BinaryReaderInterp::PrintError(const char* format,
                                                             ...) {
  WABT_SNPRINTF_ALLOCA(buffer, length, format);
  errors_->emplace_back(ErrorLevel::Error, Location(kInvalidOffset), buffer);
}

Index BinaryReaderInterp::TranslateSigIndexToEnv(Index sig_index) {
  assert(sig_index < sig_index_mapping_.size());
  return sig_index_mapping_[sig_index];
}

void BinaryReaderInterp::GetBlockSignature(Type sig_type,
                                           TypeVector* out_param_types,
                                           TypeVector* out_result_types) {
  if (IsTypeIndex(sig_type)) {
    FuncSignature* func_sig = GetSignatureByModuleIndex(GetTypeIndex(sig_type));
    *out_param_types = func_sig->param_types;
    *out_result_types = func_sig->result_types;
  } else {
    out_param_types->clear();
    *out_result_types = GetInlineTypeVector(sig_type);
  }
}

FuncSignature* BinaryReaderInterp::GetSignatureByModuleIndex(Index sig_index) {
  return env_->GetFuncSignature(TranslateSigIndexToEnv(sig_index));
}

Index BinaryReaderInterp::TranslateFuncIndexToEnv(Index func_index) {
  assert(func_index < func_index_mapping_.size());
  return func_index_mapping_[func_index];
}

Index BinaryReaderInterp::TranslateModuleFuncIndexToDefined(Index func_index) {
  assert(func_index >= num_func_imports_);
  return func_index - num_func_imports_;
}

Func* BinaryReaderInterp::GetFuncByModuleIndex(Index func_index) {
  return env_->GetFunc(TranslateFuncIndexToEnv(func_index));
}

Index BinaryReaderInterp::TranslateGlobalIndexToEnv(Index global_index) {
  return global_index_mapping_[global_index];
}

Global* BinaryReaderInterp::GetGlobalByModuleIndex(Index global_index) {
  return env_->GetGlobal(TranslateGlobalIndexToEnv(global_index));
}

Type BinaryReaderInterp::GetGlobalTypeByModuleIndex(Index global_index) {
  return GetGlobalByModuleIndex(global_index)->typed_value.type;
}

Type BinaryReaderInterp::GetLocalTypeByIndex(Func* func, Index local_index) {
  assert(!func->is_host);
  return cast<DefinedFunc>(func)->param_and_local_types[local_index];
}

IstreamOffset BinaryReaderInterp::GetIstreamOffset() {
  return istream_offset_;
}

wabt::Result BinaryReaderInterp::EmitDataAt(IstreamOffset offset,
                                            const void* data,
                                            IstreamOffset size) {
  istream_.WriteDataAt(offset, data, size);
  return istream_.result();
}

wabt::Result BinaryReaderInterp::EmitData(const void* data,
                                          IstreamOffset size) {
  CHECK_RESULT(EmitDataAt(istream_offset_, data, size));
  istream_offset_ += size;
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::EmitOpcode(Opcode opcode) {
  return EmitI32(static_cast<uint32_t>(opcode));
}

wabt::Result BinaryReaderInterp::EmitI8(uint8_t value) {
  return EmitData(&value, sizeof(value));
}

wabt::Result BinaryReaderInterp::EmitI32(uint32_t value) {
  return EmitData(&value, sizeof(value));
}

wabt::Result BinaryReaderInterp::EmitI64(uint64_t value) {
  return EmitData(&value, sizeof(value));
}

wabt::Result BinaryReaderInterp::EmitV128(v128 value) {
  return EmitData(&value, sizeof(value));
}

wabt::Result BinaryReaderInterp::EmitI32At(IstreamOffset offset,
                                           uint32_t value) {
  return EmitDataAt(offset, &value, sizeof(value));
}

wabt::Result BinaryReaderInterp::EmitDropKeep(uint32_t drop, uint32_t keep) {
  assert(drop != UINT32_MAX);
  assert(keep != UINT32_MAX);
  if (drop > 0) {
    if (drop == 1 && keep == 0) {
      CHECK_RESULT(EmitOpcode(Opcode::Drop));
    } else {
      CHECK_RESULT(EmitOpcode(Opcode::InterpDropKeep));
      CHECK_RESULT(EmitI32(drop));
      CHECK_RESULT(EmitI32(keep));
    }
  }
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::AppendFixup(
    IstreamOffsetVectorVector* fixups_vector,
    Index index) {
  if (index >= fixups_vector->size()) {
    fixups_vector->resize(index + 1);
  }
  (*fixups_vector)[index].push_back(GetIstreamOffset());
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::EmitBrOffset(Index depth,
                                              IstreamOffset offset) {
  if (offset == kInvalidIstreamOffset) {
    /* depth_fixups_ stores the depth counting up from zero, where zero is the
     * top-level function scope. */
    depth = label_stack_.size() - 1 - depth;
    CHECK_RESULT(AppendFixup(&depth_fixups_, depth));
  }
  CHECK_RESULT(EmitI32(offset));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::GetBrDropKeepCount(Index depth,
                                                    Index* out_drop_count,
                                                    Index* out_keep_count) {
  TypeChecker::Label* label;
  CHECK_RESULT(typechecker_.GetLabel(depth, &label));
  *out_keep_count = label->br_types().size();
  if (typechecker_.IsUnreachable()) {
    *out_drop_count = 0;
  } else {
    *out_drop_count =
        (typechecker_.type_stack_size() - label->type_stack_limit) -
        *out_keep_count;
  }
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::GetReturnDropKeepCount(Index* out_drop_count,
                                                        Index* out_keep_count) {
  CHECK_RESULT(GetBrDropKeepCount(label_stack_.size() - 1, out_drop_count,
                                  out_keep_count));
  *out_drop_count += current_func_->param_and_local_types.size();
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::EmitBr(Index depth,
                                        Index drop_count,
                                        Index keep_count) {
  CHECK_RESULT(EmitDropKeep(drop_count, keep_count));
  CHECK_RESULT(EmitOpcode(Opcode::Br));
  CHECK_RESULT(EmitBrOffset(depth, GetLabel(depth)->offset));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::EmitBrTableOffset(Index depth) {
  Index drop_count, keep_count;
  CHECK_RESULT(GetBrDropKeepCount(depth, &drop_count, &keep_count));
  CHECK_RESULT(EmitBrOffset(depth, GetLabel(depth)->offset));
  CHECK_RESULT(EmitI32(drop_count));
  CHECK_RESULT(EmitI32(keep_count));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::FixupTopLabel() {
  IstreamOffset offset = GetIstreamOffset();
  Index top = label_stack_.size() - 1;
  if (top >= depth_fixups_.size()) {
    /* nothing to fixup */
    return wabt::Result::Ok;
  }

  IstreamOffsetVector& fixups = depth_fixups_[top];
  for (IstreamOffset fixup : fixups)
    CHECK_RESULT(EmitI32At(fixup, offset));
  fixups.clear();
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::EmitFuncOffset(DefinedFunc* func,
                                                Index func_index) {
  if (func->offset == kInvalidIstreamOffset) {
    Index defined_index = TranslateModuleFuncIndexToDefined(func_index);
    CHECK_RESULT(AppendFixup(&func_fixups_, defined_index));
  }
  CHECK_RESULT(EmitI32(func->offset));
  return wabt::Result::Ok;
}

bool BinaryReaderInterp::OnError(const Error& error) {
  errors_->push_back(error);
  return true;
}

wabt::Result BinaryReaderInterp::OnTypeCount(Index count) {
  Index sig_count = env_->GetFuncSignatureCount();
  sig_index_mapping_.resize(count);
  for (Index i = 0; i < count; ++i)
    sig_index_mapping_[i] = sig_count + i;
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnType(Index index,
                                        Index param_count,
                                        Type* param_types,
                                        Index result_count,
                                        Type* result_types) {
  assert(TranslateSigIndexToEnv(index) == env_->GetFuncSignatureCount());
  env_->EmplaceBackFuncSignature(param_count, param_types, result_count,
                                 result_types);
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::CheckLocal(Index local_index) {
  Index max_local_index = current_func_->param_and_local_types.size();
  if (local_index >= max_local_index) {
    PrintError("invalid local_index: %" PRIindex " (max %" PRIindex ")",
               local_index, max_local_index);
    return wabt::Result::Error;
  }
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::CheckGlobal(Index global_index) {
  Index max_global_index = global_index_mapping_.size();
  if (global_index >= max_global_index) {
    PrintError("invalid global_index: %" PRIindex " (max %" PRIindex ")",
               global_index, max_global_index);
    return wabt::Result::Error;
  }
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::CheckImportKind(Import* import,
                                                 ExternalKind actual_kind) {
  if (import->kind != actual_kind) {
    PrintError("expected import \"" PRIstringview "." PRIstringview
               "\" to have kind %s, not %s",
               WABT_PRINTF_STRING_VIEW_ARG(import->module_name),
               WABT_PRINTF_STRING_VIEW_ARG(import->field_name),
               GetKindName(import->kind), GetKindName(actual_kind));
    return wabt::Result::Error;
  }
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::CheckImportLimits(
    const Limits* declared_limits,
    const Limits* actual_limits) {
  if (actual_limits->initial < declared_limits->initial) {
    PrintError("actual size (%" PRIu64 ") smaller than declared (%" PRIu64 ")",
               actual_limits->initial, declared_limits->initial);
    return wabt::Result::Error;
  }

  if (declared_limits->has_max) {
    if (!actual_limits->has_max) {
      PrintError("max size (unspecified) larger than declared (%" PRIu64 ")",
                 declared_limits->max);
      return wabt::Result::Error;
    } else if (actual_limits->max > declared_limits->max) {
      PrintError("max size (%" PRIu64 ") larger than declared (%" PRIu64 ")",
                 actual_limits->max, declared_limits->max);
      return wabt::Result::Error;
    }
  }

  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::AppendExport(Module* module,
                                              ExternalKind kind,
                                              Index item_index,
                                              string_view name) {
  // Host modules are allowed to have duplicated exports; e.g. "spectest.print"
  if (isa<DefinedModule>(module) &&
      module->export_bindings.FindIndex(name) != kInvalidIndex) {
    PrintError("duplicate export \"" PRIstringview "\"",
               WABT_PRINTF_STRING_VIEW_ARG(name));
    return wabt::Result::Error;
  }

  module->AppendExport(kind, item_index, name);
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::FindRegisteredModule(string_view module_name,
                                                      Module** out_module) {
  Module* module = env_->FindRegisteredModule(module_name);
  if (!module) {
    PrintError("unknown import module \"" PRIstringview "\"",
               WABT_PRINTF_STRING_VIEW_ARG(module_name));
    return wabt::Result::Error;
  }

  *out_module = module;
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::GetModuleExport(Module* module,
                                                 string_view field_name,
                                                 Export** out_export) {
  Export* export_ = module->GetExport(field_name);
  if (!export_) {
    PrintError("unknown module field \"" PRIstringview "\"",
               WABT_PRINTF_STRING_VIEW_ARG(field_name));
    return wabt::Result::Error;
  }

  *out_export = export_;
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnImportFunc(Index import_index,
                                              string_view module_name,
                                              string_view field_name,
                                              Index func_index,
                                              Index sig_index) {
  module_->func_imports.emplace_back(module_name, field_name);
  FuncImport* import = &module_->func_imports.back();
  import->sig_index = TranslateSigIndexToEnv(sig_index);

  Module* import_module;
  CHECK_RESULT(FindRegisteredModule(import->module_name, &import_module));

  Export* export_ =
      import_module->GetFuncExport(env_, field_name, import->sig_index);
  if (!export_) {
    // If GetFuncExport fails then GetModuleExport will fail too. But it's
    // useful to call here to share the same error handling code as other
    // imports.
    CHECK_RESULT(GetModuleExport(import_module, import->field_name, &export_));
  }

  CHECK_RESULT(CheckImportKind(import, export_->kind));

  Func* func = env_->GetFunc(export_->index);
  if (!env_->FuncSignaturesAreEqual(import->sig_index, func->sig_index)) {
    PrintError("import signature mismatch");
    return wabt::Result::Error;
  }

  func_index_mapping_.push_back(export_->index);
  num_func_imports_++;
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnImportTable(Index import_index,
                                               string_view module_name,
                                               string_view field_name,
                                               Index table_index,
                                               Type elem_type,
                                               const Limits* elem_limits) {
  if (module_->table_index != kInvalidIndex) {
    PrintError("only one table allowed");
    return wabt::Result::Error;
  }

  module_->table_imports.emplace_back(module_name, field_name);
  TableImport* import = &module_->table_imports.back();

  Module* import_module;
  CHECK_RESULT(FindRegisteredModule(import->module_name, &import_module));

  Export* export_;
  CHECK_RESULT(GetModuleExport(import_module, import->field_name, &export_));
  CHECK_RESULT(CheckImportKind(import, export_->kind));

  Table* table = env_->GetTable(export_->index);
  CHECK_RESULT(CheckImportLimits(elem_limits, &table->limits));

  import->limits = *elem_limits;
  module_->table_index = export_->index;
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnImportMemory(Index import_index,
                                                string_view module_name,
                                                string_view field_name,
                                                Index memory_index,
                                                const Limits* page_limits) {
  if (module_->memory_index != kInvalidIndex) {
    PrintError("only one memory allowed");
    return wabt::Result::Error;
  }

  module_->memory_imports.emplace_back(module_name, field_name);
  MemoryImport* import = &module_->memory_imports.back();

  Module* import_module;
  CHECK_RESULT(FindRegisteredModule(import->module_name, &import_module));

  Export* export_;
  CHECK_RESULT(GetModuleExport(import_module, import->field_name, &export_));
  CHECK_RESULT(CheckImportKind(import, export_->kind));

  Memory* memory = env_->GetMemory(export_->index);
  CHECK_RESULT(CheckImportLimits(page_limits, &memory->page_limits));

  import->limits = *page_limits;
  module_->memory_index = export_->index;
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnImportGlobal(Index import_index,
                                                string_view module_name,
                                                string_view field_name,
                                                Index global_index,
                                                Type type,
                                                bool mutable_) {
  module_->global_imports.emplace_back(module_name, field_name);
  GlobalImport* import = &module_->global_imports.back();

  Module* import_module;
  CHECK_RESULT(FindRegisteredModule(import->module_name, &import_module));

  Export* export_;
  CHECK_RESULT(GetModuleExport(import_module, import->field_name, &export_));
  CHECK_RESULT(CheckImportKind(import, export_->kind));

  Global* exported_global = env_->GetGlobal(export_->index);
  if (exported_global->typed_value.type != type) {
    PrintError("type mismatch in imported global, expected %s but got %s.",
               GetTypeName(exported_global->typed_value.type),
               GetTypeName(type));
    return wabt::Result::Error;
  }

  if (exported_global->mutable_ != mutable_) {
    const char* kMutableNames[] = {"immutable", "mutable"};
    PrintError(
        "mutability mismatch in imported global, expected %s but got %s.",
        kMutableNames[exported_global->mutable_], kMutableNames[mutable_]);
    return wabt::Result::Error;
  }

  import->type = type;
  import->mutable_ = mutable_;
  global_index_mapping_.push_back(export_->index);
  num_global_imports_++;
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnFunctionCount(Index count) {
  for (Index i = 0; i < count; ++i)
    func_index_mapping_.push_back(env_->GetFuncCount() + i);
  func_fixups_.resize(count);
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnFunction(Index index, Index sig_index) {
  env_->EmplaceBackFunc(new DefinedFunc(TranslateSigIndexToEnv(sig_index)));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnTable(Index index,
                                         Type elem_type,
                                         const Limits* elem_limits) {
  if (module_->table_index != kInvalidIndex) {
    PrintError("only one table allowed");
    return wabt::Result::Error;
  }
  env_->EmplaceBackTable(elem_type, *elem_limits);
  module_->table_index = env_->GetTableCount() - 1;
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnMemory(Index index,
                                          const Limits* page_limits) {
  if (module_->memory_index != kInvalidIndex) {
    PrintError("only one memory allowed");
    return wabt::Result::Error;
  }
  env_->EmplaceBackMemory(*page_limits);
  module_->memory_index = env_->GetMemoryCount() - 1;
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnGlobalCount(Index count) {
  for (Index i = 0; i < count; ++i)
    global_index_mapping_.push_back(env_->GetGlobalCount() + i);
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::BeginGlobal(Index index,
                                             Type type,
                                             bool mutable_) {
  assert(TranslateGlobalIndexToEnv(index) == env_->GetGlobalCount());
  env_->EmplaceBackGlobal(TypedValue(type), mutable_);
  init_expr_value_.type = Type::Void;
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::EndGlobalInitExpr(Index index) {
  Global* global = GetGlobalByModuleIndex(index);
  if (init_expr_value_.type != global->typed_value.type) {
    PrintError("type mismatch in global, expected %s but got %s.",
               GetTypeName(global->typed_value.type),
               GetTypeName(init_expr_value_.type));
    return wabt::Result::Error;
  }
  global->typed_value = init_expr_value_;
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnInitExprF32ConstExpr(Index index,
                                                        uint32_t value_bits) {
  init_expr_value_.type = Type::F32;
  init_expr_value_.value.f32_bits = value_bits;
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnInitExprF64ConstExpr(Index index,
                                                        uint64_t value_bits) {
  init_expr_value_.type = Type::F64;
  init_expr_value_.value.f64_bits = value_bits;
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnInitExprV128ConstExpr(Index index,
                                                         v128 value_bits) {
  init_expr_value_.type = Type::V128;
  init_expr_value_.value.v128_bits = value_bits;
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnInitExprGlobalGetExpr(Index index,
                                                         Index global_index) {
  if (global_index >= num_global_imports_) {
    PrintError("initializer expression can only reference an imported global");
    return wabt::Result::Error;
  }
  Global* ref_global = GetGlobalByModuleIndex(global_index);
  if (ref_global->mutable_) {
    PrintError("initializer expression cannot reference a mutable global");
    return wabt::Result::Error;
  }
  init_expr_value_ = ref_global->typed_value;
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnInitExprI32ConstExpr(Index index,
                                                        uint32_t value) {
  init_expr_value_.type = Type::I32;
  init_expr_value_.value.i32 = value;
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnInitExprI64ConstExpr(Index index,
                                                        uint64_t value) {
  init_expr_value_.type = Type::I64;
  init_expr_value_.value.i64 = value;
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnExport(Index index,
                                          ExternalKind kind,
                                          Index item_index,
                                          string_view name) {
  switch (kind) {
    case ExternalKind::Func:
      item_index = TranslateFuncIndexToEnv(item_index);
      break;

    case ExternalKind::Table:
      item_index = module_->table_index;
      break;

    case ExternalKind::Memory:
      item_index = module_->memory_index;
      break;

    case ExternalKind::Global: {
      item_index = TranslateGlobalIndexToEnv(item_index);
      Global* global = env_->GetGlobal(item_index);
      if (global->mutable_ && !features_.mutable_globals_enabled()) {
        PrintError("mutable globals cannot be exported");
        return wabt::Result::Error;
      }
      break;
    }

    case ExternalKind::Event:
      // TODO(karlschimpf) Define
      WABT_FATAL("BinaryReaderInterp::OnExport(event) not implemented");
      break;
  }
  return AppendExport(module_, kind, item_index, name);
}

wabt::Result BinaryReaderInterp::OnStartFunction(Index func_index) {
  Index start_func_index = TranslateFuncIndexToEnv(func_index);
  Func* start_func = env_->GetFunc(start_func_index);
  FuncSignature* sig = env_->GetFuncSignature(start_func->sig_index);
  if (sig->param_types.size() != 0) {
    PrintError("start function must be nullary");
    return wabt::Result::Error;
  }
  if (sig->result_types.size() != 0) {
    PrintError("start function must not return anything");
    return wabt::Result::Error;
  }
  module_->start_func_index = start_func_index;
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::EndElemSegmentInitExpr(Index index) {
  assert(init_expr_value_.type == Type::I32);
  table_offset_ = init_expr_value_.value.i32;
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnElemSegmentFunctionIndexCount(Index index,
                                                                 Index count) {
  assert(module_->table_index != kInvalidIndex);
  Table* table = env_->GetTable(module_->table_index);
  // Check both cases, as table_offset_ + count may overflow.
  if (table_offset_ > table->func_indexes.size() ||
      table_offset_ + count > table->func_indexes.size()) {
    PrintError("elem segment is out of bounds: [%u, %u) >= max value %" PRIzd,
               table_offset_, table_offset_ + count,
               table->func_indexes.size());
    return wabt::Result::Error;
  }
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnElemSegmentFunctionIndex(Index index,
                                                            Index func_index) {
  Index max_func_index = func_index_mapping_.size();
  if (func_index >= max_func_index) {
    PrintError("invalid func_index: %" PRIindex " (max %" PRIindex ")",
               func_index, max_func_index);
    return wabt::Result::Error;
  }

  Table* table = env_->GetTable(module_->table_index);
  elem_segment_infos_.emplace_back(&table->func_indexes[table_offset_++],
                                   TranslateFuncIndexToEnv(func_index));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnDataSegmentData(Index index,
                                                   const void* src_data,
                                                   Address size) {
  assert(module_->memory_index != kInvalidIndex);
  Memory* memory = env_->GetMemory(module_->memory_index);
  assert(init_expr_value_.type == Type::I32);
  Address address = init_expr_value_.value.i32;
  uint64_t end_address =
      static_cast<uint64_t>(address) + static_cast<uint64_t>(size);
  if (end_address > memory->data.size()) {
    PrintError("data segment is out of bounds: [%" PRIaddress ", %" PRIu64
               ") >= max value %" PRIzd,
               address, end_address, memory->data.size());
    return wabt::Result::Error;
  }

  if (size > 0) {
    data_segment_infos_.emplace_back(&memory->data[address], src_data, size);
  }

  return wabt::Result::Ok;
}

void BinaryReaderInterp::PushLabel(IstreamOffset offset,
                                   IstreamOffset fixup_offset) {
  label_stack_.emplace_back(offset, fixup_offset);
}

void BinaryReaderInterp::PopLabel() {
  label_stack_.pop_back();
  /* reduce the depth_fixups_ stack as well, but it may be smaller than
   * label_stack_ so only do it conditionally. */
  if (depth_fixups_.size() > label_stack_.size()) {
    depth_fixups_.erase(depth_fixups_.begin() + label_stack_.size(),
                        depth_fixups_.end());
  }
}

wabt::Result BinaryReaderInterp::BeginFunctionBody(Index index, Offset size) {
  auto* func = cast<DefinedFunc>(GetFuncByModuleIndex(index));
  FuncSignature* sig = env_->GetFuncSignature(func->sig_index);

  func->offset = GetIstreamOffset();
  func->local_decl_count = 0;
  func->local_count = 0;

  current_func_ = func;
  depth_fixups_.clear();
  label_stack_.clear();

  /* fixup function references */
  Index defined_index = TranslateModuleFuncIndexToDefined(index);
  IstreamOffsetVector& fixups = func_fixups_[defined_index];
  for (IstreamOffset fixup : fixups)
    CHECK_RESULT(EmitI32At(fixup, func->offset));

  /* append param types */
  for (Type param_type : sig->param_types)
    func->param_and_local_types.push_back(param_type);

  CHECK_RESULT(typechecker_.BeginFunction(sig->result_types));

  /* push implicit func label (equivalent to return) */
  PushLabel(kInvalidIstreamOffset, kInvalidIstreamOffset);
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::EndFunctionBody(Index index) {
  FixupTopLabel();
  Index drop_count, keep_count;
  CHECK_RESULT(GetReturnDropKeepCount(&drop_count, &keep_count));
  CHECK_RESULT(typechecker_.EndFunction());
  CHECK_RESULT(EmitDropKeep(drop_count, keep_count));
  CHECK_RESULT(EmitOpcode(Opcode::Return));
  PopLabel();
  current_func_ = nullptr;
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnLocalDeclCount(Index count) {
  current_func_->local_decl_count = count;
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnLocalDecl(Index decl_index,
                                             Index count,
                                             Type type) {
  current_func_->local_count += count;

  for (Index i = 0; i < count; ++i)
    current_func_->param_and_local_types.push_back(type);

  if (decl_index == current_func_->local_decl_count - 1) {
    /* last local declaration, allocate space for all locals. */
    CHECK_RESULT(EmitOpcode(Opcode::InterpAlloca));
    CHECK_RESULT(EmitI32(current_func_->local_count));
  }
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::CheckHasMemory(wabt::Opcode opcode) {
  if (module_->memory_index == kInvalidIndex) {
    PrintError("%s requires an imported or defined memory.", opcode.GetName());
    return wabt::Result::Error;
  }
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::CheckAlign(uint32_t alignment_log2,
                                            Address natural_alignment) {
  if (alignment_log2 >= 32 || (1U << alignment_log2) > natural_alignment) {
    PrintError("alignment must not be larger than natural alignment (%u)",
               natural_alignment);
    return wabt::Result::Error;
  }
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::CheckAtomicAlign(uint32_t alignment_log2,
                                                  Address natural_alignment) {
  if (alignment_log2 >= 32 || (1U << alignment_log2) != natural_alignment) {
    PrintError("alignment must be equal to natural alignment (%u)",
               natural_alignment);
    return wabt::Result::Error;
  }
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnUnaryExpr(wabt::Opcode opcode) {
  CHECK_RESULT(typechecker_.OnUnary(opcode));
  CHECK_RESULT(EmitOpcode(opcode));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnTernaryExpr(wabt::Opcode opcode) {
  CHECK_RESULT(typechecker_.OnTernary(opcode));
  CHECK_RESULT(EmitOpcode(opcode));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnSimdLaneOpExpr(wabt::Opcode opcode,
                                                  uint64_t value) {
  CHECK_RESULT(typechecker_.OnSimdLaneOp(opcode, value));
  CHECK_RESULT(EmitOpcode(opcode));
  CHECK_RESULT(EmitI8(static_cast<uint8_t>(value)));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnSimdShuffleOpExpr(wabt::Opcode opcode,
                                                     v128 value) {
  CHECK_RESULT(typechecker_.OnSimdShuffleOp(opcode, value));
  CHECK_RESULT(EmitOpcode(opcode));
  CHECK_RESULT(EmitV128(value));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnAtomicLoadExpr(Opcode opcode,
                                                  uint32_t alignment_log2,
                                                  Address offset) {
  CHECK_RESULT(CheckHasMemory(opcode));
  CHECK_RESULT(CheckAtomicAlign(alignment_log2, opcode.GetMemorySize()));
  CHECK_RESULT(typechecker_.OnAtomicLoad(opcode));
  CHECK_RESULT(EmitOpcode(opcode));
  CHECK_RESULT(EmitI32(module_->memory_index));
  CHECK_RESULT(EmitI32(offset));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnAtomicStoreExpr(Opcode opcode,
                                                   uint32_t alignment_log2,
                                                   Address offset) {
  CHECK_RESULT(CheckHasMemory(opcode));
  CHECK_RESULT(CheckAtomicAlign(alignment_log2, opcode.GetMemorySize()));
  CHECK_RESULT(typechecker_.OnAtomicStore(opcode));
  CHECK_RESULT(EmitOpcode(opcode));
  CHECK_RESULT(EmitI32(module_->memory_index));
  CHECK_RESULT(EmitI32(offset));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnAtomicRmwExpr(Opcode opcode,
                                                 uint32_t alignment_log2,
                                                 Address offset) {
  CHECK_RESULT(CheckHasMemory(opcode));
  CHECK_RESULT(CheckAtomicAlign(alignment_log2, opcode.GetMemorySize()));
  CHECK_RESULT(typechecker_.OnAtomicRmw(opcode));
  CHECK_RESULT(EmitOpcode(opcode));
  CHECK_RESULT(EmitI32(module_->memory_index));
  CHECK_RESULT(EmitI32(offset));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnAtomicRmwCmpxchgExpr(Opcode opcode,
                                                        uint32_t alignment_log2,
                                                        Address offset) {
  CHECK_RESULT(CheckHasMemory(opcode));
  CHECK_RESULT(CheckAtomicAlign(alignment_log2, opcode.GetMemorySize()));
  CHECK_RESULT(typechecker_.OnAtomicRmwCmpxchg(opcode));
  CHECK_RESULT(EmitOpcode(opcode));
  CHECK_RESULT(EmitI32(module_->memory_index));
  CHECK_RESULT(EmitI32(offset));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnBinaryExpr(wabt::Opcode opcode) {
  CHECK_RESULT(typechecker_.OnBinary(opcode));
  CHECK_RESULT(EmitOpcode(opcode));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnBlockExpr(Type sig_type) {
  TypeVector param_types, result_types;
  GetBlockSignature(sig_type, &param_types, &result_types);
  CHECK_RESULT(typechecker_.OnBlock(param_types, result_types));
  PushLabel(kInvalidIstreamOffset, kInvalidIstreamOffset);
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnLoopExpr(Type sig_type) {
  TypeVector param_types, result_types;
  GetBlockSignature(sig_type, &param_types, &result_types);
  CHECK_RESULT(typechecker_.OnLoop(param_types, result_types));
  PushLabel(GetIstreamOffset(), kInvalidIstreamOffset);
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnIfExpr(Type sig_type) {
  TypeVector param_types, result_types;
  GetBlockSignature(sig_type, &param_types, &result_types);
  CHECK_RESULT(typechecker_.OnIf(param_types, result_types));
  CHECK_RESULT(EmitOpcode(Opcode::InterpBrUnless));
  IstreamOffset fixup_offset = GetIstreamOffset();
  CHECK_RESULT(EmitI32(kInvalidIstreamOffset));
  PushLabel(kInvalidIstreamOffset, fixup_offset);
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnElseExpr() {
  CHECK_RESULT(typechecker_.OnElse());
  Label* label = TopLabel();
  IstreamOffset fixup_cond_offset = label->fixup_offset;
  CHECK_RESULT(EmitOpcode(Opcode::Br));
  label->fixup_offset = GetIstreamOffset();
  CHECK_RESULT(EmitI32(kInvalidIstreamOffset));
  CHECK_RESULT(EmitI32At(fixup_cond_offset, GetIstreamOffset()));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnEndExpr() {
  TypeChecker::Label* label;
  CHECK_RESULT(typechecker_.GetLabel(0, &label));
  LabelType label_type = label->label_type;
  CHECK_RESULT(typechecker_.OnEnd());
  if (label_type == LabelType::If || label_type == LabelType::Else) {
    CHECK_RESULT(EmitI32At(TopLabel()->fixup_offset, GetIstreamOffset()));
  }
  FixupTopLabel();
  PopLabel();
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnBrExpr(Index depth) {
  Index drop_count, keep_count;
  CHECK_RESULT(GetBrDropKeepCount(depth, &drop_count, &keep_count));
  CHECK_RESULT(typechecker_.OnBr(depth));
  CHECK_RESULT(EmitBr(depth, drop_count, keep_count));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnBrIfExpr(Index depth) {
  Index drop_count, keep_count;
  CHECK_RESULT(typechecker_.OnBrIf(depth));
  CHECK_RESULT(GetBrDropKeepCount(depth, &drop_count, &keep_count));
  /* flip the br_if so if <cond> is true it can drop values from the stack */
  CHECK_RESULT(EmitOpcode(Opcode::InterpBrUnless));
  IstreamOffset fixup_br_offset = GetIstreamOffset();
  CHECK_RESULT(EmitI32(kInvalidIstreamOffset));
  CHECK_RESULT(EmitBr(depth, drop_count, keep_count));
  CHECK_RESULT(EmitI32At(fixup_br_offset, GetIstreamOffset()));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnBrTableExpr(Index num_targets,
                                               Index* target_depths,
                                               Index default_target_depth) {
  CHECK_RESULT(typechecker_.BeginBrTable());
  CHECK_RESULT(EmitOpcode(Opcode::BrTable));
  CHECK_RESULT(EmitI32(num_targets));
  IstreamOffset fixup_table_offset = GetIstreamOffset();
  CHECK_RESULT(EmitI32(kInvalidIstreamOffset));
  /* not necessary for the interp, but it makes it easier to disassemble.
   * This opcode specifies how many bytes of data follow. */
  CHECK_RESULT(EmitOpcode(Opcode::InterpData));
  CHECK_RESULT(EmitI32((num_targets + 1) * WABT_TABLE_ENTRY_SIZE));
  CHECK_RESULT(EmitI32At(fixup_table_offset, GetIstreamOffset()));

  for (Index i = 0; i <= num_targets; ++i) {
    Index depth = i != num_targets ? target_depths[i] : default_target_depth;
    CHECK_RESULT(typechecker_.OnBrTableTarget(depth));
    CHECK_RESULT(EmitBrTableOffset(depth));
  }

  CHECK_RESULT(typechecker_.EndBrTable());
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnCallExpr(Index func_index) {
  Func* func = GetFuncByModuleIndex(func_index);
  FuncSignature* sig = env_->GetFuncSignature(func->sig_index);
  CHECK_RESULT(typechecker_.OnCall(sig->param_types, sig->result_types));

  if (func->is_host) {
    CHECK_RESULT(EmitOpcode(Opcode::InterpCallHost));
    CHECK_RESULT(EmitI32(TranslateFuncIndexToEnv(func_index)));
  } else {
    CHECK_RESULT(EmitOpcode(Opcode::Call));
    CHECK_RESULT(EmitFuncOffset(cast<DefinedFunc>(func), func_index));
  }

  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnCallIndirectExpr(Index sig_index, Index table_index) {
  if (module_->table_index == kInvalidIndex) {
    PrintError("found call_indirect operator, but no table");
    return wabt::Result::Error;
  }
  FuncSignature* sig = GetSignatureByModuleIndex(sig_index);
  CHECK_RESULT(
      typechecker_.OnCallIndirect(sig->param_types, sig->result_types));

  CHECK_RESULT(EmitOpcode(Opcode::CallIndirect));
  CHECK_RESULT(EmitI32(module_->table_index));
  CHECK_RESULT(EmitI32(TranslateSigIndexToEnv(sig_index)));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnReturnCallExpr(Index func_index) {
  Func* func = GetFuncByModuleIndex(func_index);
  FuncSignature* sig = env_->GetFuncSignature(func->sig_index);
  CHECK_RESULT(typechecker_.OnReturnCall(sig->param_types, sig->result_types));

  Index drop_count, keep_count;
  CHECK_RESULT(GetReturnDropKeepCount(&drop_count, &keep_count));

  keep_count = static_cast<Index>(sig->param_types.size());

  CHECK_RESULT(EmitDropKeep(drop_count, keep_count));

  if (func->is_host) {
    CHECK_RESULT(EmitOpcode(Opcode::InterpCallHost));
    CHECK_RESULT(EmitI32(TranslateFuncIndexToEnv(func_index)));
    CHECK_RESULT(EmitOpcode(Opcode::Return));
  } else {
    CHECK_RESULT(EmitOpcode(Opcode::ReturnCall));
    CHECK_RESULT(EmitFuncOffset(cast<DefinedFunc>(func), func_index));
  }

  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnReturnCallIndirectExpr(Index sig_index, Index table_index) {
  if (module_->table_index == kInvalidIndex) {
    PrintError("found return_call_indirect operator, but no table");
    return wabt::Result::Error;
  }
  FuncSignature* sig = GetSignatureByModuleIndex(sig_index);
  CHECK_RESULT(
    typechecker_.OnReturnCallIndirect(sig->param_types, sig->result_types));

  Index drop_count, keep_count;
  CHECK_RESULT(GetReturnDropKeepCount(&drop_count, &keep_count));

  keep_count = static_cast<Index>(sig->param_types.size()+1); // Include the index of the function

  CHECK_RESULT(EmitDropKeep(drop_count, keep_count));

  CHECK_RESULT(EmitOpcode(Opcode::ReturnCallIndirect));
  CHECK_RESULT(EmitI32(module_->table_index));
  CHECK_RESULT(EmitI32(TranslateSigIndexToEnv(sig_index)));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnCompareExpr(wabt::Opcode opcode) {
  return OnBinaryExpr(opcode);
}

wabt::Result BinaryReaderInterp::OnConvertExpr(wabt::Opcode opcode) {
  return OnUnaryExpr(opcode);
}

wabt::Result BinaryReaderInterp::OnDropExpr() {
  CHECK_RESULT(typechecker_.OnDrop());
  CHECK_RESULT(EmitOpcode(Opcode::Drop));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnI32ConstExpr(uint32_t value) {
  CHECK_RESULT(typechecker_.OnConst(Type::I32));
  CHECK_RESULT(EmitOpcode(Opcode::I32Const));
  CHECK_RESULT(EmitI32(value));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnI64ConstExpr(uint64_t value) {
  CHECK_RESULT(typechecker_.OnConst(Type::I64));
  CHECK_RESULT(EmitOpcode(Opcode::I64Const));
  CHECK_RESULT(EmitI64(value));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnF32ConstExpr(uint32_t value_bits) {
  CHECK_RESULT(typechecker_.OnConst(Type::F32));
  CHECK_RESULT(EmitOpcode(Opcode::F32Const));
  CHECK_RESULT(EmitI32(value_bits));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnF64ConstExpr(uint64_t value_bits) {
  CHECK_RESULT(typechecker_.OnConst(Type::F64));
  CHECK_RESULT(EmitOpcode(Opcode::F64Const));
  CHECK_RESULT(EmitI64(value_bits));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnV128ConstExpr(v128 value_bits) {
  CHECK_RESULT(typechecker_.OnConst(Type::V128));
  CHECK_RESULT(EmitOpcode(Opcode::V128Const));
  CHECK_RESULT(EmitV128(value_bits));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnGlobalGetExpr(Index global_index) {
  CHECK_RESULT(CheckGlobal(global_index));
  Type type = GetGlobalTypeByModuleIndex(global_index);
  CHECK_RESULT(typechecker_.OnGlobalGet(type));
  CHECK_RESULT(EmitOpcode(Opcode::GlobalGet));
  CHECK_RESULT(EmitI32(TranslateGlobalIndexToEnv(global_index)));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnGlobalSetExpr(Index global_index) {
  CHECK_RESULT(CheckGlobal(global_index));
  Global* global = GetGlobalByModuleIndex(global_index);
  if (!global->mutable_) {
    PrintError("can't global.set on immutable global at index %" PRIindex ".",
               global_index);
    return wabt::Result::Error;
  }
  CHECK_RESULT(typechecker_.OnGlobalSet(global->typed_value.type));
  CHECK_RESULT(EmitOpcode(Opcode::GlobalSet));
  CHECK_RESULT(EmitI32(TranslateGlobalIndexToEnv(global_index)));
  return wabt::Result::Ok;
}

Index BinaryReaderInterp::TranslateLocalIndex(Index local_index) {
  return typechecker_.type_stack_size() +
         current_func_->param_and_local_types.size() - local_index;
}

wabt::Result BinaryReaderInterp::OnLocalGetExpr(Index local_index) {
  CHECK_RESULT(CheckLocal(local_index));
  Type type = GetLocalTypeByIndex(current_func_, local_index);
  // Get the translated index before calling typechecker_.OnLocalGet because it
  // will update the type stack size. We need the index to be relative to the
  // old stack size.
  Index translated_local_index = TranslateLocalIndex(local_index);
  CHECK_RESULT(typechecker_.OnLocalGet(type));
  CHECK_RESULT(EmitOpcode(Opcode::LocalGet));
  CHECK_RESULT(EmitI32(translated_local_index));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnLocalSetExpr(Index local_index) {
  CHECK_RESULT(CheckLocal(local_index));
  Type type = GetLocalTypeByIndex(current_func_, local_index);
  CHECK_RESULT(typechecker_.OnLocalSet(type));
  CHECK_RESULT(EmitOpcode(Opcode::LocalSet));
  CHECK_RESULT(EmitI32(TranslateLocalIndex(local_index)));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnLocalTeeExpr(Index local_index) {
  CHECK_RESULT(CheckLocal(local_index));
  Type type = GetLocalTypeByIndex(current_func_, local_index);
  CHECK_RESULT(typechecker_.OnLocalTee(type));
  CHECK_RESULT(EmitOpcode(Opcode::LocalTee));
  CHECK_RESULT(EmitI32(TranslateLocalIndex(local_index)));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnLoadExpr(wabt::Opcode opcode,
                                            uint32_t alignment_log2,
                                            Address offset) {
  CHECK_RESULT(CheckHasMemory(opcode));
  CHECK_RESULT(CheckAlign(alignment_log2, opcode.GetMemorySize()));
  CHECK_RESULT(typechecker_.OnLoad(opcode));
  CHECK_RESULT(EmitOpcode(opcode));
  CHECK_RESULT(EmitI32(module_->memory_index));
  CHECK_RESULT(EmitI32(offset));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnStoreExpr(wabt::Opcode opcode,
                                             uint32_t alignment_log2,
                                             Address offset) {
  CHECK_RESULT(CheckHasMemory(opcode));
  CHECK_RESULT(CheckAlign(alignment_log2, opcode.GetMemorySize()));
  CHECK_RESULT(typechecker_.OnStore(opcode));
  CHECK_RESULT(EmitOpcode(opcode));
  CHECK_RESULT(EmitI32(module_->memory_index));
  CHECK_RESULT(EmitI32(offset));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnMemoryGrowExpr() {
  CHECK_RESULT(CheckHasMemory(wabt::Opcode::MemoryGrow));
  CHECK_RESULT(typechecker_.OnMemoryGrow());
  CHECK_RESULT(EmitOpcode(Opcode::MemoryGrow));
  CHECK_RESULT(EmitI32(module_->memory_index));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnMemorySizeExpr() {
  CHECK_RESULT(CheckHasMemory(wabt::Opcode::MemorySize));
  CHECK_RESULT(typechecker_.OnMemorySize());
  CHECK_RESULT(EmitOpcode(Opcode::MemorySize));
  CHECK_RESULT(EmitI32(module_->memory_index));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnNopExpr() {
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnReturnExpr() {
  Index drop_count, keep_count;
  CHECK_RESULT(GetReturnDropKeepCount(&drop_count, &keep_count));
  CHECK_RESULT(typechecker_.OnReturn());
  CHECK_RESULT(EmitDropKeep(drop_count, keep_count));
  CHECK_RESULT(EmitOpcode(Opcode::Return));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnSelectExpr() {
  CHECK_RESULT(typechecker_.OnSelect());
  CHECK_RESULT(EmitOpcode(Opcode::Select));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnUnreachableExpr() {
  CHECK_RESULT(typechecker_.OnUnreachable());
  CHECK_RESULT(EmitOpcode(Opcode::Unreachable));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnAtomicWaitExpr(Opcode opcode,
                                                  uint32_t alignment_log2,
                                                  Address offset) {
  CHECK_RESULT(CheckHasMemory(opcode));
  CHECK_RESULT(CheckAtomicAlign(alignment_log2, opcode.GetMemorySize()));
  CHECK_RESULT(typechecker_.OnAtomicWait(opcode));
  CHECK_RESULT(EmitOpcode(opcode));
  CHECK_RESULT(EmitI32(module_->memory_index));
  CHECK_RESULT(EmitI32(offset));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::OnAtomicNotifyExpr(Opcode opcode,
                                                    uint32_t alignment_log2,
                                                    Address offset) {
  CHECK_RESULT(CheckHasMemory(opcode));
  CHECK_RESULT(CheckAtomicAlign(alignment_log2, opcode.GetMemorySize()));
  CHECK_RESULT(typechecker_.OnAtomicNotify(opcode));
  CHECK_RESULT(EmitOpcode(opcode));
  CHECK_RESULT(EmitI32(module_->memory_index));
  CHECK_RESULT(EmitI32(offset));
  return wabt::Result::Ok;
}

wabt::Result BinaryReaderInterp::EndModule() {
  for (ElemSegmentInfo& info : elem_segment_infos_) {
    *info.dst = info.func_index;
  }
  for (DataSegmentInfo& info : data_segment_infos_) {
    memcpy(info.dst_data, info.src_data, info.size);
  }
  return wabt::Result::Ok;
}

}  // end anonymous namespace

wabt::Result ReadBinaryInterp(Environment* env,
                              const void* data,
                              size_t size,
                              const ReadBinaryOptions& options,
                              Errors* errors,
                              DefinedModule** out_module) {
  // Need to mark before taking ownership of env->istream.
  Environment::MarkPoint mark = env->Mark();

  std::unique_ptr<OutputBuffer> istream = env->ReleaseIstream();
  IstreamOffset istream_offset = istream->size();
  DefinedModule* module = new DefinedModule();

  BinaryReaderInterp reader(env, module, std::move(istream), errors,
                            options.features);
  env->EmplaceBackModule(module);

  wabt::Result result = ReadBinary(data, size, &reader, options);
  env->SetIstream(reader.ReleaseOutputBuffer());

  if (Succeeded(result)) {
    module->istream_start = istream_offset;
    module->istream_end = env->istream().size();
    *out_module = module;
  } else {
    env->ResetToMarkPoint(mark);
    *out_module = nullptr;
  }
  return result;
}

}  // namespace wabt
