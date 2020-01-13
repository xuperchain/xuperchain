/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package common

import (
	"errors"
	"strings"

	"github.com/xuperchain/xuperchain/core/pb"
)

var (
	// ErrContractExecutionTimeout common error for contract timeout
	ErrContractExecutionTimeout = errors.New("contract execution timeout")
	// ErrContractConnectionError connect error
	ErrContractConnectionError = errors.New("can't connect contract")
	ErrKVNotFound              = errors.New("Key not found")
	ErrP2PError                = errors.New("invalid stream")
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
	if isInvalidStream(err.Error()) {
		return ErrP2PError
	}
	return err
}

func isInvalidStream(err string) bool {
	if strings.HasSuffix(err, "stream reset") {
		return true
	}
	if strings.HasSuffix(err, "connection reset by peer") {
		return true
	}
	if strings.HasSuffix(err, "stream closed") {
		return true
	}
	if strings.HasSuffix(err, "stream not valid") {
		return true
	}
	return false
}
