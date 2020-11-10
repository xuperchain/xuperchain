package evm

import (
	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/execution/evm/asm"
	"github.com/tmthrgd/go-bitset"
)

type Code struct {
	Bytecode     acm.Bytecode
	OpcodeBitset bitset.Bitset
}

// Build a Code object that includes analysis of which symbols are opcodes versus push data
func NewCode(code []byte) *Code {
	return &Code{
		Bytecode:     code,
		OpcodeBitset: opcodeBitset(code),
	}
}

func (code *Code) Length() uint64 {
	if code == nil {
		return 0
	}
	return uint64(len(code.Bytecode))
}

func (code *Code) GetBytecode() acm.Bytecode {
	if code == nil {
		return nil
	}
	return code.Bytecode
}

func (code *Code) IsOpcode(indexOfSymbolInCode uint64) bool {
	if code == nil || indexOfSymbolInCode >= uint64(code.OpcodeBitset.Len()) {
		return false
	}
	return code.OpcodeBitset.IsSet(uint(indexOfSymbolInCode))
}

func (code *Code) IsPushData(indexOfSymbolInCode uint64) bool {
	return !code.IsOpcode(indexOfSymbolInCode)
}

func (code *Code) GetSymbol(n uint64) asm.OpCode {
	if code.Length() <= n {
		return asm.STOP
	} else {
		return asm.OpCode(code.Bytecode[n])
	}
}

// If code[i] is an opcode (rather than PUSH data) then bitset.IsSet(i) will be true
func opcodeBitset(code []byte) bitset.Bitset {
	bs := bitset.New(uint(len(code)))
	for i := 0; i < len(code); i++ {
		bs.Set(uint(i))
		symbol := asm.OpCode(code[i])
		if symbol >= asm.PUSH1 && symbol <= asm.PUSH32 {
			i += int(symbol - asm.PUSH1 + 1)
		}
	}
	return bs
}
