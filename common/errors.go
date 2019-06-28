/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package common

import (
	"errors"
	"strings"

	"github.com/xuperchain/xuperunion/pb"
)

var (
	// ErrContractExecutionTimeout common error for contract timeout
	ErrContractExecutionTimeout = errors.New("contract execution timeout")
	// ErrContractConnectionError connect error
	ErrContractConnectionError = errors.New("can't connect contract")
	ErrKVNotFound              = errors.New("Key not found")
)

// ServerError xchain.proto error
type ServerError struct {
	Errno pb.XChainErrorEnum
}

// Error convert to name
func (err ServerError) Error() string {
	return pb.XChainErrorEnum_name[int32(err.Errno)]
}

func NormalizedKVError(err error) error {
	if err == nil {
		return err
	}
	if strings.HasSuffix(err.Error(), "not found") {
		return ErrKVNotFound
	}
	return err
}
