/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package common

import (
	"bytes"
	"io/ioutil"

	"github.com/golang/protobuf/proto"

	"github.com/xuperchain/xuperchain/core/pb"
)

// GetTxSerializedSize get size(in bytes) of a tx after being serialized
// https://godoc.org/github.com/golang/protobuf/proto#Size
func GetTxSerializedSize(pTx *pb.Transaction) (n int64, err error) {
	return int64(proto.Size(pTx)), nil
}

// GetBlkHeaderSerializedSize get size(in bytes) of a internal block's header info,
// which will be written to db
func GetBlkHeaderSerializedSize(pIntBlk *pb.InternalBlock) (n int64, err error) {
	txs := pIntBlk.Transactions

	pIntBlk.Transactions = []*pb.Transaction{}
	blkBuf, err := proto.Marshal(pIntBlk)
	if nil != err {
		return
	}

	n = int64(len(blkBuf))

	pIntBlk.Transactions = txs
	return
}

// GetIntBlkSerializedSize get size(in bytes) of a internal block after being serialized
// blockSize = headerSize + sum(txSize)
func GetIntBlkSerializedSize(pIntBlk *pb.InternalBlock) (n int64, err error) {

	n, err = GetBlkHeaderSerializedSize(pIntBlk)

	for _, tx := range pIntBlk.Transactions {
		s, err := GetTxSerializedSize(tx)
		if nil != err {
			return 0, err
		}
		n += s
	}
	return
}

// GetFileContent read file content and return
// moved from console/xcmd.go
func GetFileContent(file string) string {
	f, _ := ioutil.ReadFile(file)
	f = bytes.TrimRight(f, "\n")
	return string(f)
}
