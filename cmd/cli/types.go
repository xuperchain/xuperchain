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

// GasPrice proto.GasPrice
type GasPrice struct {
	CpuRate  int64 `json:"cpu_rate"`
	MemRate  int64 `json:"mem_rate"`
	DiskRate int64 `json:"disk_rate"`
	XfeeRate int64 `json:"xfee_rate"`
}

// SignatureInfo proto.SignatureInfo
type SignatureInfo struct {
	PublicKey string `json:"publickey"`
	Sign      HexID  `json:"sign"`
}

// QCState is the phase of hotstuff
type QCState int32

// QCState defination
const (
	QCState_NEW_VIEW   QCState = 0
	QCState_PREPARE    QCState = 1
	QCState_PRE_COMMIT QCState = 2
	QCState_COMMIT     QCState = 3
	QCState_DECIDE     QCState = 4
)

// SignInfo is the signature information of the
type SignInfo struct {
	Address   string `protobuf:"bytes,1,opt,name=Address,proto3" json:"Address,omitempty"`
	PublicKey string `protobuf:"bytes,2,opt,name=PublicKey,proto3" json:"PublicKey,omitempty"`
	Sign      []byte `protobuf:"bytes,3,opt,name=Sign,proto3" json:"Sign,omitempty"`
}

// QCSignInfos is the signs of the leader gathered from replicas of a specifically certType.
// A slice of signs is used at present.
// TODO @qizheng09: It will be change to Threshold-Signatures after
// Crypto lib support Threshold-Signatures.
type QCSignInfos struct {
	// QCSignInfos
	QCSignInfos []*SignInfo `protobuf:"bytes,1,rep,name=QCSignInfos,proto3" json:"QCSignInfos,omitempty"`
}

// QuorumCert is a data type that combines a collection of signatures from replicas.
type QuorumCert struct {
	// The id of Proposal this QC certified.
	ProposalId string `protobuf:"bytes,1,opt,name=ProposalId,proto3" json:"ProposalId,omitempty"`
	// The msg of Proposal this QC certified.
	ProposalMsg []byte `protobuf:"bytes,2,opt,name=ProposalMsg,proto3" json:"ProposalMsg,omitempty"`
	// The current type of this QC certified.
	// the type contains `NEW_VIEW`, `PREPARE`
	Type QCState `protobuf:"varint,3,opt,name=Type,proto3,enum=pb.QCState" json:"Type,omitempty"`
	// The view number of this QC certified.
	ViewNumber int64 `protobuf:"varint,4,opt,name=ViewNumber,proto3" json:"ViewNumber,omitempty"`
	// SignInfos is the signs of the leader gathered from replicas
	// of a specifically certType.
	SignInfos *QCSignInfos `protobuf:"bytes,5,opt,name=SignInfos,proto3" json:"SignInfos,omitempty"`
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
	ReceivedTimestamp int64            `json:"receivedTimestamp"`
	ModifyBlock       ModifyBlock      `json:"modifyBlock"`
}

type ModifyBlock struct {
	Marked          bool   `json:"marked"`
	EffectiveHeight int64  `json:"effectiveHeight"`
	EffectiveTxid   string `json:"effectiveTxid"`
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

	if tx.ModifyBlock != nil {
		t.ModifyBlock = ModifyBlock{
			EffectiveHeight: tx.ModifyBlock.EffectiveHeight,
			Marked:          tx.ModifyBlock.Marked,
			EffectiveTxid:   tx.ModifyBlock.EffectiveTxid,
		}
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
	Justify      *QuorumCert       `json:"justify"`
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
	iblock.Justify = FromPBJustify(block.Justify)
	return iblock
}

// FromPBJustify use pb.QuorumCert to construct local QuorumCert in block
func FromPBJustify(qc *pb.QuorumCert) *QuorumCert {
	justify := &QuorumCert{}
	if qc != nil {
		justify.ProposalId = hex.EncodeToString(qc.ProposalId)
		justify.ProposalMsg = qc.ProposalMsg
		justify.Type = QCState(int(qc.Type))
		justify.ViewNumber = qc.ViewNumber
		justify.SignInfos = &QCSignInfos{
			QCSignInfos: make([]*SignInfo, 0),
		}
		for _, sign := range qc.SignInfos.QCSignInfos {
			tmpSign := &SignInfo{
				Address:   sign.Address,
				PublicKey: sign.PublicKey,
				Sign:      sign.Sign,
			}
			justify.SignInfos.QCSignInfos = append(justify.SignInfos.QCSignInfos, tmpSign)
		}
	}
	return justify
}

// LedgerMeta proto.LedgerMeta
type LedgerMeta struct {
	// RootBlockid root block id
	RootBlockid HexID `json:"rootBlockid"`
	// TipBlockid TipBlockid
	TipBlockid HexID `json:"tipBlockid"`
	// TrunkHeight TrunkHeight
	TrunkHeight int64 `json:"trunkHeight"`
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
	// MaxBlockSize MaxBlockSize
	MaxBlockSize int64 `json:"maxBlockSize"`
	// ReservedContracts ReservedContracts
	ReservedContracts []InvokeRequest `json:"reservedContracts"`
	// ForbiddenContract forbidden contract
	ForbiddenContract InvokeRequest `json:"forbiddenContract"`
	// NewAccountResourceAmount resource amount of creating an account
	NewAccountResourceAmount int64 `json:"newAccountResourceAmount"`
	// IrreversibleBlockHeight irreversible block height
	IrreversibleBlockHeight int64 `json:"irreversibleBlockHeight"`
	// IrreversibleSlideWindow irreversible slide window
	IrreversibleSlideWindow int64 `json:"irreversibleSlideWindow"`
	// GasPrice gas rate to utxo for different type resources
	GasPrice GasPrice `json:"gasPrice"`
}

// ChainStatus proto.ChainStatus
type ChainStatus struct {
	Name       string     `json:"name"`
	LedgerMeta LedgerMeta `json:"ledger"`
	UtxoMeta   UtxoMeta   `json:"utxo"`
	// add BranchBlockid
	BranchBlockid []string `json:"branchBlockid"`
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
		ReservedContracts := utxoMeta.GetReservedContracts()
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
		forbiddenContract := utxoMeta.GetForbiddenContract()
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
		gasPricePB := utxoMeta.GetGasPrice()
		gasPrice := GasPrice{
			CpuRate:  gasPricePB.GetCpuRate(),
			MemRate:  gasPricePB.GetMemRate(),
			DiskRate: gasPricePB.GetDiskRate(),
			XfeeRate: gasPricePB.GetXfeeRate(),
		}
		status.ChainStatus = append(status.ChainStatus, ChainStatus{
			Name: chain.GetBcname(),
			LedgerMeta: LedgerMeta{
				RootBlockid: ledgerMeta.GetRootBlockid(),
				TipBlockid:  ledgerMeta.GetTipBlockid(),
				TrunkHeight: ledgerMeta.GetTrunkHeight(),
			},
			UtxoMeta: UtxoMeta{
				LatestBlockid:            utxoMeta.GetLatestBlockid(),
				LockKeyList:              utxoMeta.GetLockKeyList(),
				UtxoTotal:                utxoMeta.GetUtxoTotal(),
				AvgDelay:                 utxoMeta.GetAvgDelay(),
				UnconfirmTxAmount:        utxoMeta.GetUnconfirmTxAmount(),
				MaxBlockSize:             utxoMeta.GetMaxBlockSize(),
				NewAccountResourceAmount: utxoMeta.GetNewAccountResourceAmount(),
				ReservedContracts:        rcs,
				ForbiddenContract:        forbiddenContractMap,
				// Irreversible block height & slide window
				IrreversibleBlockHeight: utxoMeta.GetIrreversibleBlockHeight(),
				IrreversibleSlideWindow: utxoMeta.GetIrreversibleSlideWindow(),
				// add GasPrice value
				GasPrice: gasPrice,
			},
			BranchBlockid: chain.GetBranchBlockid(),
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
