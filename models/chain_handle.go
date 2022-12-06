package models

import (
	"math/big"
	"strconv"

	lpb "github.com/xuperchain/xupercore/bcs/ledger/xledger/xldgpb"
	xCtx "github.com/xuperchain/xupercore/kernel/common/xcontext"
	"github.com/xuperchain/xupercore/kernel/engines/xuperos/common"
	"github.com/xuperchain/xupercore/kernel/engines/xuperos/reader"
	"github.com/xuperchain/xupercore/kernel/engines/xuperos/xpb"
	"github.com/xuperchain/xupercore/lib/logs"
	"github.com/xuperchain/xupercore/protos"

	sCtx "github.com/xuperchain/xuperchain/service/context"
	aclUtils "github.com/xuperchain/xupercore/kernel/permission/acl/utils"
	cryptoHash "github.com/xuperchain/xupercore/lib/crypto/hash"
)

type ChainHandle struct {
	bcName string
	reqCtx sCtx.ReqCtx
	log    logs.Logger
	chain  common.Chain
}

func NewChainHandle(bcName string, reqCtx sCtx.ReqCtx) (*ChainHandle, error) {
	if bcName == "" || reqCtx == nil || reqCtx.GetEngine() == nil {
		return nil, common.ErrParameter
	}

	chain, err := reqCtx.GetEngine().Get(bcName)
	if err != nil {
		return nil, common.ErrChainNotExist
	}

	obj := &ChainHandle{
		bcName: bcName,
		reqCtx: reqCtx,
		log:    reqCtx.GetLog(),
		chain:  chain,
	}
	return obj, nil
}

func (h *ChainHandle) SubmitTx(tx *lpb.Transaction) error {
	return h.chain.SubmitTx(h.ctx(), tx)
}

func (h *ChainHandle) PreExec(req []*protos.InvokeRequest,
	initiator string, authRequires []string) (*protos.InvokeResponse, error) {
	return h.chain.PreExec(h.ctx(), req, initiator, authRequires)
}

func (h *ChainHandle) QueryTx(txId []byte) (*xpb.TxInfo, error) {
	return h.ledgerReader().QueryTx(txId)
}

func (h *ChainHandle) SelectUtxo(account string, need *big.Int, isLock, isExclude bool,
	pubKey string, sign []byte) (*lpb.UtxoOutput, error) {
	// 如果需要临时锁定utxo，需要校验权限
	ok := h.checkSelectUtxoSign(account, pubKey, sign, isLock, need)
	if !ok {
		h.reqCtx.GetLog().Warn("select utxo verify sign failed", "account", account, "isLock", isLock)
		return nil, common.ErrUnauthorized
	}

	return h.utxoReader().SelectUTXO(account, need,
		isLock, isExclude)
}

func (h *ChainHandle) SelectUTXOBySize(account string, isLock, isExclude bool,
	pubKey string, sign []byte) (*lpb.UtxoOutput, error) {
	// 如果需要临时锁定utxo，需要校验权限
	ok := h.checkSelectUtxoSign(account, pubKey, sign, isLock, big.NewInt(0))
	if !ok {
		h.reqCtx.GetLog().Warn("select utxo verify sign failed", "account", account, "isLock", isLock)
		return nil, common.ErrUnauthorized
	}

	return h.utxoReader().SelectUTXOBySize(account, isLock, isExclude)
}

func (h *ChainHandle) QueryContractStatData() (*protos.ContractStatData, error) {
	return h.contractReader().QueryContractStatData()
}

func (h *ChainHandle) QueryUtxoRecord(account string, count int64) (*lpb.UtxoRecordDetail, error) {
	return h.utxoReader().QueryUtxoRecord(account, count)
}

func (h *ChainHandle) QueryAccountACL(account string) (*protos.Acl, error) {
	return h.contractReader().QueryAccountACL(account)
}

func (h *ChainHandle) QueryContractMethodACL(contract, method string) (*protos.Acl, error) {
	return h.contractReader().QueryContractMethodACL(contract, method)
}

func (h *ChainHandle) GetAccountContracts(account string) ([]*protos.ContractStatus, error) {
	return h.contractReader().GetAccountContracts(account)
}

func (h *ChainHandle) GetBalance(account string) (string, error) {
	return h.utxoReader().GetBalance(account)
}

func (h *ChainHandle) GetFrozenBalance(account string) (string, error) {
	return h.utxoReader().GetFrozenBalance(account)
}

func (h *ChainHandle) GetBalanceDetail(account string) ([]*lpb.BalanceDetailInfo, error) {
	return h.utxoReader().GetBalanceDetail(account)
}

func (h *ChainHandle) QueryBlock(blkId []byte, needContent bool) (*xpb.BlockInfo, error) {
	return h.ledgerReader().QueryBlock(blkId, needContent)
}

func (h *ChainHandle) QueryChainStatus() (*xpb.ChainStatus, error) {
	return h.chainReader().GetChainStatus()
}

func (h *ChainHandle) QueryConsensusStatus() (*xpb.ConsensusStatus, error) {
	return h.chainReader().GetConsensusStatus()
}

func (h *ChainHandle) IsTrunkTipBlock(blockId []byte) (bool, error) {
	return h.chainReader().IsTrunkTipBlock(blockId)
}

func (h *ChainHandle) QueryBlockByHeight(height int64, needContent bool) (*xpb.BlockInfo, error) {
	return h.ledgerReader().QueryBlockByHeight(height, needContent)
}

func (h *ChainHandle) GetAccountByAK(address string) ([]string, error) {
	return h.contractReader().GetAccountByAK(address)
}

// Helper functions

// contractReader generate a new contract reader
func (h *ChainHandle) contractReader() reader.ContractReader {
	return reader.NewContractReader(h.chain.Context(), h.ctx())
}

// ledgerReader generate a new ledger reader
func (h *ChainHandle) ledgerReader() reader.LedgerReader {
	return reader.NewLedgerReader(h.chain.Context(), h.ctx())
}

// chainReader generate a new chain reader
func (h *ChainHandle) chainReader() reader.ChainReader {
	return reader.NewChainReader(h.chain.Context(), h.ctx())
}

// utxoReader generate a new UTXO reader
func (h *ChainHandle) utxoReader() reader.UtxoReader {
	return reader.NewUtxoReader(h.chain.Context(), h.ctx())
}

func (h *ChainHandle) ctx() xCtx.XContext {
	return &xCtx.BaseCtx{
		XLog:  h.reqCtx.GetLog(),
		Timer: h.reqCtx.GetTimer(),
	}
}

func (h *ChainHandle) checkSelectUtxoSign(account, pubKey string, sign []byte,
	isLock bool, need *big.Int) bool {
	// 只对需要临时锁定utxo的校验
	if aclUtils.IsAccount(account) || !isLock {
		return true
	}

	crypto := h.chain.Context().Crypto
	publicKey, err := crypto.GetEcdsaPublicKeyFromJsonStr(pubKey)
	if err != nil {
		return false
	}

	hashStr := h.bcName + account + need.String() + strconv.FormatBool(isLock)
	doubleHash := cryptoHash.DoubleSha256([]byte(hashStr))
	checkSignResult, err := crypto.VerifyECDSA(publicKey, sign, doubleHash)
	if err != nil {
		return false
	}
	if checkSignResult != true {
		return false
	}

	matched, _ := crypto.VerifyAddressUsingPublicKey(account, publicKey)
	return matched
}
