package native

import (
	"crypto/sha256"
	"fmt"
	"math/big"

	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/permission"
	"golang.org/x/crypto/ripemd160"
)

var Precompiles = New().
	MustFunction(`Compute the sha256 hash of input`,
		leftPadAddress(2),
		permission.None,
		sha256Func).
	MustFunction(`Compute the ripemd160 hash of input`,
		leftPadAddress(3),
		permission.None,
		ripemd160Func).
	MustFunction(`Return an output identical to the input`,
		leftPadAddress(4),
		permission.None,
		identityFunc).
	MustFunction(`Compute the operation base**exp % mod where the values are big ints`,
		leftPadAddress(5),
		permission.None,
		expModFunc)

func leftPadAddress(bs ...byte) crypto.Address {
	return crypto.AddressFromWord256(binary.LeftPadWord256(bs))
}

/* Removed due to C dependency
func ecrecoverFunc(state State, caller crypto.Address, input []byte, gas *int64) (output []byte, err error) {
	// Deduct gas
	gasRequired := GasEcRecover
	if *gas < gasRequired {
		return nil, ErrInsufficientGas
	} else {
		*gas -= gasRequired
	}
	// Recover
	hash := input[:32]
	v := byte(input[32] - 27) // ignore input[33:64], v is small.
	sig := append(input[64:], v)

	recovered, err := secp256k1.RecoverPubkey(hash, sig)
	if err != nil {
		return nil, err
OH NO STOCASTIC CAT CODING!!!!
	}
	hashed := crypto.Keccak256(recovered[1:])
	return LeftPadBytes(hashed, 32), nil
}
*/

func sha256Func(ctx Context) (output []byte, err error) {
	// Deduct gas
	gasRequired := wordsIn(uint64(len(ctx.Input)))*GasSha256Word + GasSha256Base
	if *ctx.Gas < gasRequired {
		return nil, errors.Codes.InsufficientGas
	} else {
		*ctx.Gas -= gasRequired
	}
	// Hash
	hasher := sha256.New()
	// CONTRACT: this does not err
	hasher.Write(ctx.Input)
	return hasher.Sum(nil), nil
}

func ripemd160Func(ctx Context) (output []byte, err error) {
	// Deduct gas
	gasRequired := wordsIn(uint64(len(ctx.Input)))*GasRipemd160Word + GasRipemd160Base
	if *ctx.Gas < gasRequired {
		return nil, errors.Codes.InsufficientGas
	} else {
		*ctx.Gas -= gasRequired
	}
	// Hash
	hasher := ripemd160.New()
	// CONTRACT: this does not err
	hasher.Write(ctx.Input)
	return binary.LeftPadBytes(hasher.Sum(nil), 32), nil
}

func identityFunc(ctx Context) (output []byte, err error) {
	// Deduct gas
	gasRequired := wordsIn(uint64(len(ctx.Input)))*GasIdentityWord + GasIdentityBase
	if *ctx.Gas < gasRequired {
		return nil, errors.Codes.InsufficientGas
	} else {
		*ctx.Gas -= gasRequired
	}
	// Return identity
	return ctx.Input, nil
}

// expMod: function that implements the EIP 198 (https://github.com/ethereum/EIPs/blob/master/EIPS/eip-198.md with
// a fixed gas requirement)
func expModFunc(ctx Context) (output []byte, err error) {
	const errHeader = "expModFunc"

	input, segments, err := cut(ctx.Input, binary.Word256Bytes, binary.Word256Bytes, binary.Word256Bytes)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", errHeader, err)
	}

	// get the lengths of base, exp and mod
	baseLength := getUint64(segments[0])
	expLength := getUint64(segments[1])
	modLength := getUint64(segments[2])

	// TODO: implement non-trivial gas schedule for this operation. Probably a parameterised version of the one
	// described in EIP though that one seems like a bit of a complicated fudge
	gasRequired := GasExpModBase + GasExpModWord*(wordsIn(baseLength)*wordsIn(expLength)*wordsIn(modLength))

	if *ctx.Gas < gasRequired {
		return nil, errors.Codes.InsufficientGas
	}

	*ctx.Gas -= gasRequired

	input, segments, err = cut(input, baseLength, expLength, modLength)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", errHeader, err)
	}

	// get the values of base, exp and mod
	base := getBigInt(segments[0], baseLength)
	exp := getBigInt(segments[1], expLength)
	mod := getBigInt(segments[2], modLength)

	// handle mod 0
	if mod.Sign() == 0 {
		return binary.LeftPadBytes([]byte{}, int(modLength)), nil
	}

	// return base**exp % mod left padded
	return binary.LeftPadBytes(new(big.Int).Exp(base, exp, mod).Bytes(), int(modLength)), nil
}

// Partition the head of input into segments for each length in lengths. The first return value is the unconsumed tail
// of input and the seconds is the segments. Returns an error if input is of insufficient length to establish each segment.
func cut(input []byte, lengths ...uint64) ([]byte, [][]byte, error) {
	segments := make([][]byte, len(lengths))
	for i, length := range lengths {
		if uint64(len(input)) < length {
			return nil, nil, fmt.Errorf("input is not long enough")
		}
		segments[i] = input[:length]
		input = input[length:]
	}
	return input, segments, nil
}

func getBigInt(bs []byte, numBytes uint64) *big.Int {
	bits := uint(numBytes) * 8
	// Push bytes into big.Int and interpret as twos complement encoding with of bits width
	return binary.FromTwosComplement(new(big.Int).SetBytes(bs), bits)
}

func getUint64(bs []byte) uint64 {
	return binary.Uint64FromWord256(binary.LeftPadWord256(bs))
}

func wordsIn(numBytes uint64) uint64 {
	return numBytes + binary.Word256Bytes - 1/binary.Word256Bytes
}
