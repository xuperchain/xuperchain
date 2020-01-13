package disasm

import (
	"fmt"

	ops "github.com/xuperchain/wagon/wasm/operators"
)

func appendGasInstr(instrs []Instr) ([]Instr, *interface{}) {
	gasOp, _ := ops.New(ops.CheckGas)
	gasIns := Instr{
		Op:         gasOp,
		Immediates: make([]interface{}, 1),
	}
	gasNumPtr := &gasIns.Immediates[0]
	instrs = append(instrs, gasIns)
	return instrs, gasNumPtr
}

type GasMapper interface {
	MapGas(op string) (int64, bool)
}

// AddGasInstr add checkGas instructions to instrs
// FIXME: more efficient
func AddGasInstr(instrs []Instr, gasMapper GasMapper) []Instr {
	out := make([]Instr, 0, 2*len(instrs))
	var gasNumPtr *interface{}
	for _, ins := range instrs {
		op := ins.Op
		switch op.Code {
		case ops.Else:
			out = append(out, ins)
		default:
			// case ops.Block, ops.Br, ops.BrIf, ops.BrTable, ops.Return, ops.If, ops.Else, ops.Loop:
			out, gasNumPtr = appendGasInstr(out)
			used, ok := gasMapper.MapGas(op.Name)
			// FIXME:
			if !ok {
				panic(fmt.Sprintf("gas for %s not found", op.Name))
			}
			*gasNumPtr = used
			out = append(out, ins)
		}
	}
	return out
}
