/*
 * Copyright 2019 Baidu, Inc.
 * tx_verification implements the verification related functions of Transaction
 */

package utxo

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/xuperchain/xuperunion/contract"
	"github.com/xuperchain/xuperunion/crypto/client"
	"github.com/xuperchain/xuperunion/crypto/hash"
	"github.com/xuperchain/xuperunion/pb"
	pm "github.com/xuperchain/xuperunion/permission"
	"github.com/xuperchain/xuperunion/permission/acl"
	aclu "github.com/xuperchain/xuperunion/permission/acl/utils"
	"github.com/xuperchain/xuperunion/utxo/txhash"
	"github.com/xuperchain/xuperunion/xmodel"
)

// ImmediateVerifyTx verify tx Immediately
// Transaction verification workflow:
//   1. verify transaction ID is the same with data hash
//   2. verify initiator type, should be ak
//   3. verify all signatures of initiator and auth requires
//   4. verify the account ACL of utxo input
//   5. verify the contract requests' permission
//   6. verify the permission of contract RWSet (WriteSet could including unauthorized data change)
//   7. run contract requests and verify if the RWSet result is the same with preExed RWSet (heavy
//      operation, keep it at last)
func (uv *UtxoVM) ImmediateVerifyTx(tx *pb.Transaction, isRootTx bool) (bool, error) {
	// Pre processing of tx data
	if !isRootTx && tx.Version == RootTxVersion {
		return false, ErrVersionInvalid
	}
	if tx.Version > BetaTxVersion || tx.Version < RootTxVersion {
		return false, ErrVersionInvalid
	}
	// autogen tx should not run ImmediateVerifyTx, this could be a fake tx
	if tx.Autogen {
		return false, ErrInvalidAutogenTx
	}
	if proto.Size(tx) > uv.ledger.MaxTxSizePerBlock() {
		uv.xlog.Warn("tx too large, should not be greater than half of max blocksize", "size", proto.Size(tx))
		return false, ErrTxTooLarge
	}

	// Start transaction verification workflow
	if tx.Version > RootTxVersion {
		// verify txid
		txid, err := txhash.MakeTransactionID(tx)
		if err != nil {
			uv.xlog.Warn("ImmediateVerifyTx: call MakeTransactionID failed", "error", err)
			return false, err
		}
		if bytes.Compare(tx.Txid, txid) != 0 {
			uv.xlog.Warn("ImmediateVerifyTx: txid not match", "tx.Txid", tx.Txid, "txid", txid)
			return false, fmt.Errorf("Txid verify failed")
		}

		// get digestHash
		digestHash, err := txhash.MakeTxDigestHash(tx)
		if err != nil {
			uv.xlog.Warn("ImmediateVerifyTx: call MakeTxDigestHash failed", "error", err)
			return false, err
		}

		// verify signatures
		ok, verifiedID, err := uv.verifySignatures(tx, digestHash)
		if !ok {
			uv.xlog.Warn("ImmediateVerifyTx: verifySignatures failed", "error", err)
			return ok, ErrInvalidSignature
		}

		// veify tx UTXO input permission (Account ACL)
		ok, err = uv.verifyUTXOPermission(tx, verifiedID)
		if !ok {
			uv.xlog.Warn("ImmediateVerifyTx: verifyUTXOPermission failed", "error", err)
			return ok, ErrACLNotEnough
		}

		// verify contract requests' permission using ACL
		ok, err = uv.verifyContractPermission(tx)
		if !ok {
			uv.xlog.Warn("ImmediateVerifyTx: verifyContractPermission failed", "error", err)
			return ok, ErrACLNotEnough
		}

		// verify the permission of RWSet using ACL
		ok, err = uv.verifyRWSetPermission(tx, verifiedID)
		if !ok {
			uv.xlog.Warn("ImmediateVerifyTx: verifyRWSetPermission failed", "error", err)
			return ok, ErrACLNotEnough
		}

		// verify RWSet(run contracts and compare RWSet)
		ok, err = uv.verifyTxRWSets(tx)
		if err != nil {
			uv.xlog.Warn("ImmediateVerifyTx: verifyTxRWSets failed", "error", err)
			// reset error message
			if strings.HasPrefix(err.Error(), "Gas not enough") {
				err = ErrGasNotEnough
			} else {
				err = ErrRWSetInvalid
			}
			return ok, err
		}
		if !ok {
			// always return RWSet Invalid Error if verification not passed
			return ok, ErrRWSetInvalid
		}
	}
	return true, nil
}

// verify signatures only, from V3.3, we verify all signatures ahead of permission
// Note that if tx.XuperSign is not nil, the signature verification use XuperSign process
func (uv *UtxoVM) verifySignatures(tx *pb.Transaction, digestHash []byte) (bool, map[string]bool, error) {
	// XuperSign is not empty, use XuperSign verify
	if tx.GetXuperSign() != nil {
		return uv.verifyXuperSign(tx, digestHash)
	}

	// Not XuperSign(multisig/rignsign etc.), use old signature process
	verifiedAddr := make(map[string]bool)
	if len(tx.InitiatorSigns) < 1 || len(tx.AuthRequire) != len(tx.AuthRequireSigns) {
		return false, nil, errors.New("invalid signature param")
	}

	// verify initiator
	akType := acl.IsAccount(tx.Initiator)
	if akType == 0 {
		// check initiator address signature
		ok, err := pm.IdentifyAK(tx.Initiator, tx.InitiatorSigns[0], digestHash)
		if err != nil || !ok {
			uv.xlog.Warn("verifySignatures failed", "address", tx.Initiator, "error", err)
			return false, nil, err
		}
		verifiedAddr[tx.Initiator] = true
	} else if akType == 1 {
		initiatorAddr := make([]string, 0)
		// check initiator account signatures
		for _, sign := range tx.InitiatorSigns {
			ak, err := uv.cryptoClient.GetEcdsaPublicKeyFromJSON([]byte(sign.PublicKey))
			if err != nil {
				uv.xlog.Warn("verifySignatures failed", "address", tx.Initiator, "error", err)
				return false, nil, err
			}
			addr, err := uv.cryptoClient.GetAddressFromPublicKey(ak)
			if err != nil {
				uv.xlog.Warn("verifySignatures failed", "address", tx.Initiator, "error", err)
				return false, nil, err
			}
			ok, err := pm.IdentifyAK(addr, sign, digestHash)
			if !ok {
				uv.xlog.Warn("verifySignatures failed", "address", tx.Initiator, "error", err)
				return ok, nil, err
			}
			verifiedAddr[addr] = true
			initiatorAddr = append(initiatorAddr, tx.Initiator+"/"+addr)
		}
		ok, err := pm.IdentifyAccount(tx.Initiator, initiatorAddr, uv.aclMgr)
		if !ok {
			uv.xlog.Warn("verifySignatures initiator permission check failed",
				"account", tx.Initiator, "error", err)
			return false, nil, err
		}
	} else {
		uv.xlog.Warn("verifySignatures failed, invalid address", "address", tx.Initiator)
		return false, nil, ErrInvalidSignature
	}

	// verify authRequire
	for idx, authReq := range tx.AuthRequire {
		splitRes := strings.Split(authReq, "/")
		addr := splitRes[len(splitRes)-1]
		signInfo := tx.AuthRequireSigns[idx]
		if _, has := verifiedAddr[tx.Initiator]; has {
			continue
		}
		ok, err := pm.IdentifyAK(addr, signInfo, digestHash)
		if err != nil || !ok {
			uv.xlog.Warn("verifySignatures failed", "address", addr, "error", err)
			return false, nil, err
		}
		verifiedAddr[tx.Initiator] = true
	}
	return true, verifiedAddr, nil
}

func (uv *UtxoVM) verifyXuperSign(tx *pb.Transaction, digestHash []byte) (bool, map[string]bool, error) {
	uniqueAddrs := make(map[string]bool)
	// get all addresses
	uniqueAddrs[tx.Initiator] = true
	addrList := make([]string, 0)
	addrList = append(addrList, tx.Initiator)
	for _, authReq := range tx.AuthRequire {
		splitRes := strings.Split(authReq, "/")
		addr := splitRes[len(splitRes)-1]
		if uniqueAddrs[addr] {
			continue
		}
		uniqueAddrs[addr] = true
		addrList = append(addrList, addr)
	}

	// check addresses and public keys
	if len(addrList) != len(tx.GetXuperSign().GetPublicKeys()) {
		return false, nil, errors.New("XuperSign: number of address and public key not match")
	}
	pubkeys := make([]*ecdsa.PublicKey, 0)
	for _, pubJSON := range tx.GetXuperSign().GetPublicKeys() {
		pubkey, err := uv.cryptoClient.GetEcdsaPublicKeyFromJSON(pubJSON)
		if err != nil {
			return false, nil, errors.New("XuperSign: found invalid public key")
		}
		pubkeys = append(pubkeys, pubkey)
	}
	for idx, addr := range addrList {
		ok, _ := uv.cryptoClient.VerifyAddressUsingPublicKey(addr, pubkeys[idx])
		if !ok {
			uv.xlog.Warn("XuperSign: address and public key not match", "addr", addr, "pubkey", pubkeys[idx])
			return false, nil, errors.New("XuperSign: address and public key not match")
		}
	}
	ok, err := uv.cryptoClient.XuperVerify(pubkeys, tx.GetXuperSign().GetSignature(), digestHash)
	if err != nil || !ok {
		uv.xlog.Warn("XuperSign: signature verify failed", "error", err)
		return false, nil, errors.New("XuperSign: address and public key not match")
	}
	return ok, uniqueAddrs, nil
}

// verify UTXO input permission in transaction using ACL
func (uv *UtxoVM) verifyUTXOPermission(tx *pb.Transaction, verifiedID map[string]bool) (bool, error) {
	// verify tx input ACL
	for _, txInput := range tx.TxInputs {
		name := string(txInput.FromAddr)
		if verifiedID[name] {
			// this ID(either AK or Account) is verified before
			continue
		}
		akType := acl.IsAccount(name)
		if akType == 1 {
			// Identify account
			acl, err := uv.queryAccountACL(name)
			if err != nil || acl == nil {
				// valid account should have ACL info, so this account might not exsit
				uv.xlog.Warn("verifyUTXOPermission error, account might not exist", "account", name, "error", err)
				return false, ErrInvalidAccount
			}
			if ok, err := pm.IdentifyAccount(string(name), tx.AuthRequire, uv.aclMgr); !ok {
				uv.xlog.Warn("verifyUTXOPermission error, failed to IdentifyAccount", "error", err)
				return false, ErrACLNotEnough
			}
		} else if akType == 0 {
			// Identify address failed, if address not in verifiedID then it must have no signature
			uv.xlog.Warn("verifyUTXOPermission error, address has no signature", "address", name)
			return false, ErrInvalidSignature
		} else {
			uv.xlog.Warn("verifyUTXOPermission error, Invalid account/address name", "name", name)
			return false, ErrInvalidAccount
		}
		verifiedID[name] = true
	}
	return true, nil
}

// verifyContractOwnerPermission check if the transaction has the permission of a contract owner.
// this usually happens in account management operations.
func (uv *UtxoVM) verifyContractOwnerPermission(contractName string, tx *pb.Transaction,
	verifiedID map[string]bool) (bool, error) {
	versionData, err := uv.model3.Get(aclu.GetContract2AccountBucket(), []byte(contractName))
	if err != nil || versionData == nil {
		return false, err
	}
	pureData := versionData.GetPureData()
	confirmed := versionData.GetConfirmed()
	if pureData == nil || confirmed == false {
		return false, errors.New("pure data is nil or unconfirmed")
	}
	accountName := string(pureData.GetValue())
	if verifiedID[accountName] {
		return true, nil
	}
	ok, err := pm.IdentifyAccount(accountName, tx.AuthRequire, uv.aclMgr)
	if err == nil && ok {
		verifiedID[accountName] = true
	}
	return ok, err
}

// verifyRWSetPermission verify the permission of RWSet using ACL
func (uv *UtxoVM) verifyRWSetPermission(tx *pb.Transaction, verifiedID map[string]bool) (bool, error) {
	req := tx.GetContractRequests()
	// if not contract, pass directly
	if req == nil {
		return true, nil
	}
	env, err := uv.model3.PrepareEnv(tx)
	if err != nil {
		return false, err
	}
	writeSet := env.GetOutputs()
	for _, ele := range writeSet {
		bucket := ele.GetBucket()
		key := ele.GetKey()
		switch bucket {
		case aclu.GetAccountBucket():
			// modified account data, need to check if the tx has the permission of account
			accountName := string(key)
			if verifiedID[accountName] {
				continue
			}
			ok, err := pm.IdentifyAccount(accountName, tx.AuthRequire, uv.aclMgr)
			if !ok {
				uv.xlog.Warn("verifyRWSetPermission check account bucket failed",
					"account", accountName, "AuthRequire ", tx.AuthRequire, "error", err)
				return ok, err
			}
			verifiedID[accountName] = true
		case aclu.GetContractBucket():
			// modified contact data, need to check if the tx has the permission of contract owner
			separator := aclu.GetACLSeparator()
			idx := bytes.Index(key, []byte(separator))
			if idx < 0 {
				return false, errors.New("invalid raw key")
			}
			contractName := string(key[:idx])
			ok, contractErr := uv.verifyContractOwnerPermission(contractName, tx, verifiedID)
			if !ok {
				uv.xlog.Warn("verifyRWSetPermission check contract bucket failed",
					"contract", contractName, "AuthRequire ", tx.AuthRequire, "error", err)
				return ok, contractErr
			}
		case aclu.GetContract2AccountBucket():
			// modified contract/account mapping
			// need to check if the tx has the permission of target account
			accountValue := ele.GetValue()
			if accountValue == nil {
				return false, errors.New("account name is empty")
			}
			accountName := string(accountValue)
			if verifiedID[accountName] {
				continue
			}
			ok, accountErr := pm.IdentifyAccount(accountName, tx.AuthRequire, uv.aclMgr)
			if !ok {
				uv.xlog.Warn("verifyRWSetPermission check contract2account bucket failed",
					"account", accountName, "AuthRequire ", tx.AuthRequire, "error", err)
				return ok, accountErr
			}
			verifiedID[accountName] = true
		}
	}
	return true, nil
}

// verifyContractValid verify the permission of contract requests using ACL
func (uv *UtxoVM) verifyContractPermission(tx *pb.Transaction) (bool, error) {
	req := tx.GetContractRequests()
	if req == nil {
		// if no contract requests, no need to verify
		return true, nil
	}

	for i := 0; i < len(req); i++ {
		tmpReq := req[i]
		contractName := tmpReq.GetContractName()
		methodName := tmpReq.GetMethodName()

		ok, err := pm.CheckContractMethodPerm(tx.AuthRequire, contractName, methodName, uv.aclMgr)
		if err != nil || !ok {
			uv.xlog.Warn("verify contract method ACL failed ", "contract", contractName, "method",
				methodName, "error", err)
			return ok, ErrACLNotEnough
		}
	}
	return true, nil
}

func getGasLimitFromTx(tx *pb.Transaction) (int64, error) {
	for _, output := range tx.GetTxOutputs() {
		if string(output.GetToAddr()) != "$" {
			continue
		}
		gasLimit := big.NewInt(0).SetBytes(output.GetAmount()).Int64()
		// FIXME: gasLimit从大数过来的，处理溢出问题
		if gasLimit <= 0 {
			return 0, fmt.Errorf("bad gas limit %d", gasLimit)
		}
		return gasLimit, nil
	}
	return 0, nil
}

// verifyTxRWSets verify tx read sets and write sets
func (uv *UtxoVM) verifyTxRWSets(tx *pb.Transaction) (bool, error) {
	if uv.verifyReservedWhitelist(tx) {
		uv.xlog.Info("verifyReservedWhitelist true", "txid", fmt.Sprintf("%x", tx.GetTxid()))
		return true, nil
	}

	req := tx.GetContractRequests()
	reservedRequests, err := uv.getReservedContractRequests(tx.GetContractRequests(), false)
	if err != nil {
		uv.xlog.Error("getReservedContractRequests error", "error", err.Error())
		return false, err
	}

	if !uv.verifyReservedContractRequests(reservedRequests, req) {
		uv.xlog.Error("verifyReservedContractRequests error", "reservedRequests", reservedRequests, "req", req)
		return false, fmt.Errorf("verify reservedContracts error")
	}

	if req == nil {
		if tx.GetTxInputsExt() != nil || tx.GetTxOutputsExt() != nil {
			uv.xlog.Error("verifyTxRWSets error", "error", ErrInvalidTxExt.Error())
			return false, ErrInvalidTxExt
		}
		return true, nil
	}

	env, err := uv.model3.PrepareEnv(tx)
	if err != nil {
		return false, err
	}
	contextConfig := &contract.ContextConfig{
		XMCache:      env.GetModelCache(),
		Initiator:    tx.GetInitiator(),
		AuthRequire:  tx.GetAuthRequire(),
		ContractName: "",
	}
	gasLimit, err := getGasLimitFromTx(tx)
	if err != nil {
		return false, err
	}
	uv.xlog.Trace("get gas limit from tx", "gasLimit", gasLimit, "txid", hex.EncodeToString(tx.Txid))

	for i, tmpReq := range tx.GetContractRequests() {
		moduleName := tmpReq.GetModuleName()
		vm, err := uv.vmMgr3.GetVM(moduleName)
		if err != nil {
			return false, err
		}

		limits := contract.FromPbLimits(tmpReq.GetResourceLimits())
		if i >= len(reservedRequests) {
			gasLimit -= limits.TotalGas()
		}
		if gasLimit < 0 {
			uv.xlog.Error("virifyTxRWSets error:out of gas", "contractName", tmpReq.GetContractName(),
				"txid", hex.EncodeToString(tx.Txid))
			return false, errors.New("out of gas")
		}
		contextConfig.ResourceLimits = limits
		contextConfig.ContractName = tmpReq.GetContractName()
		ctx, err := vm.NewContext(contextConfig)
		if err != nil {
			// FIXME zq @icexin: need to return contract not found
			uv.xlog.Error("verifyTxRWSets NewContext error", "err", err, "contractName", tmpReq.GetContractName())
			if i < len(reservedRequests) && (err.Error() == "leveldb: not found" || strings.HasSuffix(err.Error(), "not found")) {
				continue
			}
			return false, err
		}

		_, err = ctx.Invoke(tmpReq.MethodName, tmpReq.Args)
		if err != nil {
			ctx.Release()
			uv.xlog.Error("verifyTxRWSets Invoke error", "error", err, "contractName", tmpReq.GetContractName())
			return false, err
		}

		ctx.Release()
	}

	_, writeSet, err := env.GetModelCache().GetRWSets()
	if err != nil {
		return false, err
	}
	uv.xlog.Trace("verifyTxRWSets", "env.output", env.GetOutputs(), "writeSet", writeSet)
	ok := xmodel.Equal(env.GetOutputs(), writeSet)
	if !ok {
		return false, fmt.Errorf("Verify error")
	}
	return true, nil
}

func (uv *UtxoVM) verifyMarkedTx(tx *pb.Transaction) (bool, error) {
	bytespk := []byte(tx.ModifyBlock.PublicKey)
	xcc, err := client.CreateCryptoClientFromJSONPublicKey(bytespk)
	if err != nil {
		return false, err
	}
	ecdsaKey, err := xcc.GetEcdsaPublicKeyFromJSON(bytespk)
	if err != nil {
		return false, err
	}
	isMatch, _ := xcc.VerifyAddressUsingPublicKey(uv.modifyBlockAddr, ecdsaKey)
	if !isMatch {
		return false, errors.New("address and public key not match")
	}

	bytesign, err := hex.DecodeString(tx.ModifyBlock.Sign)
	if err != nil {
		return false, fmt.Errorf("invalide arg type: sign byte")
	}
	digestHash := hash.DoubleSha256([]byte(tx.Txid))
	ok, err := xcc.VerifyECDSA(ecdsaKey, bytesign, digestHash)
	if err != nil || !ok {
		uv.xlog.Warn("validateUpdateBlockChainData verifySignatures failed")
		return false, err
	}
	return true, nil
}

// verifyRelyOnMarkedTxs
// bool bool verify
func (uv *UtxoVM) verifyRelyOnMarkedTxs(tx *pb.Transaction) (bool, error) {
	for _, txInput := range tx.GetTxInputs() {
		reftxid := txInput.RefTxid
		ok, err := uv.checkRelyOnMarkedTxid(reftxid, tx.Blockid)
		if !ok || err != nil {
			return ok, err
		}
	}
	for _, txIn := range tx.GetTxInputsExt() {
		reftxid := txIn.RefTxid
		ok, err := uv.checkRelyOnMarkedTxid(reftxid, tx.Blockid)
		if !ok || err != nil {
			return ok, err
		}
	}

	return true, nil
}

func (uv *UtxoVM) checkRelyOnMarkedTxid(reftxid []byte, blockid []byte) (bool, error) {
	if string(reftxid) == "" {
		return false, nil
	}
	reftx, err := uv.ledger.QueryTransaction(reftxid)
	if err != nil {
		return false, nil
	}
	if reftx.GetModifyBlock() != nil && reftx.ModifyBlock.Marked {
		block, err := uv.ledger.QueryBlock(blockid)
		if err != nil {
			return false, err
		}
		if block.Height >= reftx.ModifyBlock.EffectiveHeight {
			return false, nil
		}
	}
	return true, nil
}
