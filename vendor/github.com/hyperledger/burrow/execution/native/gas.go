// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package native

const (
	GasSha3          uint64 = 1
	GasGetAccount    uint64 = 1
	GasStorageUpdate uint64 = 1
	GasCreateAccount uint64 = 1

	GasBaseOp  uint64 = 0 // TODO: make this 1
	GasStackOp uint64 = 1

	GasEcRecover     uint64 = 1
	GasSha256Word    uint64 = 1
	GasSha256Base    uint64 = 1
	GasRipemd160Word uint64 = 1
	GasRipemd160Base uint64 = 1
	GasExpModWord    uint64 = 1
	GasExpModBase    uint64 = 1
	GasIdentityWord  uint64 = 1
	GasIdentityBase  uint64 = 1
)
