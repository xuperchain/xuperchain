package models

import (
	"math/big"
	"strconv"

	lpb "github.com/xuperchain/xupercore/bcs/ledger/xledger/xldgpb"
	xctx "github.com/xuperchain/xupercore/kernel/common/xcontext"
	ecom "github.com/xuperchain/xupercore/kernel/engines/xuperos/common"
	"github.com/xuperchain/xupercore/kernel/engines/xuperos/reader"
	"github.com/xuperchain/xupercore/kernel/engines/xuperos/xpb"
	aclUtils "github.com/xuperchain/xupercore/kernel/permission/acl/utils"
	cryptoHash "github.com/xuperchain/xupercore/lib/crypto/hash"
	"github.com/xuperchain/xupercore/lib/logs"
	"github.com/xuperchain/xupercore/protos"

	sctx "github.com/xuperchain/xuperchain/service/context"
)

type ChainHandle struct {
	bcName string
	reqCtx sctx.ReqCtx
	log    logs.Logger
	chain  ecom.Chain
}

func NewChainHandle(bcName string, reqCtx sctx.ReqCtx) (*ChainHandle, error) {
	if bcName == "" || reqCtx == nil || reqCtx.GetEngine() == nil {
		return nil, ecom.ErrParameter
	}

	chain, err := reqCtx.GetEngine().Get(bcName)
	if err != nil {
		return nil, ecom.ErrChainNotExist
	}

	obj := &ChainHandle{
		bcName: bcName,
		reqCtx: reqCtx,
		log:    reqCtx.GetLog(),
		chain:  chain,
	}
	return obj, nil
}

func (t *ChainHandle) SubmitTx(tx *lpb.Transaction) error {
	return t.chain.SubmitTx(t.genXctx(), tx)
}

func (t *ChainHandle) PreExec(req []*protos.InvokeRequest,
	initiator string, authRequires []string) (*protos.InvokeResponse, error) {
	return t.chain.PreExec(t.genXctx(), req, initiator, authRequires)
}

func (t *ChainHandle) QueryTx(txId []byte) (*xpb.TxInfo, error) {
	return reader.NewLedgerReader(t.chain.Context(), t.genXctx()).QueryTx(txId)
}

func (t *ChainHandle) SelectUtxo(account string, need *big.Int, isLock, isExclude bool,
	pubKey string, sign []byte) (*lpb.UtxoOutput, error) {
	// 如果需要临时锁定utxo，需要校验权限
	ok := t.checkSelectUtxoSign(account, pubKey, sign, isLock, need)
	if !ok {
		t.reqCtx.GetLog().Warn("select utxo verify sign failed", "account", account, "isLock", isLock)
		return nil, ecom.ErrUnauthorized
	}

	return reader.NewUtxoReader(t.chain.Context(), t.genXctx()).SelectUTXO(account, need,
		isLock, isExclude)
}

func (t *ChainHandle) SelectUTXOBySize(account string, isLock, isExclude bool,
	pubKey string, sign []byte) (*lpb.UtxoOutput, error) {
	// 如果需要临时锁定utxo，需要校验权限
	ok := t.checkSelectUtxoSign(account, pubKey, sign, isLock, big.NewInt(0))
	if !ok {
		t.reqCtx.GetLog().Warn("select utxo verify sign failed", "account", account, "isLock", isLock)
		return nil, ecom.ErrUnauthorized
	}

	return reader.NewUtxoReader(t.chain.Context(), t.genXctx()).SelectUTXOBySize(account,
		isLock, isExclude)
}

func (t *ChainHandle) QueryContractStatData() (*protos.ContractStatData, error) {
	return reader.NewContractReader(t.chain.Context(), t.genXctx()).QueryContractStatData()
}

func (t *ChainHandle) QueryUtxoRecord(account string, count int64) (*lpb.UtxoRecordDetail, error) {
	return reader.NewUtxoReader(t.chain.Context(), t.genXctx()).QueryUtxoRecord(account, count)
}

func (t *ChainHandle) QueryAccountACL(account string) (*protos.Acl, error) {
	return reader.NewContractReader(t.chain.Context(), t.genXctx()).QueryAccountACL(account)
}

func (t *ChainHandle) QueryContractMethodACL(contract, method string) (*protos.Acl, error) {
	return reader.NewContractReader(t.chain.Context(),
		t.genXctx()).QueryContractMethodACL(contract, method)
}

func (t *ChainHandle) GetAccountContracts(account string) ([]*protos.ContractStatus, error) {
	return reader.NewContractReader(t.chain.Context(),
		t.genXctx()).GetAccountContracts(account)
}

func (t *ChainHandle) GetBalance(account string) (string, error) {
	return reader.NewUtxoReader(t.chain.Context(), t.genXctx()).GetBalance(account)
}

func (t *ChainHandle) GetFrozenBalance(account string) (string, error) {
	return reader.NewUtxoReader(t.chain.Context(), t.genXctx()).GetFrozenBalance(account)
}

func (t *ChainHandle) GetBalanceDetail(account string) ([]*lpb.BalanceDetailInfo, error) {
	return reader.NewUtxoReader(t.chain.Context(), t.genXctx()).GetBalanceDetail(account)
}

func (t *ChainHandle) QueryBlock(blkId []byte, needContent bool) (*xpb.BlockInfo, error) {
	return reader.NewLedgerReader(t.chain.Context(), t.genXctx()).QueryBlock(blkId, needContent)
}

func (t *ChainHandle) QueryChainStatus() (*xpb.ChainStatus, error) {
	return reader.NewChainReader(t.chain.Context(), t.genXctx()).GetChainStatus()
}

func (t *ChainHandle) QueryConsensusStatus() (*xpb.ConsensusStatus, error) {
	return reader.NewChainReader(t.chain.Context(), t.genXctx()).GetConsensusStatus()
}

func (t *ChainHandle) IsTrunkTipBlock(blockId []byte) (bool, error) {
	return reader.NewChainReader(t.chain.Context(), t.genXctx()).IsTrunkTipBlock(blockId)
}

func (t *ChainHandle) QueryBlockByHeight(height int64, needContent bool) (*xpb.BlockInfo, error) {
	return reader.NewLedgerReader(t.chain.Context(), t.genXctx()).QueryBlockByHeight(height, needContent)
}

func (t *ChainHandle) GetAccountByAK(address string) ([]string, error) {
	return reader.NewContractReader(t.chain.Context(), t.genXctx()).GetAccountByAK(address)
}

func (t *ChainHandle) genXctx() xctx.XContext {
	return &xctx.BaseCtx{
		XLog:  t.reqCtx.GetLog(),
		Timer: t.reqCtx.GetTimer(),
	}
}

func (t *ChainHandle) checkSelectUtxoSign(account, pubKey string, sign []byte,
	isLock bool, need *big.Int) bool {
	// 只对需要临时锁定utxo的校验
	if aclUtils.IsAccount(account) == 1 || !isLock {
		return true
	}

	crypto := t.chain.Context().Crypto
	publicKey, err := crypto.GetEcdsaPublicKeyFromJsonStr(pubKey)
	if err != nil {
		return false
	}

	hashStr := t.bcName + account + need.String() + strconv.FormatBool(isLock)
	doubleHash := cryptoHash.DoubleSha256([]byte(hashStr))
	checkSignResult, err := crypto.VerifyECDSA(publicKey, sign, doubleHash)
	if err != nil {
		return false
	}
	if checkSignResult != true {
		return false
	}
	addrMatchCheckResult, _ := crypto.VerifyAddressUsingPublicKey(account, publicKey)
	if addrMatchCheckResult != true {
		return false
	}

	return true
}
