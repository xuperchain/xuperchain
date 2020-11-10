package acm

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/burrow/execution/evm/asm"
	"github.com/hyperledger/burrow/execution/evm/asm/bc"
	hex "github.com/tmthrgd/go-hex"
)

type Bytecode []byte

// Builds new bytecode using the Splice helper to map byte-like and byte-slice-like types to a flat bytecode slice
func NewBytecode(bytelikes ...interface{}) (Bytecode, error) {
	return bc.Splice(bytelikes...)
}

func BytecodeFromHex(hexString string) (Bytecode, error) {
	var bc Bytecode
	err := bc.UnmarshalText([]byte(hexString))
	if err != nil {
		return nil, err
	}
	return bc, nil
}

func (bc Bytecode) Bytes() []byte {
	return bc[:]
}

func (bc Bytecode) String() string {
	return hex.EncodeUpperToString(bc[:])

}
func (bc Bytecode) MarshalJSON() ([]byte, error) {
	text, err := bc.MarshalText()
	if err != nil {
		return nil, err
	}
	return json.Marshal(string(text))
}

func (bc *Bytecode) UnmarshalJSON(data []byte) error {
	str := new(string)
	err := json.Unmarshal(data, str)
	if err != nil {
		return err
	}
	err = bc.UnmarshalText([]byte(*str))
	if err != nil {
		return err
	}
	return nil
}

func (bc Bytecode) MarshalText() ([]byte, error) {
	return ([]byte)(hex.EncodeUpperToString(bc)), nil
}

func (bc *Bytecode) UnmarshalText(text []byte) error {
	*bc = make([]byte, hex.DecodedLen(len(text)))
	_, err := hex.Decode(*bc, text)
	return err
}

// Protobuf support
func (bc Bytecode) Marshal() ([]byte, error) {
	return bc, nil
}

func (bc *Bytecode) Unmarshal(data []byte) error {
	*bc = data
	return nil
}

func (bc Bytecode) MarshalTo(data []byte) (int, error) {
	return copy(data, bc), nil
}

func (bc Bytecode) Size() int {
	return len(bc)
}

func (bc Bytecode) MustTokens() []string {
	tokens, err := bc.Tokens()
	if err != nil {
		panic(err)
	}
	return tokens
}

// Tokenises the bytecode into opcodes and values
func (bc Bytecode) Tokens() ([]string, error) {
	// Overestimate of capacity in the presence of pushes
	tokens := make([]string, 0, len(bc))

	for i := 0; i < len(bc); i++ {
		op, ok := asm.GetOpCode(bc[i])
		if !ok {
			return tokens, fmt.Errorf("did not recognise byte %#x at position %v as an OpCode:\n %s",
				bc[i], i, lexingPositionString(bc, i, tokens))
		}
		pushes := op.Pushes()
		tokens = append(tokens, op.Name())
		if pushes > 0 {
			// This is a PUSH<N> OpCode so consume N bytes from the input, render them as hex, and skip to next OpCode
			if i+pushes >= len(bc) {
				return tokens, fmt.Errorf("token %v of input is %s but not enough input remains to push %v: %s",
					i, op.Name(), pushes, lexingPositionString(bc, i, tokens))
			}
			pushedBytes := bc[i+1 : i+pushes+1]
			tokens = append(tokens, fmt.Sprintf("0x%s", pushedBytes))
			i += pushes
		}
	}
	return tokens, nil
}

func lexingPositionString(bc Bytecode, position int, produced []string) string {
	return fmt.Sprintf("%v_%v", produced, []byte(bc[position:]))
}
