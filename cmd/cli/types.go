/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import (
	"encoding/hex"
	"encoding/json"
	"math/big"

	"github.com/xuperchain/xuperunion/pb"
)

// HexID bytes
type HexID []byte

// MarshalJSON json marshal
func (h HexID) MarshalJSON() ([]byte, error) {
	hex := hex.EncodeToString(h)
	return json.Marshal(hex)
}

// TxInput proto.TxInput
type TxInput struct {
	RefTxid   HexID  `json:"refTxid"`
	RefOffset int32  `json:"refOffset"`
	FromAddr  string `json:"fromAddr"`
	Amount    BigInt `json:"amount"`
}

// TxOutput proto.TxOutput
type TxOutput struct {
	Amount BigInt `json:"amount"`
	ToAddr string `json:"toAddr"`
}

// TxInputExt proto.TxInputExt
type TxInputExt struct {
	Bucket    string `json:"bucket"`
	Key       string `json:"key"`
	RefTxid   HexID  `json:"refTxid"`
	RefOffset int32  `json:"refOffset"`
}

// TxOutputExt proto.TxOutputExt
type TxOutputExt struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
	Value  string `json:"value"`
}

type ResourceLimit struct {
	Type  string `json:"type"`
	Limit int64  `json:"limit"`
}

// InvokeRequest proto.InvokeRequest
type InvokeRequest struct {
	ModuleName    string            `json:"moduleName"`
	ContractName  string            `json:"contractName"`
	MethodName    string            `json:"methodName"`
	Args          map[string]string `json:"args"`
	ResouceLimits []ResourceLimit   `json:"resource_limits"`
}

// SignatureInfo proto.SignatureInfo
type SignatureInfo struct {
	PublicKey string `json:"publickey"`
	Sign      HexID  `json:"sign"`
}

// Transaction proto.Transaction
type Transaction struct {
	Txid              HexID            `json:"txid"`
	Blockid           HexID            `json:"blockid"`
	TxInputs          []TxInput        `json:"txInputs"`
	TxOutputs         []TxOutput       `json:"txOutputs"`
	Desc              string           `json:"desc"`
	Nonce             string           `json:"nonce"`
	Timestamp         int64            `json:"timestamp"`
	Version           int32            `json:"version"`
	Autogen           bool             `json:"autogen"`
	Coinbase          bool             `json:"coinbase"`
	TxInputsExt       []TxInputExt     `json:"txInputsExt"`
	TxOutputsExt      []TxOutputExt    `json:"txOutputsExt"`
	ContractRequests  []*InvokeRequest `json:"contractRequests"`
	Initiator         string           `json:"initiator"`
	AuthRequire       []string         `json:"authRequire"`
	InitiatorSigns    []SignatureInfo  `json:"initiatorSigns"`
	AuthRequireSigns  []SignatureInfo  `json:"authRequireSigns"`
	ReceivedTimestamp int64            `json:"receivedTimestamp:"`
}

// BigInt big int
type BigInt big.Int

// FromAmountBytes transfer bytes to bigint
func FromAmountBytes(buf []byte) BigInt {
	n := big.Int{}
	n.SetBytes(buf)
	return BigInt(n)
}

// MarshalJSON json.marshal
func (b *BigInt) MarshalJSON() ([]byte, error) {
	str := (*big.Int)(b).String()
	return json.Marshal(str)
}

// FromPBTx get tx
func FromPBTx(tx *pb.Transaction) *Transaction {
	t := &Transaction{
		Txid:              tx.Txid,
		Blockid:           tx.Blockid,
		Nonce:             tx.Nonce,
		Timestamp:         tx.Timestamp,
		Version:           tx.Version,
		Desc:              string(tx.Desc),
		Autogen:           tx.Autogen,
		Coinbase:          tx.Coinbase,
		Initiator:         tx.Initiator,
		ReceivedTimestamp: tx.ReceivedTimestamp,
	}
	for _, input := range tx.TxInputs {
		t.TxInputs = append(t.TxInputs, TxInput{
			RefTxid:   input.RefTxid,
			RefOffset: input.RefOffset,
			FromAddr:  string(input.FromAddr),
			Amount:    FromAmountBytes(input.Amount),
		})
	}
	for _, output := range tx.TxOutputs {
		t.TxOutputs = append(t.TxOutputs, TxOutput{
			Amount: FromAmountBytes(output.Amount),
			ToAddr: string(output.ToAddr),
		})
	}
	for _, inputExt := range tx.TxInputsExt {
		t.TxInputsExt = append(t.TxInputsExt, TxInputExt{
			Bucket:    inputExt.Bucket,
			Key:       string(inputExt.Key),
			RefTxid:   inputExt.RefTxid,
			RefOffset: inputExt.RefOffset,
		})
	}
	for _, outputExt := range tx.TxOutputsExt {
		t.TxOutputsExt = append(t.TxOutputsExt, TxOutputExt{
			Bucket: outputExt.Bucket,
			Key:    string(outputExt.Key),
			Value:  string(outputExt.Value),
		})
	}
	if tx.ContractRequests != nil {
		for i := 0; i < len(tx.ContractRequests); i++ {
			req := tx.ContractRequests[i]
			tmpReq := &InvokeRequest{
				ModuleName:   req.ModuleName,
				ContractName: req.ContractName,
				MethodName:   req.MethodName,
				Args:         map[string]string{},
			}
			for argKey, argV := range req.Args {
				tmpReq.Args[argKey] = string(argV)
			}
			for _, rlimit := range req.ResourceLimits {
				resource := ResourceLimit{
					Type:  rlimit.Type.String(),
					Limit: rlimit.Limit,
				}
				tmpReq.ResouceLimits = append(tmpReq.ResouceLimits, resource)
			}
			t.ContractRequests = append(t.ContractRequests, tmpReq)
		}
	}

	t.AuthRequire = append(t.AuthRequire, tx.AuthRequire...)

	for _, initsign := range tx.InitiatorSigns {
		t.InitiatorSigns = append(t.InitiatorSigns, SignatureInfo{
			PublicKey: initsign.PublicKey,
			Sign:      initsign.Sign,
		})
	}

	for _, authSign := range tx.AuthRequireSigns {
		t.AuthRequireSigns = append(t.AuthRequireSigns, SignatureInfo{
			PublicKey: authSign.PublicKey,
			Sign:      authSign.Sign,
		})
	}

	return t
}

// InternalBlock proto.InternalBlock
type InternalBlock struct {
	Version      int32             `json:"version"`
	Blockid      HexID             `json:"blockid"`
	PreHash      HexID             `json:"preHash"`
	Proposer     string            `json:"proposer"`
	Sign         HexID             `json:"sign"`
	Pubkey       string            `json:"pubkey"`
	MerkleRoot   HexID             `json:"merkleRoot"`
	Height       int64             `json:"height"`
	Timestamp    int64             `json:"timestamp"`
	Transactions []*Transaction    `json:"transactions"`
	TxCount      int32             `json:"txCount"`
	MerkleTree   []HexID           `json:"merkleTree"`
	InTrunk      bool              `json:"inTrunk"`
	NextHash     HexID             `json:"nextHash"`
	FailedTxs    map[string]string `json:"failedTxs"`
	CurTerm      int64             `json:"curTerm"`
	CurBlockNum  int64             `json:"curBlockNum"`
}

// FromInternalBlockPB block info
func FromInternalBlockPB(block *pb.InternalBlock) *InternalBlock {
	iblock := &InternalBlock{
		Version:     block.Version,
		Blockid:     block.Blockid,
		PreHash:     block.PreHash,
		Proposer:    string(block.Proposer),
		Sign:        block.Sign,
		Pubkey:      string(block.Pubkey),
		MerkleRoot:  block.MerkleRoot,
		Height:      block.Height,
		Timestamp:   block.Timestamp,
		TxCount:     block.TxCount,
		InTrunk:     block.InTrunk,
		NextHash:    block.NextHash,
		FailedTxs:   block.FailedTxs,
		CurTerm:     block.CurTerm,
		CurBlockNum: block.CurBlockNum,
	}
	iblock.MerkleTree = make([]HexID, len(block.MerkleTree))
	for i := range block.MerkleTree {
		iblock.MerkleTree[i] = block.MerkleTree[i]
	}
	iblock.Transactions = make([]*Transaction, len(block.Transactions))
	for i := range block.Transactions {
		iblock.Transactions[i] = FromPBTx(block.Transactions[i])
	}
	return iblock
}

// LedgerMeta proto.LedgerMeta
type LedgerMeta struct {
	// RootBlockid root block id
	RootBlockid HexID `json:"rootBlockid"`
	// TipBlockid TipBlockid
	TipBlockid HexID `json:"tipBlockid"`
	// TrunkHeight TrunkHeight
	TrunkHeight int64 `json:"trunkHeight"`
	// MaxBlockSize MaxBlockSize
	MaxBlockSize int64 `json:"maxBlockSize"`
	// ReservedContracts ReservedContracts
	ReservedContracts []InvokeRequest `json:"reservedContracts"`
	// ForbiddenContract forbidden contract
	ForbiddenContract InvokeRequest `json:"forbiddenContract"`
}

// UtxoMeta proto.UtxoMeta
type UtxoMeta struct {
	// LatestBlockid LatestBlockid
	LatestBlockid HexID `json:"latestBlockid"`
	// LockKeyList LockKeyList
	LockKeyList []string `json:"lockKeyList"`
	// UtxoTotal UtxoTotal
	UtxoTotal string `json:"utxoTotal"`
	// Average confirmed dealy (ms)
	AvgDelay int64 `json:"avgDelay"`
	// Current unconfirmed tx amount
	UnconfirmTxAmount int64 `json:"unconfirmed"`
}

// ChainStatus proto.ChainStatus
type ChainStatus struct {
	Name       string     `json:"name"`
	LedgerMeta LedgerMeta `json:"ledger"`
	UtxoMeta   UtxoMeta   `json:"utxo"`
}

// SystemStatus proto.SystemStatus
type SystemStatus struct {
	ChainStatus []ChainStatus `json:"blockchains"`
	Peers       []string      `json:"peers"`
	Speeds      *pb.Speeds    `json:"speeds"`
}

// FromSystemStatusPB systemstatus info
func FromSystemStatusPB(statuspb *pb.SystemsStatus) *SystemStatus {
	status := &SystemStatus{}
	for _, chain := range statuspb.GetBcsStatus() {
		ledgerMeta := chain.GetMeta()
		utxoMeta := chain.GetUtxoMeta()
		ReservedContracts := ledgerMeta.GetReservedContracts()
		rcs := []InvokeRequest{}
		for _, rcpb := range ReservedContracts {
			args := map[string]string{}
			for k, v := range rcpb.GetArgs() {
				args[k] = string(v)
			}
			rc := InvokeRequest{
				ModuleName:   rcpb.GetModuleName(),
				ContractName: rcpb.GetContractName(),
				MethodName:   rcpb.GetMethodName(),
				Args:         args,
			}
			rcs = append(rcs, rc)
		}

		forbiddenContract := ledgerMeta.GetForbiddenContract()
		args := forbiddenContract.GetArgs()
		originalArgs := map[string]string{}
		for key, value := range args {
			originalArgs[key] = string(value)
		}
		forbiddenContractMap := InvokeRequest{
			ModuleName:   forbiddenContract.GetModuleName(),
			ContractName: forbiddenContract.GetContractName(),
			MethodName:   forbiddenContract.GetMethodName(),
			Args:         originalArgs,
		}

		status.ChainStatus = append(status.ChainStatus, ChainStatus{
			Name: chain.GetBcname(),
			LedgerMeta: LedgerMeta{
				RootBlockid:       ledgerMeta.GetRootBlockid(),
				TipBlockid:        ledgerMeta.GetTipBlockid(),
				TrunkHeight:       ledgerMeta.GetTrunkHeight(),
				MaxBlockSize:      ledgerMeta.GetMaxBlockSize(),
				ReservedContracts: rcs,
				ForbiddenContract: forbiddenContractMap,
			},
			UtxoMeta: UtxoMeta{
				LatestBlockid:     utxoMeta.GetLatestBlockid(),
				LockKeyList:       utxoMeta.GetLockKeyList(),
				UtxoTotal:         utxoMeta.GetUtxoTotal(),
				AvgDelay:          utxoMeta.GetAvgDelay(),
				UnconfirmTxAmount: utxoMeta.GetUnconfirmTxAmount(),
			},
		})
	}
	status.Peers = statuspb.GetPeerUrls()
	status.Speeds = statuspb.GetSpeeds()
	return status
}

// TriggerDesc proto.TriggerDesc
type TriggerDesc struct {
	Module string      `json:"module"`
	Method string      `json:"method"`
	Args   interface{} `json:"args"`
	Height int64       `json:"height"`
}

// ContractDesc proto.ContractDesc
type ContractDesc struct {
	Module  string      `json:"module"`
	Method  string      `json:"method"`
	Args    interface{} `json:"args"`
	Trigger TriggerDesc `json:"trigger"`
}
