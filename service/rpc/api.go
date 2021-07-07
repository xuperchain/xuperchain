package rpc

import (
	"context"
	"math/big"

	"github.com/xuperchain/xuperchain/models"
	acom "github.com/xuperchain/xuperchain/service/common"
	"github.com/xuperchain/xuperchain/service/pb"
	sctx "github.com/xuperchain/xupercore/example/xchain/common/context"
	ecom "github.com/xuperchain/xupercore/kernel/engines/xuperos/common"
	"github.com/xuperchain/xupercore/kernel/network/p2p"
	"github.com/xuperchain/xupercore/lib/utils"
	"github.com/xuperchain/xupercore/protos"
)

// 注意：
// 1.rpc接口响应resp不能为nil，必须实例化
// 2.rpc接口响应err必须为ecom.Error类型的标准错误，没有错误响应err=nil
// 3.rpc接口不需要关注resp.Header，由拦截器根据err统一设置
// 4.rpc接口可以调用log库提供的SetInfoField方法附加输出到ending log

// PostTx post transaction to blockchain network
func (t *RpcServ) PostTx(gctx context.Context, req *pb.TxStatus) (*pb.CommonReply, error) {
	// 默认响应
	resp := &pb.CommonReply{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	// 校验参数
	if req == nil || req.GetTx() == nil || req.GetBcname() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}
	tx := acom.TxToXledger(req.GetTx())
	if tx == nil {
		rctx.GetLog().Warn("param error,tx convert to xledger tx failed")
		return resp, ecom.ErrParameter
	}

	// 提交交易
	handle, err := models.NewChainHandle(req.GetBcname(), rctx)
	if err != nil {
		rctx.GetLog().Warn("new chain handle failed", "err", err.Error())
		return resp, err
	}

	err = handle.SubmitTx(tx)
	if err == nil {
		msg := p2p.NewMessage(protos.XuperMessage_POSTTX, tx,
			p2p.WithBCName(req.GetBcname()),
			p2p.WithLogId(rctx.GetLog().GetLogId()),
		)
		go t.engine.Context().Net.SendMessage(rctx, msg)
	}
	rctx.GetLog().SetInfoField("bc_name", req.GetBcname())
	rctx.GetLog().SetInfoField("txid", utils.F(req.GetTxid()))
	return resp, err
}

// PreExec smart contract preExec process
func (t *RpcServ) PreExec(gctx context.Context, req *pb.InvokeRPCRequest) (*pb.InvokeRPCResponse, error) {
	// 默认响应
	resp := &pb.InvokeRPCResponse{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	// 校验参数
	if req == nil || req.GetBcname() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}
	reqs, err := acom.ConvertInvokeReq(req.GetRequests())
	if err != nil {
		rctx.GetLog().Warn("param error, convert failed", "err", err)
		return resp, ecom.ErrParameter
	}

	// 预执行
	handle, err := models.NewChainHandle(req.GetBcname(), rctx)
	if err != nil {
		rctx.GetLog().Warn("new chain handle failed", "err", err.Error())
		return resp, err
	}
	res, err := handle.PreExec(reqs, req.GetInitiator(), req.GetAuthRequire())
	rctx.GetLog().SetInfoField("bc_name", req.GetBcname())
	rctx.GetLog().SetInfoField("initiator", req.GetInitiator())
	// 设置响应
	if err == nil {
		resp.Bcname = req.GetBcname()
		resp.Response = acom.ConvertInvokeResp(res)
	}

	return resp, err
}

// PreExecWithSelectUTXO preExec + selectUtxo
func (t *RpcServ) PreExecWithSelectUTXO(gctx context.Context,
	req *pb.PreExecWithSelectUTXORequest) (*pb.PreExecWithSelectUTXOResponse, error) {

	// 默认响应
	resp := &pb.PreExecWithSelectUTXOResponse{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetBcname() == "" || req.GetRequest() == nil {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	// PreExec
	preExecRes, err := t.PreExec(gctx, req.GetRequest())
	if err != nil {
		rctx.GetLog().Warn("pre exec failed", "err", err)
		return resp, err
	}

	// SelectUTXO
	totalAmount := req.GetTotalAmount() + preExecRes.GetResponse().GetGasUsed()
	if totalAmount < 1 {
		return resp, nil
	}
	utxoInput := &pb.UtxoInput{
		Header:    req.GetHeader(),
		Bcname:    req.GetBcname(),
		Address:   req.GetAddress(),
		Publickey: req.GetSignInfo().GetPublicKey(),
		TotalNeed: big.NewInt(totalAmount).String(),
		UserSign:  req.GetSignInfo().GetSign(),
		NeedLock:  req.GetNeedLock(),
	}
	utxoOut, err := t.SelectUTXO(gctx, utxoInput)
	if err != nil {
		return resp, err
	}
	utxoOut.Header = req.GetHeader()

	// 设置响应
	resp.Bcname = req.GetBcname()
	resp.Response = preExecRes.GetResponse()
	resp.UtxoOutput = utxoOut

	return resp, nil
}

// SelectUTXO select utxo inputs depending on amount
func (t *RpcServ) SelectUTXO(gctx context.Context, req *pb.UtxoInput) (*pb.UtxoOutput, error) {
	// 默认响应
	resp := &pb.UtxoOutput{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetBcname() == "" || req.GetTotalNeed() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}
	totalNeed, ok := new(big.Int).SetString(req.GetTotalNeed(), 10)
	if !ok {
		rctx.GetLog().Warn("param error,total need set error", "totalNeed", req.GetTotalNeed())
		return resp, ecom.ErrParameter
	}

	// select utxo
	handle, err := models.NewChainHandle(req.GetBcname(), rctx)
	if err != nil {
		rctx.GetLog().Warn("new chain handle failed", "err", err.Error())
		return resp, err
	}
	out, err := handle.SelectUtxo(req.GetAddress(), totalNeed, req.GetNeedLock(), false,
		req.GetPublickey(), req.GetUserSign())
	if err != nil {
		rctx.GetLog().Warn("select utxo failed", "err", err.Error())
		return resp, err
	}

	utxoList, err := acom.UtxoListToXchain(out.GetUtxoList())
	if err != nil {
		rctx.GetLog().Warn("convert utxo failed", "err", err)
		return resp, ecom.ErrInternal
	}

	resp.UtxoList = utxoList
	resp.TotalSelected = out.GetTotalSelected()
	return resp, nil
}

// SelectUTXOBySize select utxo inputs depending on size
func (t *RpcServ) SelectUTXOBySize(gctx context.Context, req *pb.UtxoInput) (*pb.UtxoOutput, error) {
	// 默认响应
	resp := &pb.UtxoOutput{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetBcname() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	// select utxo
	handle, err := models.NewChainHandle(req.GetBcname(), rctx)
	if err != nil {
		rctx.GetLog().Warn("new chain handle failed", "err", err.Error())
		return resp, err
	}
	out, err := handle.SelectUTXOBySize(req.GetAddress(), req.GetNeedLock(), false,
		req.GetPublickey(), req.GetUserSign())
	if err != nil {
		rctx.GetLog().Warn("select utxo failed", "err", err.Error())
		return resp, err
	}

	utxoList, err := acom.UtxoListToXchain(out.GetUtxoList())
	if err != nil {
		rctx.GetLog().Warn("convert utxo failed", "err", err)
		return resp, ecom.ErrInternal
	}

	resp.UtxoList = utxoList
	resp.TotalSelected = out.GetTotalSelected()
	return resp, nil
}

// QueryContractStatData query statistic info about contract
func (t *RpcServ) QueryContractStatData(gctx context.Context,
	req *pb.ContractStatDataRequest) (*pb.ContractStatDataResponse, error) {
	// 默认响应
	resp := &pb.ContractStatDataResponse{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetBcname() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	handle, err := models.NewChainHandle(req.GetBcname(), rctx)
	if err != nil {
		rctx.GetLog().Warn("new chain handle failed", "err", err.Error())
		return resp, err
	}
	res, err := handle.QueryContractStatData()
	if err != nil {
		rctx.GetLog().Warn("query contract stat data failed", "err", err.Error())
		return resp, err
	}

	resp.Bcname = req.GetBcname()
	resp.Data = &pb.ContractStatData{
		AccountCount:  res.GetAccountCount(),
		ContractCount: res.GetContractCount(),
	}

	return resp, nil
}

// QueryUtxoRecord query utxo records
func (t *RpcServ) QueryUtxoRecord(gctx context.Context,
	req *pb.UtxoRecordDetail) (*pb.UtxoRecordDetail, error) {

	// 默认响应
	resp := &pb.UtxoRecordDetail{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetBcname() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	handle, err := models.NewChainHandle(req.GetBcname(), rctx)
	if err != nil {
		rctx.GetLog().Warn("new chain handle failed", "err", err.Error())
		return resp, err
	}
	res, err := handle.QueryUtxoRecord(req.GetAccountName(), req.GetDisplayCount())
	if err != nil {
		rctx.GetLog().Warn("query utxo record failed", "err", err.Error())
		return resp, err
	}

	resp.Bcname = req.GetBcname()
	resp.AccountName = req.GetAccountName()
	resp.OpenUtxoRecord = acom.UtxoRecordToXchain(res.GetOpenUtxo())
	resp.LockedUtxoRecord = acom.UtxoRecordToXchain(res.GetLockedUtxo())
	resp.FrozenUtxoRecord = acom.UtxoRecordToXchain(res.GetFrozenUtxo())
	resp.DisplayCount = req.GetDisplayCount()

	return resp, nil
}

// QueryACL query some account info
func (t *RpcServ) QueryACL(gctx context.Context, req *pb.AclStatus) (*pb.AclStatus, error) {
	// 默认响应
	resp := &pb.AclStatus{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetBcname() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}
	if len(req.GetAccountName()) < 1 && (len(req.GetContractName()) < 1 || len(req.GetMethodName()) < 1) {
		rctx.GetLog().Warn("param error,unset name")
		return resp, ecom.ErrParameter
	}

	handle, err := models.NewChainHandle(req.GetBcname(), rctx)
	if err != nil {
		rctx.GetLog().Warn("new chain handle failed", "err", err.Error())
		return resp, err
	}

	var aclRes *protos.Acl
	if len(req.GetAccountName()) > 0 {
		aclRes, err = handle.QueryAccountACL(req.GetAccountName())
	} else if len(req.GetContractName()) > 0 && len(req.GetMethodName()) > 0 {
		aclRes, err = handle.QueryContractMethodACL(req.GetContractName(), req.GetMethodName())
	}
	if err != nil {
		rctx.GetLog().Warn("query acl failed", "err", err)
		return resp, err
	}

	if aclRes == nil {
		resp.Confirmed = false
		return resp, nil
	}

	xchainAcl := acom.AclToXchain(aclRes)
	if xchainAcl == nil {
		rctx.GetLog().Warn("convert acl failed")
		return resp, ecom.ErrInternal
	}

	resp.Bcname = req.GetBcname()
	resp.AccountName = req.GetAccountName()
	resp.ContractName = req.GetContractName()
	resp.MethodName = req.GetMethodName()
	resp.Confirmed = true
	resp.Acl = xchainAcl

	return resp, nil
}

// GetAccountContracts get account request
func (t *RpcServ) GetAccountContracts(gctx context.Context, req *pb.GetAccountContractsRequest) (*pb.GetAccountContractsResponse, error) {
	// 默认响应
	resp := &pb.GetAccountContractsResponse{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetBcname() == "" || req.GetAccount() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	handle, err := models.NewChainHandle(req.GetBcname(), rctx)
	if err != nil {
		rctx.GetLog().Warn("new chain handle failed", "err", err.Error())
		return resp, err
	}

	var res []*protos.ContractStatus
	res, err = handle.GetAccountContracts(req.GetAccount())
	if err != nil {
		rctx.GetLog().Warn("get account contract failed", "err", err)
		return resp, err
	}
	xchainContractStatus, err := acom.ContractStatusListToXchain(res)
	if xchainContractStatus == nil {
		rctx.GetLog().Warn("convert acl failed")
		return resp, ecom.ErrInternal
	}

	resp.ContractsStatus = xchainContractStatus

	rctx.GetLog().SetInfoField("bc_name", req.GetBcname())
	rctx.GetLog().SetInfoField("account", req.GetAccount())
	return resp, nil
}

// QueryTx Get transaction details
func (t *RpcServ) QueryTx(gctx context.Context, req *pb.TxStatus) (*pb.TxStatus, error) {
	// 默认响应
	resp := &pb.TxStatus{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetBcname() == "" || len(req.GetTxid()) == 0 {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	handle, err := models.NewChainHandle(req.GetBcname(), rctx)
	if err != nil {
		rctx.GetLog().Warn("new chain handle failed", "err", err)
		return resp, ecom.ErrInternal.More("%v", err)
	}

	txInfo, err := handle.QueryTx(req.GetTxid())
	if err != nil {
		rctx.GetLog().Warn("query tx failed", "err", err)
		return resp, err
	}

	tx := acom.TxToXchain(txInfo.Tx)
	if tx == nil {
		rctx.GetLog().Warn("convert tx failed")
		return resp, ecom.ErrInternal
	}
	resp.Bcname = req.GetBcname()
	resp.Txid = req.GetTxid()
	resp.Tx = tx
	resp.Status = pb.TransactionStatus(txInfo.Status)
	resp.Distance = txInfo.Distance

	rctx.GetLog().SetInfoField("bc_name", req.GetBcname())
	rctx.GetLog().SetInfoField("account", utils.F(req.GetTxid()))
	return resp, nil
}

// GetBalance get balance for account or addr
func (t *RpcServ) GetBalance(gctx context.Context, req *pb.AddressStatus) (*pb.AddressStatus, error) {
	// 默认响应
	resp := &pb.AddressStatus{
		Bcs: make([]*pb.TokenDetail, 0),
	}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetAddress() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	for i := 0; i < len(req.Bcs); i++ {
		tmpTokenDetail := &pb.TokenDetail{
			Bcname: req.Bcs[i].Bcname,
		}
		handle, err := models.NewChainHandle(req.Bcs[i].Bcname, rctx)
		if err != nil {
			tmpTokenDetail.Error = pb.XChainErrorEnum_BLOCKCHAIN_NOTEXIST
			tmpTokenDetail.Balance = ""
			resp.Bcs = append(resp.Bcs, tmpTokenDetail)
			continue
		}
		balance, err := handle.GetBalance(req.Address)
		if err != nil {
			tmpTokenDetail.Error = pb.XChainErrorEnum_UNKNOW_ERROR
			tmpTokenDetail.Balance = ""
		} else {
			tmpTokenDetail.Error = pb.XChainErrorEnum_SUCCESS
			tmpTokenDetail.Balance = balance
		}
		resp.Bcs = append(resp.Bcs, tmpTokenDetail)
	}
	resp.Address = req.GetAddress()

	rctx.GetLog().SetInfoField("account", req.GetAddress())
	return resp, nil
}

// GetFrozenBalance get balance frozened for account or addr
func (t *RpcServ) GetFrozenBalance(gctx context.Context, req *pb.AddressStatus) (*pb.AddressStatus, error) {
	// 默认响应
	resp := &pb.AddressStatus{
		Bcs: make([]*pb.TokenDetail, 0),
	}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetAddress() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	for i := 0; i < len(req.Bcs); i++ {
		tmpTokenDetail := &pb.TokenDetail{
			Bcname: req.Bcs[i].Bcname,
		}
		handle, err := models.NewChainHandle(req.Bcs[i].Bcname, rctx)
		if err != nil {
			tmpTokenDetail.Error = pb.XChainErrorEnum_BLOCKCHAIN_NOTEXIST
			tmpTokenDetail.Balance = ""
			resp.Bcs = append(resp.Bcs, tmpTokenDetail)
			continue
		}
		balance, err := handle.GetFrozenBalance(req.Address)
		if err != nil {
			tmpTokenDetail.Error = pb.XChainErrorEnum_UNKNOW_ERROR
			tmpTokenDetail.Balance = ""
		} else {
			tmpTokenDetail.Error = pb.XChainErrorEnum_SUCCESS
			tmpTokenDetail.Balance = balance
		}
		resp.Bcs = append(resp.Bcs, tmpTokenDetail)
	}
	resp.Address = req.GetAddress()

	rctx.GetLog().SetInfoField("account", req.GetAddress())
	return resp, nil
}

// GetBalanceDetail get balance frozened for account or addr
func (t *RpcServ) GetBalanceDetail(gctx context.Context, req *pb.AddressBalanceStatus) (*pb.AddressBalanceStatus, error) {
	// 默认响应
	resp := &pb.AddressBalanceStatus{
		Tfds: make([]*pb.TokenFrozenDetails, 0),
	}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetAddress() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	for i := 0; i < len(req.Tfds); i++ {
		tmpFrozenDetails := &pb.TokenFrozenDetails{
			Bcname: req.Tfds[i].Bcname,
		}
		handle, err := models.NewChainHandle(req.Tfds[i].Bcname, rctx)
		if err != nil {
			tmpFrozenDetails.Error = pb.XChainErrorEnum_BLOCKCHAIN_NOTEXIST
			tmpFrozenDetails.Tfd = nil
			resp.Tfds = append(resp.Tfds, tmpFrozenDetails)
			continue
		}
		tfd, err := handle.GetBalanceDetail(req.GetAddress())
		if err != nil {
			tmpFrozenDetails.Error = pb.XChainErrorEnum_UNKNOW_ERROR
			tmpFrozenDetails.Tfd = nil
		} else {
			xchainTfd, err := acom.BalanceDetailsToXchain(tfd)
			if err != nil {
				tmpFrozenDetails.Error = pb.XChainErrorEnum_UNKNOW_ERROR
				tmpFrozenDetails.Tfd = nil
			}
			tmpFrozenDetails.Error = pb.XChainErrorEnum_SUCCESS
			tmpFrozenDetails.Tfd = xchainTfd
		}
		resp.Tfds = append(resp.Tfds, tmpFrozenDetails)
	}
	resp.Address = req.GetAddress()

	rctx.GetLog().SetInfoField("account", req.GetAddress())
	return resp, nil
}

// GetBlock get block info according to blockID
func (t *RpcServ) GetBlock(gctx context.Context, req *pb.BlockID) (*pb.Block, error) {
	// 默认响应
	resp := &pb.Block{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetBcname() == "" || len(req.GetBlockid()) == 0 {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	handle, err := models.NewChainHandle(req.GetBcname(), rctx)
	if err != nil {
		rctx.GetLog().Warn("new chain handle failed", "err", err.Error())
		return resp, err
	}

	blockInfo, err := handle.QueryBlock(req.GetBlockid(), true)
	if err != nil {
		rctx.GetLog().Warn("query block error", "error", err)
		return resp, err
	}

	block := acom.BlockToXchain(blockInfo.Block)
	if block == nil {
		rctx.GetLog().Warn("convert block failed")
		return resp, ecom.ErrInternal
	}

	resp.Block = block
	resp.Status = pb.Block_EBlockStatus(blockInfo.Status)
	resp.Bcname = req.Bcname
	resp.Blockid = req.Blockid

	rctx.GetLog().SetInfoField("blockid", req.GetBlockid())
	rctx.GetLog().SetInfoField("height", blockInfo.GetBlock().GetHeight())
	return resp, nil
}

// GetBlockChainStatus get systemstatus
func (t *RpcServ) GetBlockChainStatus(gctx context.Context, req *pb.BCStatus) (*pb.BCStatus, error) {
	// 默认响应
	resp := &pb.BCStatus{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetBcname() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	handle, err := models.NewChainHandle(req.GetBcname(), rctx)
	if err != nil {
		rctx.GetLog().Warn("new chain handle failed", "err", err.Error())
		return resp, err
	}

	status, err := handle.QueryChainStatus()
	if err != nil {
		rctx.GetLog().Warn("get chain status error", "error", err)
		return resp, err
	}

	block := acom.BlockToXchain(status.Block)
	if block == nil {
		rctx.GetLog().Warn("convert block failed")
		return resp, err
	}
	ledgerMeta := acom.LedgerMetaToXchain(status.LedgerMeta)
	if ledgerMeta == nil {
		rctx.GetLog().Warn("convert ledger meta failed")
		return resp, err
	}
	utxoMeta := acom.UtxoMetaToXchain(status.UtxoMeta)
	if utxoMeta == nil {
		rctx.GetLog().Warn("convert utxo meta failed")
		return resp, err
	}
	resp.Bcname = req.Bcname
	resp.Meta = ledgerMeta
	resp.Block = block
	resp.UtxoMeta = utxoMeta
	resp.BranchBlockid = status.BranchIds

	rctx.GetLog().SetInfoField("bc_name", req.GetBcname())
	rctx.GetLog().SetInfoField("blockid", utils.F(resp.Block.Blockid))
	return resp, nil
}

// ConfirmBlockChainStatus confirm is_trunk
func (t *RpcServ) ConfirmBlockChainStatus(gctx context.Context, req *pb.BCStatus) (*pb.BCTipStatus, error) {
	// 默认响应
	resp := &pb.BCTipStatus{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetBcname() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	handle, err := models.NewChainHandle(req.GetBcname(), rctx)
	if err != nil {
		rctx.GetLog().Warn("new chain handle failed", "err", err.Error())
		return resp, err
	}

	isTrunkTip, err := handle.IsTrunkTipBlock(req.GetBlock().GetBlockid())
	if err != nil {
		rctx.GetLog().Warn("query is trunk tip block fail", "err", err.Error())
		return resp, err
	}

	resp.IsTrunkTip = isTrunkTip
	rctx.GetLog().SetInfoField("blockid", utils.F(req.GetBlock().GetBlockid()))
	rctx.GetLog().SetInfoField("is_trunk_tip", isTrunkTip)

	return resp, nil
}

// GetBlockChains get BlockChains
func (t *RpcServ) GetBlockChains(gctx context.Context, req *pb.CommonIn) (*pb.BlockChains, error) {
	// 默认响应
	resp := &pb.BlockChains{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}
	resp.Blockchains = t.engine.GetChains()
	return resp, nil
}

// GetSystemStatus get systemstatus
func (t *RpcServ) GetSystemStatus(gctx context.Context, req *pb.CommonIn) (*pb.SystemsStatusReply, error) {
	// 默认响应
	resp := &pb.SystemsStatusReply{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	systemsStatus := &pb.SystemsStatus{
		Speeds: &pb.Speeds{
			SumSpeeds: make(map[string]float64),
			BcSpeeds:  make(map[string]*pb.BCSpeeds),
		},
	}
	bcs := t.engine.GetChains()
	for _, bcName := range bcs {
		bcStatus := &pb.BCStatus{Header: req.Header, Bcname: bcName}
		status, err := t.GetBlockChainStatus(gctx, bcStatus)
		if err != nil {
			rctx.GetLog().Warn("get chain status error", "error", err)
		}

		systemsStatus.BcsStatus = append(systemsStatus.BcsStatus, status)
	}

	if req.ViewOption == pb.ViewOption_NONE || req.ViewOption == pb.ViewOption_PEERS {
		peerInfo := t.engine.Context().Net.PeerInfo()
		peerUrls := acom.PeerInfoToStrings(peerInfo)
		systemsStatus.PeerUrls = peerUrls
	}

	resp.SystemsStatus = systemsStatus
	return resp, nil
}

// GetNetURL get net url in p2p_base
func (t *RpcServ) GetNetURL(gctx context.Context, req *pb.CommonIn) (*pb.RawUrl, error) {
	// 默认响应
	resp := &pb.RawUrl{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}
	peerInfo := t.engine.Context().Net.PeerInfo()
	resp.RawUrl = peerInfo.Address

	rctx.GetLog().SetInfoField("raw_url", resp.RawUrl)
	return resp, nil
}

// GetBlockByHeight  get trunk block by height
func (t *RpcServ) GetBlockByHeight(gctx context.Context, req *pb.BlockHeight) (*pb.Block, error) {
	// 默认响应
	resp := &pb.Block{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetBcname() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	handle, err := models.NewChainHandle(req.GetBcname(), rctx)
	if err != nil {
		rctx.GetLog().Warn("new chain handle failed", "err", err.Error())
		return resp, err
	}
	blockInfo, err := handle.QueryBlockByHeight(req.GetHeight(), true)
	if err != nil {
		rctx.GetLog().Warn("query block error", "bc", req.GetBcname(), "height", req.GetHeight())
		return resp, err
	}

	block := acom.BlockToXchain(blockInfo.Block)
	if block == nil {
		rctx.GetLog().Warn("convert block failed")
		return resp, ecom.ErrInternal
	}
	resp.Block = block
	resp.Status = pb.Block_EBlockStatus(blockInfo.Status)
	resp.Bcname = req.GetBcname()
	resp.Blockid = blockInfo.Block.Blockid

	rctx.GetLog().SetInfoField("height", req.GetHeight())
	rctx.GetLog().SetInfoField("blockid", utils.F(blockInfo.Block.Blockid))
	return resp, nil
}

// GetAccountByAK get account list with contain ak
func (t *RpcServ) GetAccountByAK(gctx context.Context, req *pb.AK2AccountRequest) (*pb.AK2AccountResponse, error) {
	// 默认响应
	resp := &pb.AK2AccountResponse{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetBcname() == "" || req.GetAddress() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	handle, err := models.NewChainHandle(req.GetBcname(), rctx)
	if err != nil {
		rctx.GetLog().Warn("new chain handle failed", "err", err.Error())
		return resp, err
	}

	accounts, err := handle.GetAccountByAK(req.GetAddress())
	if err != nil || accounts == nil {
		rctx.GetLog().Warn("QueryAccountContainAK error", "error", err)
		return resp, err
	}

	resp.Account = accounts
	resp.Bcname = req.GetBcname()

	rctx.GetLog().SetInfoField("address", req.GetAddress())
	return resp, err
}

// GetAddressContracts get contracts of accounts contain a specific address
func (t *RpcServ) GetAddressContracts(gctx context.Context, req *pb.AddressContractsRequest) (*pb.AddressContractsResponse, error) {
	// 默认响应
	resp := &pb.AddressContractsResponse{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetBcname() == "" || req.GetAddress() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	handle, err := models.NewChainHandle(req.GetBcname(), rctx)
	if err != nil {
		rctx.GetLog().Warn("new chain handle failed", "err", err.Error())
		return resp, err
	}

	accounts, err := handle.GetAccountByAK(req.GetAddress())
	if err != nil || accounts == nil {
		rctx.GetLog().Warn("GetAccountByAK error", "error", err)
		return resp, err
	}

	// get contracts for each account
	resp.Contracts = make(map[string]*pb.ContractList)
	for _, account := range accounts {
		contracts, err := handle.GetAccountContracts(account)
		if err != nil {
			rctx.GetLog().Warn("GetAddressContracts partial account error", "logid", req.Header.Logid, "error", err)
			continue
		}

		if len(contracts) > 0 {
			xchainContracts, err := acom.ContractStatusListToXchain(contracts)
			if err != nil || xchainContracts == nil {
				rctx.GetLog().Warn("convert contracts failed")
				continue
			}
			resp.Contracts[account] = &pb.ContractList{
				ContractStatus: xchainContracts,
			}
		}
	}

	rctx.GetLog().SetInfoField("address", req.GetAddress())
	return resp, nil
}

func (t *RpcServ) GetConsensusStatus(gctx context.Context, req *pb.ConsensusStatRequest) (*pb.ConsensusStatus, error) {
	// 默认响应
	resp := &pb.ConsensusStatus{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetBcname() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}
	handle, err := models.NewChainHandle(req.GetBcname(), rctx)
	if err != nil {
		rctx.GetLog().Warn("new chain handle failed", "err", err.Error())
		return resp, ecom.ErrForbidden
	}

	status, err := handle.QueryConsensusStatus()
	if err != nil {
		rctx.GetLog().Warn("get chain status error", "err", err)
		return resp, ecom.ErrForbidden
	}
	resp.ConsensusName = status.ConsensusName
	resp.Version = status.Version
	resp.StartHeight = status.StartHeight
	resp.ValidatorsInfo = status.ValidatorsInfo
	return resp, nil
}

// DposCandidates get all candidates of the tdpos consensus
func (t *RpcServ) DposCandidates(gctx context.Context, req *pb.DposCandidatesRequest) (*pb.DposCandidatesResponse, error) {
	// 默认响应
	resp := &pb.DposCandidatesResponse{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetBcname() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	return resp, ecom.ErrForbidden
}

// DposNominateRecords get all records nominated by an user
func (t *RpcServ) DposNominateRecords(gctx context.Context, req *pb.DposNominateRecordsRequest) (*pb.DposNominateRecordsResponse, error) {
	// 默认响应
	resp := &pb.DposNominateRecordsResponse{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetBcname() == "" || req.GetAddress() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	return resp, ecom.ErrForbidden
}

// DposNomineeRecords get nominated record of a candidate
func (t *RpcServ) DposNomineeRecords(gctx context.Context, req *pb.DposNomineeRecordsRequest) (*pb.DposNomineeRecordsResponse, error) {
	// 默认响应
	resp := &pb.DposNomineeRecordsResponse{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetBcname() == "" || req.GetAddress() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	return resp, ecom.ErrForbidden
}

// DposVoteRecords get all vote records voted by an user
func (t *RpcServ) DposVoteRecords(gctx context.Context, req *pb.DposVoteRecordsRequest) (*pb.DposVoteRecordsResponse, error) {
	// 默认响应
	resp := &pb.DposVoteRecordsResponse{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetBcname() == "" || req.GetAddress() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	return resp, ecom.ErrForbidden
}

// DposVotedRecords get all vote records of a candidate
func (t *RpcServ) DposVotedRecords(gctx context.Context, req *pb.DposVotedRecordsRequest) (*pb.DposVotedRecordsResponse, error) {
	// 默认响应
	resp := &pb.DposVotedRecordsResponse{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetBcname() == "" || req.GetAddress() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	return resp, ecom.ErrForbidden
}

// DposCheckResults get check results of a specific term
func (t *RpcServ) DposCheckResults(gctx context.Context, req *pb.DposCheckResultsRequest) (*pb.DposCheckResultsResponse, error) {
	// 默认响应
	resp := &pb.DposCheckResultsResponse{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetBcname() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	return resp, ecom.ErrForbidden
}

// DposStatus get dpos status
func (t *RpcServ) DposStatus(gctx context.Context, req *pb.DposStatusRequest) (*pb.DposStatusResponse, error) {
	// 默认响应
	resp := &pb.DposStatusResponse{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetBcname() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	return resp, ecom.ErrForbidden
}
