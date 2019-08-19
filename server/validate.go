/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package server

import (
	"errors"
	"strconv"

	xchaincore "github.com/xuperchain/xuperunion/core"
	"github.com/xuperchain/xuperunion/crypto/hash"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/permission/acl"
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

func validUtxoAccess(in *pb.UtxoInput, bc *xchaincore.XChainCore) bool {
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
	checkSignResult, err := bc.CryptoClient.VerifyECDSA(publicKey, in.UserSign, hash.DoubleSha256([]byte(in.Bcname+in.Address+in.TotalNeed+strconv.FormatBool(in.NeedLock))))
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
