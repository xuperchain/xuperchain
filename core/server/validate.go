/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package server

import (
	"errors"
	"strconv"

	xchaincore "github.com/xuperchain/xuperchain/core/core"
	"github.com/xuperchain/xuperchain/core/crypto/hash"
	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/permission/acl"
)

func validateSendBlock(block *pb.Block) error {
	if len(block.Blockid) == 0 {
		return errors.New("validation error: validateSendBlock Block.Blockid can't be null")
	}

	if nil == block.Block {
		return errors.New("validation error: validateSendBlock Block.Block can't be null")
	}
	return nil
}

func validUtxoAccess(in *pb.UtxoInput, bc *xchaincore.XChainCore, requestAmount int64) bool {
	if bc == nil {
		return false
	}
	account := in.GetAddress()
	needLock := in.GetNeedLock()
	if acl.IsAccount(account) == 1 || !needLock {
		return true
	}
	publicKey, err := bc.CryptoClient.GetEcdsaPublicKeyFromJSON([]byte(in.Publickey))
	if err != nil {
		return false
	}
	checkSignResult, err := bc.CryptoClient.VerifyECDSA(publicKey, in.UserSign, hash.DoubleSha256([]byte(in.Bcname+in.Address+strconv.FormatInt(requestAmount, 10)+strconv.FormatBool(in.NeedLock))))
	if err != nil {
		return false
	}
	if checkSignResult != true {
		return false
	}
	addrMatchCheckResult, _ := bc.CryptoClient.VerifyAddressUsingPublicKey(in.Address, publicKey)
	if addrMatchCheckResult != true {
		return false
	}

	return true
}
