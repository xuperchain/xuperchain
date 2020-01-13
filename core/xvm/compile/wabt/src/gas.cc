#include "src/cast.h"
#include "src/common.h"
#include "src/ir.h"
#include "src/literal.h"
#include "src/stream.h"
#include "src/string-view.h"
#include "src/gas-cost-table.h"

namespace wabt {
namespace {

class GasWriter {
 public:
  void WriteExprList(ExprList* exprs);
  void WriteFunc(Func* func);
};

void GasWriter::WriteExprList(ExprList* exprs) {
  for (auto it = exprs->begin(); it != exprs->end();) {
    auto firstExpr = it;
    bool block_end = false;
    int64_t cost = 0;
    Opcode::Enum opcode_type;
    for (; it != exprs->end() && !block_end; it++) {
      switch (it->type()) {
        case ExprType::Binary: {
          BinaryExpr& expr = *cast<BinaryExpr>(&*it);
          opcode_type = expr.opcode;
          break;
        }

        case ExprType::Block: {
          Block& block = cast<BlockExpr>(&*it)->block;
          WriteExprList(&block.exprs);
          block_end = true;
          opcode_type = Opcode::Enum::Block;
          break;
        }

        case ExprType::Br: {
          block_end = true;
          opcode_type = Opcode::Enum::Br;
          break;
        }

        case ExprType::BrIf: {
          block_end = true;
          opcode_type = Opcode::Enum::BrIf;
          break;
        }

        case ExprType::BrTable:
          block_end = true;
          opcode_type = Opcode::Enum::BrTable;
          break;

        case ExprType::Call:
          opcode_type = Opcode::Enum::Call;
          break;

        case ExprType::CallIndirect:
          opcode_type = Opcode::Enum::CallIndirect;
          break;

        case ExprType::Compare: {
          CompareExpr& expr = *cast<CompareExpr>(&*it);
          opcode_type = expr.opcode;
          break;
        }

        // 不同位的操作全部归一化到64位
        case ExprType::Const: {
          opcode_type = Opcode::Enum::I64Const;
          break;
        }

        case ExprType::Convert: {
          ConvertExpr& expr = *cast<ConvertExpr>(&*it);
          opcode_type = expr.opcode;
          break;
        }

        case ExprType::Drop:
          opcode_type = Opcode::Enum::Drop;
          break;

        case ExprType::GlobalGet:
          opcode_type = Opcode::Enum::GlobalGet;
          break;

        case ExprType::GlobalSet:
          opcode_type = Opcode::Enum::GlobalSet;
          break;

        case ExprType::If: {
          IfExpr& if_ = *cast<IfExpr>(&*it);
          WriteExprList(&if_.true_.exprs);
          if (!if_.false_.empty()) {
            WriteExprList(&if_.false_);
          }
          opcode_type = Opcode::Enum::If;
          block_end = true;
          break;
        }

        case ExprType::Load: {
          LoadExpr& expr = *cast<LoadExpr>(&*it);
          opcode_type = expr.opcode;
          break;
        }

        case ExprType::LocalGet:
          opcode_type = Opcode::Enum::LocalGet;
          break;
        case ExprType::LocalSet:
          opcode_type = Opcode::Enum::LocalSet;
          break;
        case ExprType::LocalTee:
          opcode_type = Opcode::Enum::LocalTee;
          break;

        case ExprType::Loop: {
          Block& block = cast<LoopExpr>(&*it)->block;
          WriteExprList(&block.exprs);
          opcode_type = Opcode::Enum::Loop;
          block_end = true;
          break;
        }

        case ExprType::MemoryCopy:
        case ExprType::DataDrop:
        case ExprType::MemoryInit:
        case ExprType::MemoryFill:
        case ExprType::TableCopy:
        case ExprType::ElemDrop:
        case ExprType::TableInit:
        case ExprType::TableGet:
        case ExprType::TableSet:
        case ExprType::TableGrow:
        case ExprType::TableSize:
        case ExprType::RefNull:
        case ExprType::RefIsNull:
          fprintf(stderr, "unimplemented");
          abort();
          break;

        case ExprType::MemoryGrow:
          opcode_type = Opcode::Enum::MemoryGrow;
          break;

        case ExprType::MemorySize:
          opcode_type = Opcode::Enum::MemorySize;
          break;

        case ExprType::Nop:
          opcode_type = Opcode::Enum::Nop;
          break;

        case ExprType::Return:
          block_end = true;
          opcode_type = Opcode::Enum::Return;
          break;

        case ExprType::Select:
          opcode_type = Opcode::Enum::Select;
          break;

        case ExprType::Store: {
          StoreExpr& expr = *cast<StoreExpr>(&*it);
          opcode_type = expr.opcode;
          break;
        }

        case ExprType::Unary: {
          UnaryExpr& expr = *cast<UnaryExpr>(&*it);
          opcode_type = expr.opcode;
          break;
        }

        case ExprType::Ternary: {
          TernaryExpr& expr = *cast<TernaryExpr>(&*it);
          opcode_type = expr.opcode;
          break;
        }

        case ExprType::SimdLaneOp: {
          SimdLaneOpExpr& expr = *cast<SimdLaneOpExpr>(&*it);
          opcode_type = expr.opcode;
          break;
        }

        case ExprType::SimdShuffleOp: {
          SimdShuffleOpExpr& expr = *cast<SimdShuffleOpExpr>(&*it);
          opcode_type = expr.opcode;
          break;
        }

        case ExprType::Unreachable: {
          opcode_type = Opcode::Enum::Unreachable;
          break;
        }
        case ExprType::AtomicLoad:
        case ExprType::AtomicRmw:
        case ExprType::AtomicRmwCmpxchg:
        case ExprType::AtomicStore:
        case ExprType::AtomicWait:
        case ExprType::AtomicNotify:
        case ExprType::BrOnExn:
        case ExprType::Rethrow:
        case ExprType::ReturnCall:
        case ExprType::ReturnCallIndirect:
        case ExprType::Throw:
        case ExprType::Try:
          fprintf(stderr, "unimplemented");
          abort();
          break;

        default:
          fprintf(stderr, "unknow type, %s\n", GetExprTypeName(*it));
          opcode_type = Opcode::Enum::Invalid;
          break;
      }
      Opcode op = Opcode(opcode_type);
      cost += kGasCostTable[op.GetName()];
    }
    if (cost == 0) {
      continue;
    }
    auto gas_expr = MakeUnique<AddGasExpr>(cost);
    exprs->insert(firstExpr, std::move(gas_expr));
  }
}

void GasWriter::WriteFunc(Func* func) {
  ExprList& exprs = func->exprs;
  WriteExprList(&exprs);
}
}

void WriteModuleGas(Module* module) {
  GasWriter writer;
  for (Func* func : module->funcs) {
    writer.WriteFunc(func);
  }
}
}
