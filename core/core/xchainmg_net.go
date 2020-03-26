package xchaincore

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/golang/protobuf/proto"

	"github.com/xuperchain/xuperchain/core/common/config"
	"github.com/xuperchain/xuperchain/core/global"
	p2p_base "github.com/xuperchain/xuperchain/core/p2p/base"
	xuper_p2p "github.com/xuperchain/xuperchain/core/p2p/pb"
	"github.com/xuperchain/xuperchain/core/pb"
)

// RegisterSubscriber register p2p_base msg type
func (xm *XChainMG) RegisterSubscriber() error {
	xm.Log.Trace("Start to Register Subscriber")
	if _, err := xm.P2pSvr.Register(xm.P2pSvr.NewSubscriber(xm.msgChan, xuper_p2p.XuperMessage_POSTTX, nil, "", xm.Log)); err != nil {
		return err
	}

	if _, err := xm.P2pSvr.Register(xm.P2pSvr.NewSubscriber(xm.msgChan, xuper_p2p.XuperMessage_SENDBLOCK, nil, "", xm.Log)); err != nil {
		return err
	}

	if _, err := xm.P2pSvr.Register(xm.P2pSvr.NewSubscriber(xm.msgChan, xuper_p2p.XuperMessage_BATCHPOSTTX, nil, "", xm.Log)); err != nil {
		return err
	}

	if _, err := xm.P2pSvr.Register(xm.P2pSvr.NewSubscriber(xm.msgChan, xuper_p2p.XuperMessage_NEW_BLOCKID, nil, "", xm.Log)); err != nil {
		return err
	}

	if _, err := xm.P2pSvr.Register(xm.P2pSvr.NewSubscriber(nil, xuper_p2p.XuperMessage_GET_BLOCK, xm.handleGetBlock, "", xm.Log)); err != nil {
		return err
	}

	if _, err := xm.P2pSvr.Register(xm.P2pSvr.NewSubscriber(nil, xuper_p2p.XuperMessage_GET_BLOCKCHAINSTATUS, xm.handleGetBlockChainStatus, "", xm.Log)); err != nil {
		return err
	}

	if _, err := xm.P2pSvr.Register(xm.P2pSvr.NewSubscriber(nil, xuper_p2p.XuperMessage_CONFIRM_BLOCKCHAINSTATUS, xm.handleConfirmBlockChainStatus, "", xm.Log)); err != nil {
		return err
	}

	if _, err := xm.P2pSvr.Register(xm.P2pSvr.NewSubscriber(nil, xuper_p2p.XuperMessage_GET_RPC_PORT, xm.handleGetRPCPort, "", xm.Log)); err != nil {
		return err
	}

	xm.Log.Trace("Stop to Register Subscriber")
	return nil
}

// StartLoop dispatch msg received
func (xm *XChainMG) StartLoop() {
	xm.Log.Info("XchainMg start loop to process net msg")
	for {
		select {
		case msg := <-xm.msgChan:
			// handle received msg
			xm.Log.Info("XchainMG get msg", "logid", msg.GetHeader().GetLogid(), "msgType", msg.GetHeader().GetType(), "checksum", msg.GetHeader().GetDataCheckSum())
			go xm.handleReceivedMsg(msg)
		}
	}
}

func (xm *XChainMG) handleReceivedMsg(msg *xuper_p2p.XuperMessage) {
	bcname := msg.GetHeader().GetBcname()
	from := msg.GetHeader().GetFrom()
	if !xm.IsPeerInGroupChain(bcname, from) {
		xm.Log.Warn("remote node ip is not in white list, refuse it")
		return
	}
	// check msg type
	msgType := msg.GetHeader().GetType()
	if msgType != xuper_p2p.XuperMessage_POSTTX && msgType != xuper_p2p.XuperMessage_SENDBLOCK && msgType !=
		xuper_p2p.XuperMessage_BATCHPOSTTX && msgType != xuper_p2p.XuperMessage_NEW_BLOCKID {
		xm.Log.Warn("Received msg cannot handled!", "logid", msg.GetHeader().GetLogid())
		return
	}

	// verify msg
	if !p2p_base.VerifyDataCheckSum(msg) {
		xm.Log.Warn("Verify Data error!", "logid", msg.GetHeader().GetLogid())
		return
	}

	// process msg
	switch msgType {
	case xuper_p2p.XuperMessage_POSTTX:
		xm.handlePostTx(msg)
	case xuper_p2p.XuperMessage_SENDBLOCK:
		xm.HandleSendBlock(msg)
	case xuper_p2p.XuperMessage_BATCHPOSTTX:
		xm.handleBatchPostTx(msg)
	case xuper_p2p.XuperMessage_NEW_BLOCKID:
		xm.handleNewBlockID(msg)
	}
}

func (xm *XChainMG) handlePostTx(msg *xuper_p2p.XuperMessage) {
	txStatus := &pb.TxStatus{}

	txStatusBuf, err := p2p_base.Uncompress(msg)
	if txStatusBuf == nil || err != nil {
		xm.Log.Error("handlePostTx xuper_p2p uncompressed error", "error", err)
		return
	}
	// Unmarshal msg
	err = proto.Unmarshal(txStatusBuf, txStatus)
	if err != nil {
		xm.Log.Error("handlePostTx Unmarshal msg to tx error", "logid", msg.GetHeader().GetLogid())
		return
	}

	// process tx
	if txStatus.Header == nil {
		txStatus.Header = global.GHeader()
	}
	if _, needRepost, _ := xm.ProcessTx(txStatus); needRepost {
		opts := []p2p_base.MessageOption{
			p2p_base.WithFilters([]p2p_base.FilterStrategy{p2p_base.DefaultStrategy}),
			p2p_base.WithBcName(msg.GetHeader().GetBcname()),
		}
		go xm.P2pSvr.SendMessage(context.Background(), msg, opts...)
	}
	return
}

// ProcessTx process tx, move from server/server.go
func (xm *XChainMG) ProcessTx(in *pb.TxStatus) (*pb.CommonReply, bool, error) {
	out := &pb.CommonReply{Header: &pb.Header{Logid: in.Header.Logid}}

	if err := validatePostTx(in); err != nil {
		out.Header.Error = pb.XChainErrorEnum_VALIDATE_ERROR
		xm.Log.Trace("PostTx validate param errror", "logid", in.Header.Logid, "error", err.Error())
		return out, false, err
	}

	if len(in.Tx.TxInputs) == 0 && !xm.Cfg.Utxo.NonUtxo {
		out.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE // 拒绝
		xm.Log.Warn("PostTx TxInputs can not be null while need utxo!", "logid", in.Header.Logid)
		return out, false, nil
	}

	bc := xm.Get(in.Bcname)
	if bc == nil {
		out.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE // 拒绝
		return out, false, nil
	}

	hd := &global.XContext{Timer: global.NewXTimer()}

	if bc.GetNodeMode() == config.NodeModeFastSync {
		out.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE // 拒绝
		xm.Log.Warn("PostTx NodeMode is FAST_SYNC, refused!")
		return out, false, nil
	}
	out, needRepost := bc.PostTx(in, hd)

	if !needRepost {
		go produceTransactionEvent(xm.EventService, in.GetTx(), in.GetBcname(), pb.TransactionStatus_FAILED)
	}
	return out, needRepost, nil
}

// HandleSendBlock handle SENDBLOCK type msg
func (xm *XChainMG) HandleSendBlock(msg *xuper_p2p.XuperMessage) {
	block := &pb.Block{}
	blockBuf, err := p2p_base.Uncompress(msg)
	if blockBuf == nil || err != nil {
		xm.Log.Error("HandleSendBlock xuper_p2p uncompressed error", "error", err)
		return
	}
	xm.Log.Trace("Start to HandleSendBlock", "logid", msg.GetHeader().GetLogid(), "checksum", msg.GetHeader().GetDataCheckSum())
	// Unmarshal msg
	err = proto.Unmarshal(blockBuf, block)
	if err != nil {
		xm.Log.Error("HandleSendBlock Unmarshal msg to block error", "logid", msg.GetHeader().GetLogid())
		return
	}
	// process block
	if block.Header == nil {
		block.Header = global.GHeader()
	}
	xm.Log.Trace("Start to HandleSendBlock", "block.header.logid", block.GetHeader().GetLogid())
	if err := xm.ProcessBlock(block); err != nil {
		if err == ErrBlockExist {
			xm.Log.Debug("ProcessBlock SendBlock block exists")
			return
		}
		xm.Log.Error("HandleSendBlock ProcessBlock error", "error", err.Error())
		return
	}

	// send block to peers
	bcname := block.GetBcname()
	bc := xm.Get(bcname)
	if xm.Cfg.BlockBroadcaseMode == 0 {
		// send full block to peers in Full_BroadCast_Mode
		filters := []p2p_base.FilterStrategy{p2p_base.DefaultStrategy}
		if bc.NeedCoreConnection() {
			filters = append(filters, p2p_base.CorePeersStrategy)
		}
		opts := []p2p_base.MessageOption{
			p2p_base.WithFilters(filters),
			p2p_base.WithBcName(bcname),
		}
		go xm.P2pSvr.SendMessage(context.Background(), msg, opts...)
	} else {
		// send block id in Interactive_BroadCast_Mode or Mixed_BroadCast_Mode
		// we could use Interactive_BroadCast_Mode to avoid duplicate messages
		blockidMsg := &pb.Block{
			Bcname:  bcname,
			Blockid: block.Blockid,
		}
		msgInfo, err := proto.Marshal(blockidMsg)
		if err != nil {
			xm.Log.Error("HandleSendBlock marshal NEW_BLOCKID message failed", "logid", msg.GetHeader().GetLogid())
			return
		}
		msg, _ := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion1, bcname, "", xuper_p2p.XuperMessage_NEW_BLOCKID, msgInfo, xuper_p2p.XuperMessage_NONE)
		filters := []p2p_base.FilterStrategy{p2p_base.DefaultStrategy}
		if bc.NeedCoreConnection() {
			filters = append(filters, p2p_base.CorePeersStrategy)
		}
		opts := []p2p_base.MessageOption{
			p2p_base.WithFilters(filters),
			p2p_base.WithBcName(bcname),
		}
		go xm.P2pSvr.SendMessage(context.Background(), msg, opts...)
	}
	return
}

// ProcessBlock process block
func (xm *XChainMG) ProcessBlock(block *pb.Block) error {
	if err := validateSendBlock(block); err != nil {
		xm.Log.Error("ProcessBlock validateSendBlock error", "error", err.Error())
		return err
	}

	xm.Log.Trace("Start to dealwith SendBlock", "blockid", global.F(block.GetBlockid()))
	bc := xm.Get(block.GetBcname())
	if bc == nil {
		xm.Log.Error("ProcessBlock error", "error", "bc not exist")
		return ErrBlockChainNotExist
	}
	hd := &global.XContext{Timer: global.NewXTimer()}
	if err := bc.ProcessSendBlock(block, hd); err != nil {
		if err == ErrBlockExist {
			xm.Log.Debug("ProcessBlock SendBlock block exists")
			return err
		}
		xm.Log.Error("ProcessBlock SendBlock error", "err", err)
		return err
	}
	meta := bc.Ledger.GetMeta()
	xm.Log.Info("SendBlock", "cost", hd.Timer.Print(), "genesis", fmt.Sprintf("%x", meta.RootBlockid),
		"last", fmt.Sprintf("%x", meta.TipBlockid),
		"height", meta.TrunkHeight, "utxo", global.F(bc.Utxovm.GetLatestBlockid()))
	return nil
}

func (xm *XChainMG) handleBatchPostTx(msg *xuper_p2p.XuperMessage) {
	batchTxs := &pb.BatchTxs{}
	batchTxsBuf, err := p2p_base.Uncompress(msg)
	if batchTxsBuf == nil || err != nil {
		xm.Log.Error("handleBatchPostTx xuper_p2p uncompressed error", "error", err)
		return
	}
	// Unmarshal msg
	err = proto.Unmarshal(batchTxsBuf, batchTxs)
	if err != nil {
		xm.Log.Error("handleBatchPostTx Unmarshal msg to BatchTxs error", "logid", msg.GetHeader().GetLogid())
		return
	}

	// process batch post tx
	txs, err := xm.ProcessBatchTx(batchTxs)
	if err != nil {
		xm.Log.Error("HandleSendBlock ProcessBlock error", "error", err.Error())
		return
	}
	if len(txs.Txs) != 0 {
		txsData, err := proto.Marshal(txs)
		if err != nil {
			xm.Log.Error("handleBatchPostTx Marshal txs error", "error", err)
			return
		}
		header := msg.GetHeader()
		msg, _ := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion1, header.GetBcname(), header.GetLogid(), xuper_p2p.XuperMessage_BATCHPOSTTX, txsData, xuper_p2p.XuperMessage_SUCCESS)
		opts := []p2p_base.MessageOption{
			p2p_base.WithFilters([]p2p_base.FilterStrategy{p2p_base.DefaultStrategy}),
			p2p_base.WithBcName(msg.GetHeader().GetBcname()),
			p2p_base.WithCompress(xm.enableCompress),
		}
		go xm.P2pSvr.SendMessage(context.Background(), msg, opts...)
	}
	return
}

// ProcessBatchTx process batch tx
func (xm *XChainMG) ProcessBatchTx(batchTx *pb.BatchTxs) (*pb.BatchTxs, error) {
	succTxs := []*pb.TxStatus{}
	for _, v := range batchTx.Txs {
		_, needRepost, _ := xm.ProcessTx(v)
		if needRepost {
			succTxs = append(succTxs, v)
		} else {
			break // no need to continue
		}
	}
	batchTx.Txs = succTxs
	return batchTx, nil
}

// 处理getBlock消息回调函数
func (xm *XChainMG) handleGetBlock(ctx context.Context, msg *xuper_p2p.XuperMessage) (*xuper_p2p.XuperMessage, error) {
	bcname := msg.GetHeader().GetBcname()
	logid := msg.GetHeader().GetLogid()
	from := msg.GetHeader().GetFrom()
	if !xm.IsPeerInGroupChain(bcname, from) {
		xm.Log.Warn("remote node ip is not in white list, refuse it")
		res, _ := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, bcname, logid,
			xuper_p2p.XuperMessage_GET_BLOCK_RES, []byte("unknown"), xuper_p2p.XuperMessage_CHECK_SUM_ERROR)
		return res, errors.New("remote node ip is not in white list, refuse it")
	}
	xm.Log.Trace("Start to handleGetBlock", "bcname", bcname, "logid", logid)
	block := &pb.Block{Header: global.GHeader()}
	if !p2p_base.VerifyDataCheckSum(msg) {
		xm.Log.Warn("handleGetBlock verify msg error", "log_id", logid)
		resBuf, _ := proto.Marshal(block)
		res, _ := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, bcname, logid,
			xuper_p2p.XuperMessage_GET_BLOCK_RES, resBuf, xuper_p2p.XuperMessage_CHECK_SUM_ERROR)
		return res, errors.New("verify msg error")
	}
	bid := &pb.BlockID{}
	err := proto.Unmarshal(msg.GetData().GetMsgInfo(), bid)

	if err != nil {
		xm.Log.Error("handleGetBlock unmarshal msg error", "error", err.Error())
		resBuf, _ := proto.Marshal(block)
		res, _ := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, bcname, logid,
			xuper_p2p.XuperMessage_GET_BLOCK_RES, resBuf, xuper_p2p.XuperMessage_UNMARSHAL_MSG_BODY_ERROR)
		return res, errors.New("unmarshal msg error")
	}

	bc := xm.Get(bcname)
	if bc == nil {
		xm.Log.Error("handleGetBlock Get blockchain error", "error", "blockchain not exit", "bcname", bcname)
		resBuf, _ := proto.Marshal(block)
		res, _ := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, bcname, logid,
			xuper_p2p.XuperMessage_GET_BLOCK_RES, resBuf, xuper_p2p.XuperMessage_BLOCKCHAIN_NOTEXIST)
		return res, errors.New("blockChain not exit")
	}
	block = bc.GetBlock(bid)
	xm.Log.Trace("Start to dealwith GetBlock result", "logid", logid,
		"blockid", block.GetBlock().GetBlockid(), "height", block.GetBlock().GetHeight())
	if block.GetHeader().GetError() != pb.XChainErrorEnum_SUCCESS {
		xm.Log.Error("handleGetBlock GetBlock error", "error", block.GetHeader().GetError())
		resBuf, _ := proto.Marshal(block)
		res, _ := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, bcname, logid,
			xuper_p2p.XuperMessage_GET_BLOCK_RES, resBuf, xuper_p2p.XuperMessage_GET_BLOCK_ERROR)
		return res, errors.New("getBlock error")
	}

	resBuf, _ := proto.Marshal(block)
	res, err := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, bcname, logid,
		xuper_p2p.XuperMessage_GET_BLOCK_RES, resBuf, xuper_p2p.XuperMessage_SUCCESS)
	if xm.enableCompress {
		res = p2p_base.Compress(res)
	}
	return res, err
}

// 处理getBlockChainStatus消息回调函数
func (xm *XChainMG) handleGetBlockChainStatus(ctx context.Context, msg *xuper_p2p.XuperMessage) (*xuper_p2p.XuperMessage, error) {
	bcname := msg.GetHeader().GetBcname()
	logid := msg.GetHeader().GetLogid()
	from := msg.GetHeader().GetFrom()
	if !xm.IsPeerInGroupChain(bcname, from) {
		xm.Log.Warn("remote node ip is not in white list, refuse it")
		res, _ := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, bcname, logid,
			xuper_p2p.XuperMessage_GET_BLOCK_RES, []byte("unknown"), xuper_p2p.XuperMessage_CHECK_SUM_ERROR)
		return res, errors.New("remote node ip is not in white list, refuse it")
	}
	xm.Log.Trace("Start to handleGetBlockChainStatus", "bcname", bcname, "logid", logid)
	if !p2p_base.VerifyDataCheckSum(msg) {
		xm.Log.Warn("handleGetBlockChainStatus verify msg error", "log_id", logid)
		res, _ := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, bcname, logid,
			xuper_p2p.XuperMessage_GET_BLOCKCHAINSTATUS_RES, nil, xuper_p2p.XuperMessage_CHECK_SUM_ERROR)
		return res, errors.New("verify msg error")
	}
	bcStatus := &pb.BCStatus{}
	err := proto.Unmarshal(msg.GetData().GetMsgInfo(), bcStatus)
	if err != nil {
		xm.Log.Error("handleGetBlockChainStatus unmarshal msg error", "error", err.Error())
		res, _ := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, bcname, logid,
			xuper_p2p.XuperMessage_GET_BLOCKCHAINSTATUS_RES, nil, xuper_p2p.XuperMessage_UNMARSHAL_MSG_BODY_ERROR)
		return res, errors.New("unmarshal msg error")
	}
	bc := xm.Get(bcname)
	if bc == nil {
		xm.Log.Error("handleGetBlockChainStatus Get blockchain error", "error", "blockchain not exit", "bcname", bcname)
		res, _ := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, bcname, logid,
			xuper_p2p.XuperMessage_GET_BLOCKCHAINSTATUS_RES, nil, xuper_p2p.XuperMessage_BLOCKCHAIN_NOTEXIST)
		return res, errors.New("blockChain not exit")
	}
	bcStatusRes := bc.GetBlockChainStatus(bcStatus)
	// no need to transfer branch id msg
	bcStatusRes.BranchBlockid = nil
	if bcStatusRes.GetHeader().GetError() != pb.XChainErrorEnum_SUCCESS {
		xm.Log.Error("handleGetBlockChainStatus Get blockchain error", "error", bcStatusRes.GetHeader().GetError())
		res, _ := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, bcname, logid,
			xuper_p2p.XuperMessage_GET_BLOCKCHAINSTATUS_RES, nil, xuper_p2p.XuperMessage_GET_BLOCKCHAIN_ERROR)
		return res, errors.New("get BlockChainStatus error")
	}
	resBuf, _ := proto.Marshal(bcStatusRes)
	res, err := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, bcname, logid,
		xuper_p2p.XuperMessage_GET_BLOCKCHAINSTATUS_RES, resBuf, xuper_p2p.XuperMessage_SUCCESS)
	return res, err
}

// 处理confirm blockChain status 回调函数
func (xm *XChainMG) handleConfirmBlockChainStatus(ctx context.Context, msg *xuper_p2p.XuperMessage) (*xuper_p2p.XuperMessage, error) {
	bcname := msg.GetHeader().GetBcname()
	logid := msg.GetHeader().GetLogid()
	from := msg.GetHeader().GetFrom()
	if !xm.IsPeerInGroupChain(bcname, from) {
		xm.Log.Warn("remote node ip is not in white list, refuse it")
		res, _ := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, bcname, logid,
			xuper_p2p.XuperMessage_GET_BLOCK_RES, []byte("unknown"), xuper_p2p.XuperMessage_CHECK_SUM_ERROR)
		return res, errors.New("remote node ip is not in white list, refuse it")
	}
	xm.Log.Trace("Start to handleConfirmBlockChainStatus", "bcname", bcname, "logid", logid)
	if !p2p_base.VerifyDataCheckSum(msg) {
		xm.Log.Warn("handleConfirmBlockChainStatus verify msg error", "log_id", logid)
		res, _ := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, bcname, logid,
			xuper_p2p.XuperMessage_CONFIRM_BLOCKCHAINSTATUS_RES, nil, xuper_p2p.XuperMessage_CHECK_SUM_ERROR)
		return res, errors.New("verify msg error")
	}

	bcStatus := &pb.BCStatus{}
	err := proto.Unmarshal(msg.GetData().GetMsgInfo(), bcStatus)
	if err != nil {
		xm.Log.Error("handleConfirmBlockChainStatus unmarshal msg error", "error", err.Error())
		res, _ := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, bcname, logid,
			xuper_p2p.XuperMessage_CONFIRM_BLOCKCHAINSTATUS_RES, nil, xuper_p2p.XuperMessage_UNMARSHAL_MSG_BODY_ERROR)
		return res, errors.New("unmarshal msg error")
	}

	bc := xm.Get(bcname)
	if bc == nil {
		xm.Log.Error("handleConfirmBlockChainStatus Get blockchain error", "error", "blockchain not exit", "bcname", bcname)
		res, _ := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, bcname, logid,
			xuper_p2p.XuperMessage_CONFIRM_BLOCKCHAINSTATUS_RES, nil, xuper_p2p.XuperMessage_BLOCKCHAIN_NOTEXIST)
		return res, errors.New("blockChain not exit")
	}
	tipStatus := bc.ConfirmTipBlockChainStatus(bcStatus)
	if tipStatus.GetHeader().GetError() != pb.XChainErrorEnum_SUCCESS {
		xm.Log.Error("handleConfirmBlockChainStatus ConfirmTipBlockChainStatus error", "error", tipStatus.GetHeader().GetError())
		res, _ := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, bcname, logid,
			xuper_p2p.XuperMessage_CONFIRM_BLOCKCHAINSTATUS_RES, nil, xuper_p2p.XuperMessage_CONFIRM_BLOCKCHAINSTATUS_ERROR)
		return res, errors.New("confirmBlockChainStatus error")
	}
	resBuf, _ := proto.Marshal(tipStatus)
	res, err := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, bcname, logid,
		xuper_p2p.XuperMessage_CONFIRM_BLOCKCHAINSTATUS_RES, resBuf, xuper_p2p.XuperMessage_SUCCESS)
	return res, err
}

// 处理获取RPC端口回调函数
func (xm *XChainMG) handleGetRPCPort(ctx context.Context, msg *xuper_p2p.XuperMessage) (*xuper_p2p.XuperMessage, error) {
	bcname := msg.GetHeader().GetBcname()
	from := msg.GetHeader().GetFrom()
	if !xm.IsPeerInGroupChain(bcname, from) {
		xm.Log.Warn("remote node ip is not in white list, refuse it")
		res, _ := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, msg.GetHeader().GetBcname(), msg.GetHeader().GetLogid(),
			xuper_p2p.XuperMessage_GET_BLOCK_RES, []byte("unknown"), xuper_p2p.XuperMessage_CHECK_SUM_ERROR)
		return res, errors.New("remote node ip is not in white list, refuse it")
	}
	xm.Log.Trace("Start to handleGetRPCPort", "logid", msg.GetHeader().GetLogid())
	_, port, err := net.SplitHostPort(xm.Cfg.TCPServer.Port)
	if err != nil {
		xm.Log.Error("handleGetRPCPort SplitHostPort error", "error", err.Error())
		return nil, err
	}
	return p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, "", msg.GetHeader().GetLogid(), xuper_p2p.XuperMessage_GET_RPC_PORT_RES, []byte(":"+port), xuper_p2p.XuperMessage_NONE)
}

// handleNewBlockID handle NEW_BLOCKID message
func (xm *XChainMG) handleNewBlockID(msg *xuper_p2p.XuperMessage) {
	xm.Log.Trace("Start to handleNewBlockID", "logid", msg.GetHeader().GetLogid())
	block := &pb.Block{}
	blockBuf, err := p2p_base.Uncompress(msg)
	if err != nil || blockBuf == nil {
		xm.Log.Error("handleNewBlockID uncompressed error", "error", err, "logid", msg.GetHeader().GetLogid())
		return
	}
	err = proto.Unmarshal(blockBuf, block)
	if err != nil {
		xm.Log.Warn("handleNewBlockID received unknown message", "error", err, "logid", msg.GetHeader().GetLogid())
		return
	}

	// handle get blockidin xchaincore
	bcname := block.GetBcname()
	bc := xm.Get(bcname)
	if bc == nil {
		xm.Log.Warn("handleNewBlockID get bc is nil", "logid", msg.GetHeader().GetLogid())
		return
	}
	ctx := context.Background()
	blockRes, err := bc.handleNewBlockID(ctx, block.GetBlockid(), msg.GetHeader().GetFrom())
	if err != nil {
		xm.Log.Warn("handleNewBlockID process message failed", "error", err, "logid", msg.GetHeader().GetLogid())
		return
	}

	if blockRes == nil {
		xm.Log.Trace("handleNewBlockID may received this block id before", "blockid", block.GetBlockid(),
			"logid", msg.GetHeader().GetLogid())
		return
	}

	// process block
	if blockRes.Header == nil {
		blockRes.Header = global.GHeader()
	}
	if err := xm.ProcessBlock(blockRes); err != nil {
		if err == ErrBlockExist {
			xm.Log.Debug("handleNewBlockID: ProcessBlock block exists")
			return
		}
		xm.Log.Error("handleNewBlockID ProcessBlock error", "error", err.Error())
		return
	}

	// broadcast New_BlockID message
	// since the origin message is New_BlockID, so ignore Full_BroadCast_Mode
	filters := []p2p_base.FilterStrategy{p2p_base.DefaultStrategy}
	if bc.NeedCoreConnection() {
		filters = append(filters, p2p_base.CorePeersStrategy)
	}
	opts := []p2p_base.MessageOption{
		p2p_base.WithFilters(filters),
		p2p_base.WithBcName(bcname),
	}
	go xm.P2pSvr.SendMessage(context.Background(), msg, opts...)
}
