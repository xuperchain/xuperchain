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

#include "src/binary-reader-ir.h"

#include <cassert>
#include <cinttypes>
#include <cstdarg>
#include <cstdint>
#include <cstdio>
#include <vector>

#include "src/binary-reader-nop.h"
#include "src/cast.h"
#include "src/common.h"
#include "src/ir.h"

namespace wabt {

namespace {

struct LabelNode {
  LabelNode(LabelType, ExprList* exprs, Expr* context = nullptr);

  LabelType label_type;
  ExprList* exprs;
  Expr* context;
};

LabelNode::LabelNode(LabelType label_type, ExprList* exprs, Expr* context)
    : label_type(label_type), exprs(exprs), context(context) {}

class BinaryReaderIR : public BinaryReaderNop {
 public:
  BinaryReaderIR(Module* out_module,
                 const char* filename,
                 Errors* errors);

  bool OnError(const Error&) override;

  Result OnTypeCount(Index count) override;
  Result OnType(Index index,
                Index param_count,
                Type* param_types,
                Index result_count,
                Type* result_types) override;

  Result OnImportCount(Index count) override;
  Result OnImportFunc(Index import_index,
                      string_view module_name,
                      string_view field_name,
                      Index func_index,
                      Index sig_index) override;
  Result OnImportTable(Index import_index,
                       string_view module_name,
                       string_view field_name,
                       Index table_index,
                       Type elem_type,
                       const Limits* elem_limits) override;
  Result OnImportMemory(Index import_index,
                        string_view module_name,
                        string_view field_name,
                        Index memory_index,
                        const Limits* page_limits) override;
  Result OnImportGlobal(Index import_index,
                        string_view module_name,
                        string_view field_name,
                        Index global_index,
                        Type type,
                        bool mutable_) override;
  Result OnImportEvent(Index import_index,
                       string_view module_name,
                       string_view field_name,
                       Index event_index,
                       Index sig_index) override;

  Result OnFunctionCount(Index count) override;
  Result OnFunction(Index index, Index sig_index) override;

  Result OnTableCount(Index count) override;
  Result OnTable(Index index,
                 Type elem_type,
                 const Limits* elem_limits) override;

  Result OnMemoryCount(Index count) override;
  Result OnMemory(Index index, const Limits* limits) override;

  Result OnGlobalCount(Index count) override;
  Result BeginGlobal(Index index, Type type, bool mutable_) override;
  Result BeginGlobalInitExpr(Index index) override;
  Result EndGlobalInitExpr(Index index) override;

  Result OnExportCount(Index count) override;
  Result OnExport(Index index,
                  ExternalKind kind,
                  Index item_index,
                  string_view name) override;

  Result OnStartFunction(Index func_index) override;

  Result OnFunctionBodyCount(Index count) override;
  Result BeginFunctionBody(Index index, Offset size) override;
  Result OnLocalDecl(Index decl_index, Index count, Type type) override;

  Result OnAtomicLoadExpr(Opcode opcode,
                          uint32_t alignment_log2,
                          Address offset) override;
  Result OnAtomicStoreExpr(Opcode opcode,
                           uint32_t alignment_log2,
                           Address offset) override;
  Result OnAtomicRmwExpr(Opcode opcode,
                         uint32_t alignment_log2,
                         Address offset) override;
  Result OnAtomicRmwCmpxchgExpr(Opcode opcode,
                                uint32_t alignment_log2,
                                Address offset) override;
  wabt::Result OnAtomicWaitExpr(Opcode opcode,
                                uint32_t alignment_log2,
                                Address offset) override;
  wabt::Result OnAtomicNotifyExpr(Opcode opcode,
                                  uint32_t alignment_log2,
                                  Address offset) override;
  Result OnBinaryExpr(Opcode opcode) override;
  Result OnBlockExpr(Type sig_type) override;
  Result OnBrExpr(Index depth) override;
  Result OnBrIfExpr(Index depth) override;
  Result OnBrOnExnExpr(Index depth, Index event_index) override;
  Result OnBrTableExpr(Index num_targets,
                       Index* target_depths,
                       Index default_target_depth) override;
  Result OnCallExpr(Index func_index) override;
  Result OnCatchExpr() override;
  Result OnCallIndirectExpr(Index sig_index, Index table_index) override;
  Result OnReturnCallExpr(Index func_index) override;
  Result OnReturnCallIndirectExpr(Index sig_index, Index table_index) override;
  Result OnCompareExpr(Opcode opcode) override;
  Result OnConvertExpr(Opcode opcode) override;
  Result OnDropExpr() override;
  Result OnElseExpr() override;
  Result OnEndExpr() override;
  Result OnF32ConstExpr(uint32_t value_bits) override;
  Result OnF64ConstExpr(uint64_t value_bits) override;
  Result OnV128ConstExpr(v128 value_bits) override;
  Result OnGlobalGetExpr(Index global_index) override;
  Result OnGlobalSetExpr(Index global_index) override;
  Result OnI32ConstExpr(uint32_t value) override;
  Result OnI64ConstExpr(uint64_t value) override;
  Result OnIfExpr(Type sig_type) override;
  Result OnLoadExpr(Opcode opcode,
                    uint32_t alignment_log2,
                    Address offset) override;
  Result OnLocalGetExpr(Index local_index) override;
  Result OnLocalSetExpr(Index local_index) override;
  Result OnLocalTeeExpr(Index local_index) override;
  Result OnLoopExpr(Type sig_type) override;
  Result OnMemoryCopyExpr() override;
  Result OnDataDropExpr(Index segment_index) override;
  Result OnMemoryFillExpr() override;
  Result OnMemoryGrowExpr() override;
  Result OnMemoryInitExpr(Index segment_index) override;
  Result OnMemorySizeExpr() override;
  Result OnTableCopyExpr() override;
  Result OnElemDropExpr(Index segment_index) override;
  Result OnTableInitExpr(Index segment_index) override;
  Result OnTableGetExpr(Index table_index) override;
  Result OnTableSetExpr(Index table_index) override;
  Result OnTableGrowExpr(Index table_index) override;
  Result OnTableSizeExpr(Index table_index) override;
  Result OnRefNullExpr() override;
  Result OnRefIsNullExpr() override;
  Result OnNopExpr() override;
  Result OnRethrowExpr() override;
  Result OnReturnExpr() override;
  Result OnSelectExpr() override;
  Result OnStoreExpr(Opcode opcode,
                     uint32_t alignment_log2,
                     Address offset) override;
  Result OnThrowExpr(Index event_index) override;
  Result OnTryExpr(Type sig_type) override;
  Result OnUnaryExpr(Opcode opcode) override;
  Result OnTernaryExpr(Opcode opcode) override;
  Result OnUnreachableExpr() override;
  Result EndFunctionBody(Index index) override;
  Result OnSimdLaneOpExpr(Opcode opcode, uint64_t value) override;
  Result OnSimdShuffleOpExpr(Opcode opcode, v128 value) override;

  Result OnElemSegmentCount(Index count) override;
  Result BeginElemSegment(Index index, Index table_index, bool passive) override;
  Result BeginElemSegmentInitExpr(Index index) override;
  Result EndElemSegmentInitExpr(Index index) override;
  Result OnElemSegmentFunctionIndexCount(Index index, Index count) override;
  Result OnElemSegmentFunctionIndex(Index index, Index func_index) override;

  Result OnDataSegmentCount(Index count) override;
  Result BeginDataSegment(Index index, Index memory_index, bool passive) override;
  Result BeginDataSegmentInitExpr(Index index) override;
  Result EndDataSegmentInitExpr(Index index) override;
  Result OnDataSegmentData(Index index,
                           const void* data,
                           Address size) override;

  Result OnModuleName(string_view module_name) override;
  Result OnFunctionNamesCount(Index num_functions) override;
  Result OnFunctionName(Index function_index,
                        string_view function_name) override;
  Result OnLocalNameLocalCount(Index function_index, Index num_locals) override;
  Result OnLocalName(Index function_index,
                     Index local_index,
                     string_view local_name) override;

  Result BeginEventSection(Offset size) override { return Result::Ok; }
  Result OnEventCount(Index count) override { return Result::Ok; }
  Result OnEventType(Index index, Index sig_index) override;
  Result EndEventSection() override { return Result::Ok; }

  Result OnInitExprF32ConstExpr(Index index, uint32_t value) override;
  Result OnInitExprF64ConstExpr(Index index, uint64_t value) override;
  Result OnInitExprV128ConstExpr(Index index, v128 value) override;
  Result OnInitExprGlobalGetExpr(Index index, Index global_index) override;
  Result OnInitExprI32ConstExpr(Index index, uint32_t value) override;
  Result OnInitExprI64ConstExpr(Index index, uint64_t value) override;

 private:
  Location GetLocation() const;
  void PrintError(const char* format, ...);
  void PushLabel(LabelType label_type,
                 ExprList* first,
                 Expr* context = nullptr);
  Result PopLabel();
  Result GetLabelAt(LabelNode** label, Index depth);
  Result TopLabel(LabelNode** label);
  Result TopLabelExpr(LabelNode** label, Expr** expr);
  Result AppendExpr(std::unique_ptr<Expr> expr);
  void SetBlockDeclaration(BlockDeclaration* decl, Type sig_type);

  std::string GetUniqueName(BindingHash* bindings,
                            const std::string& original_name);

  Errors* errors_ = nullptr;
  Module* module_ = nullptr;

  Func* current_func_ = nullptr;
  std::vector<LabelNode> label_stack_;
  ExprList* current_init_expr_ = nullptr;
  const char* filename_;
};

BinaryReaderIR::BinaryReaderIR(Module* out_module,
                               const char* filename,
                               Errors* errors)
    : errors_(errors), module_(out_module), filename_(filename) {}

Location BinaryReaderIR::GetLocation() const {
  Location loc;
  loc.filename = filename_;
  loc.offset = state->offset;
  return loc;
}

void WABT_PRINTF_FORMAT(2, 3) BinaryReaderIR::PrintError(const char* format,
                                                         ...) {
  WABT_SNPRINTF_ALLOCA(buffer, length, format);
  errors_->emplace_back(ErrorLevel::Error, Location(kInvalidOffset), buffer);
}

void BinaryReaderIR::PushLabel(LabelType label_type,
                               ExprList* first,
                               Expr* context) {
  label_stack_.emplace_back(label_type, first, context);
}

Result BinaryReaderIR::PopLabel() {
  if (label_stack_.size() == 0) {
    PrintError("popping empty label stack");
    return Result::Error;
  }

  label_stack_.pop_back();
  return Result::Ok;
}

Result BinaryReaderIR::GetLabelAt(LabelNode** label, Index depth) {
  if (depth >= label_stack_.size()) {
    PrintError("accessing stack depth: %" PRIindex " >= max: %" PRIzd, depth,
               label_stack_.size());
    return Result::Error;
  }

  *label = &label_stack_[label_stack_.size() - depth - 1];
  return Result::Ok;
}

Result BinaryReaderIR::TopLabel(LabelNode** label) {
  return GetLabelAt(label, 0);
}

Result BinaryReaderIR::TopLabelExpr(LabelNode** label, Expr** expr) {
  CHECK_RESULT(TopLabel(label));
  LabelNode* parent_label;
  CHECK_RESULT(GetLabelAt(&parent_label, 1));
  *expr = &parent_label->exprs->back();
  return Result::Ok;
}

Result BinaryReaderIR::AppendExpr(std::unique_ptr<Expr> expr) {
  expr->loc = GetLocation();
  LabelNode* label;
  CHECK_RESULT(TopLabel(&label));
  label->exprs->push_back(std::move(expr));
  return Result::Ok;
}

void BinaryReaderIR::SetBlockDeclaration(BlockDeclaration* decl,
                                         Type sig_type) {
  if (IsTypeIndex(sig_type)) {
    Index type_index = GetTypeIndex(sig_type);
    decl->has_func_type = true;
    decl->type_var = Var(type_index);
    decl->sig = module_->func_types[type_index]->sig;
  } else {
    decl->has_func_type = false;
    decl->sig.param_types.clear();
    decl->sig.result_types = GetInlineTypeVector(sig_type);
  }
}

std::string BinaryReaderIR::GetUniqueName(BindingHash* bindings,
                                          const std::string& orig_name) {
  int counter = 1;
  std::string unique_name = orig_name;
  while (bindings->count(unique_name) != 0) {
    unique_name = orig_name + "." + std::to_string(counter++);
  }
  return unique_name;
}

bool BinaryReaderIR::OnError(const Error& error) {
  errors_->push_back(error);
  return true;
}

Result BinaryReaderIR::OnTypeCount(Index count) {
  WABT_TRY
  module_->func_types.reserve(count);
  WABT_CATCH_BAD_ALLOC
  return Result::Ok;
}

Result BinaryReaderIR::OnType(Index index,
                              Index param_count,
                              Type* param_types,
                              Index result_count,
                              Type* result_types) {
  auto field = MakeUnique<FuncTypeModuleField>(GetLocation());
  FuncType& func_type = field->func_type;
  func_type.sig.param_types.assign(param_types, param_types + param_count);
  func_type.sig.result_types.assign(result_types, result_types + result_count);
  module_->AppendField(std::move(field));
  return Result::Ok;
}

Result BinaryReaderIR::OnImportCount(Index count) {
  WABT_TRY
  module_->imports.reserve(count);
  WABT_CATCH_BAD_ALLOC
  return Result::Ok;
}

Result BinaryReaderIR::OnImportFunc(Index import_index,
                                    string_view module_name,
                                    string_view field_name,
                                    Index func_index,
                                    Index sig_index) {
  auto import = MakeUnique<FuncImport>();
  import->module_name = module_name.to_string();
  import->field_name = field_name.to_string();
  import->func.decl.has_func_type = true;
  import->func.decl.type_var = Var(sig_index, GetLocation());
  import->func.decl.sig = module_->func_types[sig_index]->sig;
  module_->AppendField(
      MakeUnique<ImportModuleField>(std::move(import), GetLocation()));
  return Result::Ok;
}

Result BinaryReaderIR::OnImportTable(Index import_index,
                                     string_view module_name,
                                     string_view field_name,
                                     Index table_index,
                                     Type elem_type,
                                     const Limits* elem_limits) {
  auto import = MakeUnique<TableImport>();
  import->module_name = module_name.to_string();
  import->field_name = field_name.to_string();
  import->table.elem_limits = *elem_limits;
  module_->AppendField(
      MakeUnique<ImportModuleField>(std::move(import), GetLocation()));
  return Result::Ok;
}

Result BinaryReaderIR::OnImportMemory(Index import_index,
                                      string_view module_name,
                                      string_view field_name,
                                      Index memory_index,
                                      const Limits* page_limits) {
  auto import = MakeUnique<MemoryImport>();
  import->module_name = module_name.to_string();
  import->field_name = field_name.to_string();
  import->memory.page_limits = *page_limits;
  module_->AppendField(
      MakeUnique<ImportModuleField>(std::move(import), GetLocation()));
  return Result::Ok;
}

Result BinaryReaderIR::OnImportGlobal(Index import_index,
                                      string_view module_name,
                                      string_view field_name,
                                      Index global_index,
                                      Type type,
                                      bool mutable_) {
  auto import = MakeUnique<GlobalImport>();
  import->module_name = module_name.to_string();
  import->field_name = field_name.to_string();
  import->global.type = type;
  import->global.mutable_ = mutable_;
  module_->AppendField(
      MakeUnique<ImportModuleField>(std::move(import), GetLocation()));
  return Result::Ok;
}

Result BinaryReaderIR::OnImportEvent(Index import_index,
                                     string_view module_name,
                                     string_view field_name,
                                     Index event_index,
                                     Index sig_index) {
  auto import = MakeUnique<EventImport>();
  import->module_name = module_name.to_string();
  import->field_name = field_name.to_string();
  import->event.decl.has_func_type = true;
  import->event.decl.type_var = Var(sig_index, GetLocation());
  import->event.decl.sig = module_->func_types[sig_index]->sig;
  module_->AppendField(
      MakeUnique<ImportModuleField>(std::move(import), GetLocation()));
  return Result::Ok;
}

Result BinaryReaderIR::OnFunctionCount(Index count) {
  WABT_TRY
  module_->funcs.reserve(module_->num_func_imports + count);
  WABT_CATCH_BAD_ALLOC
  return Result::Ok;
}

Result BinaryReaderIR::OnFunction(Index index, Index sig_index) {
  auto field = MakeUnique<FuncModuleField>(GetLocation());
  Func& func = field->func;
  func.decl.has_func_type = true;
  func.decl.type_var = Var(sig_index, GetLocation());
  func.decl.sig = module_->func_types[sig_index]->sig;
  module_->AppendField(std::move(field));
  return Result::Ok;
}

Result BinaryReaderIR::OnTableCount(Index count) {
  WABT_TRY
  module_->tables.reserve(module_->num_table_imports + count);
  WABT_CATCH_BAD_ALLOC
  return Result::Ok;
}

Result BinaryReaderIR::OnTable(Index index,
                               Type elem_type,
                               const Limits* elem_limits) {
  auto field = MakeUnique<TableModuleField>(GetLocation());
  Table& table = field->table;
  table.elem_limits = *elem_limits;
  table.elem_type = elem_type;
  module_->AppendField(std::move(field));
  return Result::Ok;
}

Result BinaryReaderIR::OnMemoryCount(Index count) {
  WABT_TRY
  module_->memories.reserve(module_->num_memory_imports + count);
  WABT_CATCH_BAD_ALLOC
  return Result::Ok;
}

Result BinaryReaderIR::OnMemory(Index index, const Limits* page_limits) {
  auto field = MakeUnique<MemoryModuleField>(GetLocation());
  Memory& memory = field->memory;
  memory.page_limits = *page_limits;
  module_->AppendField(std::move(field));
  return Result::Ok;
}

Result BinaryReaderIR::OnGlobalCount(Index count) {
  WABT_TRY
  module_->globals.reserve(module_->num_global_imports + count);
  WABT_CATCH_BAD_ALLOC
  return Result::Ok;
}

Result BinaryReaderIR::BeginGlobal(Index index, Type type, bool mutable_) {
  auto field = MakeUnique<GlobalModuleField>(GetLocation());
  Global& global = field->global;
  global.type = type;
  global.mutable_ = mutable_;
  module_->AppendField(std::move(field));
  return Result::Ok;
}

Result BinaryReaderIR::BeginGlobalInitExpr(Index index) {
  assert(index == module_->globals.size() - 1);
  Global* global = module_->globals[index];
  current_init_expr_ = &global->init_expr;
  return Result::Ok;
}

Result BinaryReaderIR::EndGlobalInitExpr(Index index) {
  current_init_expr_ = nullptr;
  return Result::Ok;
}

Result BinaryReaderIR::OnExportCount(Index count) {
  WABT_TRY
  module_->exports.reserve(count);
  WABT_CATCH_BAD_ALLOC
  return Result::Ok;
}

Result BinaryReaderIR::OnExport(Index index,
                                ExternalKind kind,
                                Index item_index,
                                string_view name) {
  auto field = MakeUnique<ExportModuleField>(GetLocation());
  Export& export_ = field->export_;
  export_.name = name.to_string();
  switch (kind) {
    case ExternalKind::Func:
      assert(item_index < module_->funcs.size());
      break;
    case ExternalKind::Table:
      assert(item_index < module_->tables.size());
      break;
    case ExternalKind::Memory:
      assert(item_index < module_->memories.size());
      break;
    case ExternalKind::Global:
      assert(item_index < module_->globals.size());
      break;
    case ExternalKind::Event:
      assert(item_index < module_->events.size());
      break;
  }
  export_.var = Var(item_index, GetLocation());
  export_.kind = kind;
  module_->AppendField(std::move(field));
  return Result::Ok;
}

Result BinaryReaderIR::OnStartFunction(Index func_index) {
  assert(func_index < module_->funcs.size());
  Var start(func_index, GetLocation());
  module_->AppendField(MakeUnique<StartModuleField>(start, GetLocation()));
  return Result::Ok;
}

Result BinaryReaderIR::OnFunctionBodyCount(Index count) {
  assert(module_->num_func_imports + count == module_->funcs.size());
  return Result::Ok;
}

Result BinaryReaderIR::BeginFunctionBody(Index index, Offset size) {
  current_func_ = module_->funcs[index];
  PushLabel(LabelType::Func, &current_func_->exprs);
  return Result::Ok;
}

Result BinaryReaderIR::OnLocalDecl(Index decl_index, Index count, Type type) {
  current_func_->local_types.AppendDecl(type, count);
  return Result::Ok;
}

Result BinaryReaderIR::OnAtomicLoadExpr(Opcode opcode,
                                        uint32_t alignment_log2,
                                        Address offset) {
  return AppendExpr(
      MakeUnique<AtomicLoadExpr>(opcode, 1 << alignment_log2, offset));
}

Result BinaryReaderIR::OnAtomicStoreExpr(Opcode opcode,
                                         uint32_t alignment_log2,
                                         Address offset) {
  return AppendExpr(
      MakeUnique<AtomicStoreExpr>(opcode, 1 << alignment_log2, offset));
}

Result BinaryReaderIR::OnAtomicRmwExpr(Opcode opcode,
                                       uint32_t alignment_log2,
                                       Address offset) {
  return AppendExpr(
      MakeUnique<AtomicRmwExpr>(opcode, 1 << alignment_log2, offset));
}

Result BinaryReaderIR::OnAtomicRmwCmpxchgExpr(Opcode opcode,
                                              uint32_t alignment_log2,
                                              Address offset) {
  return AppendExpr(
      MakeUnique<AtomicRmwCmpxchgExpr>(opcode, 1 << alignment_log2, offset));
}

wabt::Result BinaryReaderIR::OnAtomicWaitExpr(Opcode opcode,
                                              uint32_t alignment_log2,
                                              Address offset) {
  return AppendExpr(
      MakeUnique<AtomicWaitExpr>(opcode, 1 << alignment_log2, offset));
}

wabt::Result BinaryReaderIR::OnAtomicNotifyExpr(Opcode opcode,
                                                uint32_t alignment_log2,
                                                Address offset) {
  return AppendExpr(
      MakeUnique<AtomicNotifyExpr>(opcode, 1 << alignment_log2, offset));
}

Result BinaryReaderIR::OnBinaryExpr(Opcode opcode) {
  return AppendExpr(MakeUnique<BinaryExpr>(opcode));
}

Result BinaryReaderIR::OnBlockExpr(Type sig_type) {
  auto expr = MakeUnique<BlockExpr>();
  SetBlockDeclaration(&expr->block.decl, sig_type);
  ExprList* expr_list = &expr->block.exprs;
  CHECK_RESULT(AppendExpr(std::move(expr)));
  PushLabel(LabelType::Block, expr_list);
  return Result::Ok;
}

Result BinaryReaderIR::OnBrExpr(Index depth) {
  return AppendExpr(MakeUnique<BrExpr>(Var(depth)));
}

Result BinaryReaderIR::OnBrIfExpr(Index depth) {
  return AppendExpr(MakeUnique<BrIfExpr>(Var(depth)));
}

Result BinaryReaderIR::OnBrOnExnExpr(Index depth, Index event_index) {
  auto expr = MakeUnique<BrOnExnExpr>();
  expr->label_var = Var(depth);
  expr->event_var = Var(event_index);
  return AppendExpr(std::move(expr));
}

Result BinaryReaderIR::OnBrTableExpr(Index num_targets,
                                     Index* target_depths,
                                     Index default_target_depth) {
  auto expr = MakeUnique<BrTableExpr>();
  expr->default_target = Var(default_target_depth);
  expr->targets.resize(num_targets);
  for (Index i = 0; i < num_targets; ++i) {
    expr->targets[i] = Var(target_depths[i]);
  }
  return AppendExpr(std::move(expr));
}

Result BinaryReaderIR::OnCallExpr(Index func_index) {
  assert(func_index < module_->funcs.size());
  return AppendExpr(MakeUnique<CallExpr>(Var(func_index)));
}

Result BinaryReaderIR::OnCallIndirectExpr(Index sig_index, Index table_index) {
  assert(sig_index < module_->func_types.size());
  auto expr = MakeUnique<CallIndirectExpr>();
  expr->decl.has_func_type = true;
  expr->decl.type_var = Var(sig_index, GetLocation());
  expr->decl.sig = module_->func_types[sig_index]->sig;
  expr->table = Var(table_index);
  return AppendExpr(std::move(expr));
}

Result BinaryReaderIR::OnReturnCallExpr(Index func_index) {
  assert(func_index < module_->funcs.size());
  return AppendExpr(MakeUnique<ReturnCallExpr>(Var(func_index)));
}

Result BinaryReaderIR::OnReturnCallIndirectExpr(Index sig_index, Index table_index) {
  assert(sig_index < module_->func_types.size());
  auto expr = MakeUnique<ReturnCallIndirectExpr>();
  expr->decl.has_func_type = true;
  expr->decl.type_var = Var(sig_index, GetLocation());
  expr->decl.sig = module_->func_types[sig_index]->sig;
  expr->table = Var(table_index);
  return AppendExpr(std::move(expr));
}

Result BinaryReaderIR::OnCompareExpr(Opcode opcode) {
  return AppendExpr(MakeUnique<CompareExpr>(opcode));
}

Result BinaryReaderIR::OnConvertExpr(Opcode opcode) {
  return AppendExpr(MakeUnique<ConvertExpr>(opcode));
}

Result BinaryReaderIR::OnDropExpr() {
  return AppendExpr(MakeUnique<DropExpr>());
}

Result BinaryReaderIR::OnElseExpr() {
  LabelNode* label;
  Expr* expr;
  CHECK_RESULT(TopLabelExpr(&label, &expr));

  if (label->label_type == LabelType::If) {
    auto* if_expr = cast<IfExpr>(expr);
    if_expr->true_.end_loc = GetLocation();
    label->exprs = &if_expr->false_;
    label->label_type = LabelType::Else;
  } else {
    PrintError("else expression without matching if");
    return Result::Error;
  }

  return Result::Ok;
}

Result BinaryReaderIR::OnEndExpr() {
  LabelNode* label;
  Expr* expr;
  CHECK_RESULT(TopLabelExpr(&label, &expr));
  switch (label->label_type) {
    case LabelType::Block:
      cast<BlockExpr>(expr)->block.end_loc = GetLocation();
      break;
    case LabelType::Loop:
      cast<LoopExpr>(expr)->block.end_loc = GetLocation();
      break;
    case LabelType::If:
      cast<IfExpr>(expr)->true_.end_loc = GetLocation();
      break;
    case LabelType::Else:
      cast<IfExpr>(expr)->false_end_loc = GetLocation();
      break;
    case LabelType::Try:
      cast<TryExpr>(expr)->block.end_loc = GetLocation();
      break;

    case LabelType::Func:
    case LabelType::Catch:
      break;
  }

  return PopLabel();
}

Result BinaryReaderIR::OnF32ConstExpr(uint32_t value_bits) {
  return AppendExpr(
      MakeUnique<ConstExpr>(Const::F32(value_bits, GetLocation())));
}

Result BinaryReaderIR::OnF64ConstExpr(uint64_t value_bits) {
  return AppendExpr(
      MakeUnique<ConstExpr>(Const::F64(value_bits, GetLocation())));
}

Result BinaryReaderIR::OnV128ConstExpr(v128 value_bits) {
  return AppendExpr(
      MakeUnique<ConstExpr>(Const::V128(value_bits, GetLocation())));
}

Result BinaryReaderIR::OnGlobalGetExpr(Index global_index) {
  return AppendExpr(
      MakeUnique<GlobalGetExpr>(Var(global_index, GetLocation())));
}

Result BinaryReaderIR::OnLocalGetExpr(Index local_index) {
  return AppendExpr(MakeUnique<LocalGetExpr>(Var(local_index, GetLocation())));
}

Result BinaryReaderIR::OnI32ConstExpr(uint32_t value) {
  return AppendExpr(MakeUnique<ConstExpr>(Const::I32(value, GetLocation())));
}

Result BinaryReaderIR::OnI64ConstExpr(uint64_t value) {
  return AppendExpr(MakeUnique<ConstExpr>(Const::I64(value, GetLocation())));
}

Result BinaryReaderIR::OnIfExpr(Type sig_type) {
  auto expr = MakeUnique<IfExpr>();
  SetBlockDeclaration(&expr->true_.decl, sig_type);
  ExprList* expr_list = &expr->true_.exprs;
  CHECK_RESULT(AppendExpr(std::move(expr)));
  PushLabel(LabelType::If, expr_list);
  return Result::Ok;
}

Result BinaryReaderIR::OnLoadExpr(Opcode opcode,
                                  uint32_t alignment_log2,
                                  Address offset) {
  return AppendExpr(MakeUnique<LoadExpr>(opcode, 1 << alignment_log2, offset));
}

Result BinaryReaderIR::OnLoopExpr(Type sig_type) {
  auto expr = MakeUnique<LoopExpr>();
  SetBlockDeclaration(&expr->block.decl, sig_type);
  ExprList* expr_list = &expr->block.exprs;
  CHECK_RESULT(AppendExpr(std::move(expr)));
  PushLabel(LabelType::Loop, expr_list);
  return Result::Ok;
}

Result BinaryReaderIR::OnMemoryCopyExpr() {
  return AppendExpr(MakeUnique<MemoryCopyExpr>());
}

Result BinaryReaderIR::OnDataDropExpr(Index segment) {
  return AppendExpr(MakeUnique<DataDropExpr>(Var(segment)));
}

Result BinaryReaderIR::OnMemoryFillExpr() {
  return AppendExpr(MakeUnique<MemoryFillExpr>());
}

Result BinaryReaderIR::OnMemoryGrowExpr() {
  return AppendExpr(MakeUnique<MemoryGrowExpr>());
}

Result BinaryReaderIR::OnMemoryInitExpr(Index segment) {
  return AppendExpr(MakeUnique<MemoryInitExpr>(Var(segment)));
}

Result BinaryReaderIR::OnMemorySizeExpr() {
  return AppendExpr(MakeUnique<MemorySizeExpr>());
}

Result BinaryReaderIR::OnTableCopyExpr() {
  return AppendExpr(MakeUnique<TableCopyExpr>());
}

Result BinaryReaderIR::OnElemDropExpr(Index segment) {
  return AppendExpr(MakeUnique<ElemDropExpr>(Var(segment)));
}

Result BinaryReaderIR::OnTableInitExpr(Index segment) {
  return AppendExpr(MakeUnique<TableInitExpr>(Var(segment)));
}

Result BinaryReaderIR::OnTableGetExpr(Index table_index) {
  return AppendExpr(MakeUnique<TableGetExpr>(Var(table_index)));
}

Result BinaryReaderIR::OnTableSetExpr(Index table_index) {
  return AppendExpr(MakeUnique<TableSetExpr>(Var(table_index)));
}

Result BinaryReaderIR::OnTableGrowExpr(Index table_index) {
  return AppendExpr(MakeUnique<TableGrowExpr>(Var(table_index)));
}

Result BinaryReaderIR::OnTableSizeExpr(Index table_index) {
  return AppendExpr(MakeUnique<TableSizeExpr>(Var(table_index)));
}

Result BinaryReaderIR::OnRefNullExpr() {
  return AppendExpr(MakeUnique<RefNullExpr>());
}

Result BinaryReaderIR::OnRefIsNullExpr() {
  return AppendExpr(MakeUnique<RefIsNullExpr>());
}

Result BinaryReaderIR::OnNopExpr() {
  return AppendExpr(MakeUnique<NopExpr>());
}

Result BinaryReaderIR::OnRethrowExpr() {
  return AppendExpr(MakeUnique<RethrowExpr>());
}

Result BinaryReaderIR::OnReturnExpr() {
  return AppendExpr(MakeUnique<ReturnExpr>());
}

Result BinaryReaderIR::OnSelectExpr() {
  return AppendExpr(MakeUnique<SelectExpr>());
}

Result BinaryReaderIR::OnGlobalSetExpr(Index global_index) {
  return AppendExpr(
      MakeUnique<GlobalSetExpr>(Var(global_index, GetLocation())));
}

Result BinaryReaderIR::OnLocalSetExpr(Index local_index) {
  return AppendExpr(MakeUnique<LocalSetExpr>(Var(local_index, GetLocation())));
}

Result BinaryReaderIR::OnStoreExpr(Opcode opcode,
                                   uint32_t alignment_log2,
                                   Address offset) {
  return AppendExpr(MakeUnique<StoreExpr>(opcode, 1 << alignment_log2, offset));
}

Result BinaryReaderIR::OnThrowExpr(Index event_index) {
  return AppendExpr(MakeUnique<ThrowExpr>(Var(event_index, GetLocation())));
}

Result BinaryReaderIR::OnLocalTeeExpr(Index local_index) {
  return AppendExpr(MakeUnique<LocalTeeExpr>(Var(local_index, GetLocation())));
}

Result BinaryReaderIR::OnTryExpr(Type sig_type) {
  auto expr_ptr = MakeUnique<TryExpr>();
  // Save expr so it can be used below, after expr_ptr has been moved.
  TryExpr* expr = expr_ptr.get();
  ExprList* expr_list = &expr->block.exprs;
  SetBlockDeclaration(&expr->block.decl, sig_type);
  CHECK_RESULT(AppendExpr(std::move(expr_ptr)));
  PushLabel(LabelType::Try, expr_list, expr);
  return Result::Ok;
}

Result BinaryReaderIR::OnCatchExpr() {
  LabelNode* label;
  CHECK_RESULT(TopLabel(&label));
  if (label->label_type != LabelType::Try) {
    PrintError("catch expression without matching try");
    return Result::Error;
  }

  LabelNode* parent_label;
  CHECK_RESULT(GetLabelAt(&parent_label, 1));

  label->label_type = LabelType::Catch;
  label->exprs = &cast<TryExpr>(&parent_label->exprs->back())->catch_;
  return Result::Ok;
}

Result BinaryReaderIR::OnUnaryExpr(Opcode opcode) {
  return AppendExpr(MakeUnique<UnaryExpr>(opcode));
}

Result BinaryReaderIR::OnTernaryExpr(Opcode opcode) {
  return AppendExpr(MakeUnique<TernaryExpr>(opcode));
}

Result BinaryReaderIR::OnUnreachableExpr() {
  return AppendExpr(MakeUnique<UnreachableExpr>());
}

Result BinaryReaderIR::EndFunctionBody(Index index) {
  CHECK_RESULT(PopLabel());
  current_func_ = nullptr;
  return Result::Ok;
}

Result BinaryReaderIR::OnSimdLaneOpExpr(Opcode opcode, uint64_t value) {
  return AppendExpr(MakeUnique<SimdLaneOpExpr>(opcode, value));
}

Result BinaryReaderIR::OnSimdShuffleOpExpr(Opcode opcode, v128 value) {
  return AppendExpr(MakeUnique<SimdShuffleOpExpr>(opcode, value));
}

Result BinaryReaderIR::OnElemSegmentCount(Index count) {
  WABT_TRY
  module_->elem_segments.reserve(count);
  WABT_CATCH_BAD_ALLOC
  return Result::Ok;
}

Result BinaryReaderIR::BeginElemSegment(Index index,
                                        Index table_index,
                                        bool passive) {
  auto field = MakeUnique<ElemSegmentModuleField>(GetLocation());
  ElemSegment& elem_segment = field->elem_segment;
  elem_segment.table_var = Var(table_index, GetLocation());
  elem_segment.passive = passive;
  module_->AppendField(std::move(field));
  return Result::Ok;
}

Result BinaryReaderIR::BeginElemSegmentInitExpr(Index index) {
  assert(index == module_->elem_segments.size() - 1);
  ElemSegment* segment = module_->elem_segments[index];
  current_init_expr_ = &segment->offset;
  return Result::Ok;
}

Result BinaryReaderIR::EndElemSegmentInitExpr(Index index) {
  current_init_expr_ = nullptr;
  return Result::Ok;
}

Result BinaryReaderIR::OnElemSegmentFunctionIndexCount(Index index,
                                                       Index count) {
  assert(index == module_->elem_segments.size() - 1);
  ElemSegment* segment = module_->elem_segments[index];
  WABT_TRY
  segment->vars.reserve(count);
  WABT_CATCH_BAD_ALLOC
  return Result::Ok;
}

Result BinaryReaderIR::OnElemSegmentFunctionIndex(Index segment_index,
                                                  Index func_index) {
  assert(segment_index == module_->elem_segments.size() - 1);
  ElemSegment* segment = module_->elem_segments[segment_index];
  segment->vars.emplace_back();
  Var* var = &segment->vars.back();
  *var = Var(func_index, GetLocation());
  return Result::Ok;
}

Result BinaryReaderIR::OnDataSegmentCount(Index count) {
  WABT_TRY
  module_->data_segments.reserve(count);
  WABT_CATCH_BAD_ALLOC
  return Result::Ok;
}

Result BinaryReaderIR::BeginDataSegment(Index index,
                                        Index memory_index,
                                        bool passive) {
  auto field = MakeUnique<DataSegmentModuleField>(GetLocation());
  DataSegment& data_segment = field->data_segment;
  data_segment.memory_var = Var(memory_index, GetLocation());
  data_segment.passive = passive;
  module_->AppendField(std::move(field));
  return Result::Ok;
}

Result BinaryReaderIR::BeginDataSegmentInitExpr(Index index) {
  assert(index == module_->data_segments.size() - 1);
  DataSegment* segment = module_->data_segments[index];
  current_init_expr_ = &segment->offset;
  return Result::Ok;
}

Result BinaryReaderIR::EndDataSegmentInitExpr(Index index) {
  current_init_expr_ = nullptr;
  return Result::Ok;
}

Result BinaryReaderIR::OnDataSegmentData(Index index,
                                         const void* data,
                                         Address size) {
  assert(index == module_->data_segments.size() - 1);
  DataSegment* segment = module_->data_segments[index];
  segment->data.resize(size);
  if (size > 0) {
    memcpy(segment->data.data(), data, size);
  }
  return Result::Ok;
}

Result BinaryReaderIR::OnFunctionNamesCount(Index count) {
  if (count > module_->funcs.size()) {
    PrintError("expected function name count (%" PRIindex
               ") <= function count (%" PRIzd ")",
               count, module_->funcs.size());
    return Result::Error;
  }
  return Result::Ok;
}

static std::string MakeDollarName(string_view name) {
  return std::string("$") + name.to_string();
}

Result BinaryReaderIR::OnModuleName(string_view name) {
  if (name.empty()) {
    return Result::Ok;
  }

  module_->name = MakeDollarName(name);
  return Result::Ok;
}

Result BinaryReaderIR::OnFunctionName(Index index, string_view name) {
  if (name.empty()) {
    return Result::Ok;
  }

  Func* func = module_->funcs[index];
  std::string dollar_name =
      GetUniqueName(&module_->func_bindings, MakeDollarName(name));
  func->name = dollar_name;
  module_->func_bindings.emplace(dollar_name, Binding(index));
  return Result::Ok;
}

Result BinaryReaderIR::OnLocalNameLocalCount(Index index, Index count) {
  assert(index < module_->funcs.size());
  Func* func = module_->funcs[index];
  Index num_params_and_locals = func->GetNumParamsAndLocals();
  if (count > num_params_and_locals) {
    PrintError("expected local name count (%" PRIindex
               ") <= local count (%" PRIindex ")",
               count, num_params_and_locals);
    return Result::Error;
  }
  return Result::Ok;
}

Result BinaryReaderIR::OnInitExprF32ConstExpr(Index index, uint32_t value) {
  Location loc = GetLocation();
  current_init_expr_->push_back(
      MakeUnique<ConstExpr>(Const::F32(value, loc), loc));
  return Result::Ok;
}

Result BinaryReaderIR::OnInitExprF64ConstExpr(Index index, uint64_t value) {
  Location loc = GetLocation();
  current_init_expr_->push_back(
      MakeUnique<ConstExpr>(Const::F64(value, loc), loc));
  return Result::Ok;
}

Result BinaryReaderIR::OnInitExprV128ConstExpr(Index index, v128 value) {
  Location loc = GetLocation();
  current_init_expr_->push_back(
      MakeUnique<ConstExpr>(Const::V128(value, loc), loc));
  return Result::Ok;
}

Result BinaryReaderIR::OnInitExprGlobalGetExpr(Index index,
                                               Index global_index) {
  Location loc = GetLocation();
  current_init_expr_->push_back(
      MakeUnique<GlobalGetExpr>(Var(global_index, loc), loc));
  return Result::Ok;
}

Result BinaryReaderIR::OnInitExprI32ConstExpr(Index index, uint32_t value) {
  Location loc = GetLocation();
  current_init_expr_->push_back(
      MakeUnique<ConstExpr>(Const::I32(value, loc), loc));
  return Result::Ok;
}

Result BinaryReaderIR::OnInitExprI64ConstExpr(Index index, uint64_t value) {
  Location loc = GetLocation();
  current_init_expr_->push_back(
      MakeUnique<ConstExpr>(Const::I64(value, loc), loc));
  return Result::Ok;
}

Result BinaryReaderIR::OnLocalName(Index func_index,
                                   Index local_index,
                                   string_view name) {
  if (name.empty()) {
    return Result::Ok;
  }

  Func* func = module_->funcs[func_index];
  func->bindings.emplace(GetUniqueName(&func->bindings, MakeDollarName(name)),
                         Binding(local_index));
  return Result::Ok;
}

Result BinaryReaderIR::OnEventType(Index index, Index sig_index) {
  auto field = MakeUnique<EventModuleField>(GetLocation());
  Event& event = field->event;
  event.decl.has_func_type = true;
  event.decl.type_var = Var(sig_index, GetLocation());
  event.decl.sig = module_->func_types[sig_index]->sig;
  module_->AppendField(std::move(field));
  return Result::Ok;
}

}  // end anonymous namespace

Result ReadBinaryIr(const char* filename,
                    const void* data,
                    size_t size,
                    const ReadBinaryOptions& options,
                    Errors* errors,
                    Module* out_module) {
  BinaryReaderIR reader(out_module, filename, errors);
  return ReadBinary(data, size, &reader, options);
}

}  // namespace wabt
