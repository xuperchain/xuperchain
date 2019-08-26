/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package server

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	_ "net/http/pprof"
	"strconv"

	"github.com/golang/protobuf/proto"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/xuperchain/log15"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	"github.com/xuperchain/xuperunion/common"
	"github.com/xuperchain/xuperunion/consensus"
	xchaincore "github.com/xuperchain/xuperunion/core"
	"github.com/xuperchain/xuperunion/crypto/hash"
	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/p2pv2"
	xuper_p2p "github.com/xuperchain/xuperunion/p2pv2/pb"
	"github.com/xuperchain/xuperunion/pb"
)

type server struct {
	log            log.Logger
	mg             *xchaincore.XChainMG
	dedupCache     *common.LRUCache
	dedupTimeLimit int
}

// PostTx Update db
func (s *server) PostTx(ctx context.Context, in *pb.TxStatus) (*pb.CommonReply, error) {
	if in.Header == nil {
		in.Header = global.GHeader()
	}

	out, needRepost, err := s.mg.ProcessTx(in)
	if needRepost {
		msgInfo, _ := proto.Marshal(in)
		msg, _ := xuper_p2p.NewXuperMessage(xuper_p2p.XuperMsgVersion1, in.GetBcname(), in.GetHeader().GetLogid(), xuper_p2p.XuperMessage_POSTTX, msgInfo, xuper_p2p.XuperMessage_NONE)
		opts := []p2pv2.MessageOption{
			p2pv2.WithFilters([]p2pv2.FilterStrategy{p2pv2.DefaultStrategy}),
			p2pv2.WithBcName(in.GetBcname()),
		}
		go s.mg.P2pv2.SendMessage(context.Background(), msg, opts...)
	}
	return out, err
}

// BatchPostTx batch update db
func (s *server) BatchPostTx(ctx context.Context, in *pb.BatchTxs) (*pb.CommonReply, error) {
	if in.Header == nil {
		in.Header = global.GHeader()
	}
	out := &pb.CommonReply{Header: &pb.Header{Logid: in.Header.Logid}}
	succTxs := []*pb.TxStatus{}
	for _, v := range in.Txs {
		oneOut, needRepost, _ := s.mg.ProcessTx(v)
		if oneOut.Header.Error != pb.XChainErrorEnum_SUCCESS {
			if oneOut.Header.Error != pb.XChainErrorEnum_UTXOVM_ALREADY_UNCONFIRM_ERROR {
				s.log.Warn("BatchPostTx processTx error", "logid", in.Header.Logid, "error", oneOut.Header.Error, "txid", global.F(v.Txid))
			}
		} else if needRepost {
			succTxs = append(succTxs, v)
		}
	}
	in.Txs = succTxs //只广播成功的
	if len(in.Txs) > 0 {
		txsData, err := proto.Marshal(in)
		if err != nil {
			s.log.Error("handleBatchPostTx Marshal txs error", "error", err)
			return out, nil
		}

		msg, _ := xuper_p2p.NewXuperMessage(xuper_p2p.XuperMsgVersion1, "", in.GetHeader().GetLogid(), xuper_p2p.XuperMessage_BATCHPOSTTX, txsData, xuper_p2p.XuperMessage_NONE)
		opts := []p2pv2.MessageOption{
			p2pv2.WithFilters([]p2pv2.FilterStrategy{p2pv2.DefaultStrategy}),
			p2pv2.WithBcName(in.Txs[0].GetBcname()),
		}
		go s.mg.P2pv2.SendMessage(context.Background(), msg, opts...)
	}
	return out, nil
}

// QueryAcl query some account info
func (s *server) QueryACL(ctx context.Context, in *pb.AclStatus) (*pb.AclStatus, error) {
	s.mg.Speed.Add("QueryAcl")
	if in.Header == nil {
		in.Header = global.GHeader()
	}
	out := &pb.AclStatus{Header: &pb.Header{}}
	bc := s.mg.Get(in.Bcname)

	if bc == nil {
		out.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE // 拒绝
		s.log.Trace("refused a connection at function call QueryAcl", "logid", in.Header.Logid)
		return out, nil
	}

	accountName := in.GetAccountName()
	contractName := in.GetContractName()
	methodName := in.GetMethodName()
	if len(accountName) > 0 {
		acl, confirmed, err := bc.QueryAccountACL(accountName)
		out.Confirmed = confirmed
		if err != nil {
			return out, err
		}
		out.Acl = acl
		return out, nil
	} else if len(contractName) > 0 {
		if len(methodName) > 0 {
			acl, confirmed, err := bc.QueryContractMethodACL(contractName, methodName)
			out.Confirmed = confirmed
			if err != nil {
				return out, err
			}
			out.Acl = acl
			return out, nil
		}
	}
	return out, nil
}

// GetAccountContractsRequest get account request
func (s *server) GetAccountContracts(ctx context.Context, in *pb.GetAccountContractsRequest) (*pb.GetAccountContractsResponse, error) {
	if in.Header == nil {
		in.Header = global.GHeader()
	}
	out := &pb.GetAccountContractsResponse{Header: &pb.Header{Logid: in.GetHeader().GetLogid()}}
	bc := s.mg.Get(in.GetBcname())
	if bc == nil {
		// bc not found
		out.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE
		s.log.Trace("refused a connection while GetAccountContracts", "logid", in.Header.Logid)
		return out, nil
	}
	contractsStatus, err := bc.GetAccountContractsStatus(in.GetAccount())
	if err != nil {
		out.Header.Error = pb.XChainErrorEnum_ACCOUNT_CONTRACT_STATUS_ERROR
		s.log.Warn("GetAccountContracts error", "logid", in.Header.Logid, "error", err.Error())
		return out, err
	}
	out.ContractsStatus = contractsStatus
	return out, nil
}

// QueryTx Get transaction details
func (s *server) QueryTx(ctx context.Context, in *pb.TxStatus) (*pb.TxStatus, error) {
	s.mg.Speed.Add("QueryTx")
	if in.Header == nil {
		in.Header = global.GHeader()
	}
	out := &pb.TxStatus{Header: &pb.Header{}}
	bc := s.mg.Get(in.Bcname)
	if bc == nil {
		out.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE // 拒绝
		return out, nil
	}
	if bc.QueryTxFromForbidden(in.Txid) {
		return out, errors.New("tx has been forbidden")
	}
	out = bc.QueryTx(in)
	return out, nil
}

// GetBalance get balance for account or addr
func (s *server) GetBalance(ctx context.Context, in *pb.AddressStatus) (*pb.AddressStatus, error) {
	s.mg.Speed.Add("GetBalance")
	if in.Header == nil {
		in.Header = global.GHeader()
	}
	for i := 0; i < len(in.Bcs); i++ {
		bc := s.mg.Get(in.Bcs[i].Bcname)
		if bc == nil {
			in.Bcs[i].Error = pb.XChainErrorEnum_BLOCKCHAIN_NOTEXIST
			in.Bcs[i].Balance = ""
		} else {
			bi, err := bc.GetBalance(in.Address)
			if err != nil {
				in.Bcs[i].Error = HandleBlockCoreError(err)
				in.Bcs[i].Balance = ""
			} else {
				in.Bcs[i].Error = pb.XChainErrorEnum_SUCCESS
				in.Bcs[i].Balance = bi
			}
		}
	}
	return in, nil
}

// GetFrozenBalance get balance frozened for account or addr
func (s *server) GetFrozenBalance(ctx context.Context, in *pb.AddressStatus) (*pb.AddressStatus, error) {
	s.mg.Speed.Add("GetFrozenBalance")
	if in.Header == nil {
		in.Header = global.GHeader()
	}
	for i := 0; i < len(in.Bcs); i++ {
		bc := s.mg.Get(in.Bcs[i].Bcname)
		if bc == nil {
			in.Bcs[i].Error = pb.XChainErrorEnum_BLOCKCHAIN_NOTEXIST
			in.Bcs[i].Balance = ""
		} else {
			bi, err := bc.GetFrozenBalance(in.Address)
			if err != nil {
				in.Bcs[i].Error = HandleBlockCoreError(err)
				in.Bcs[i].Balance = ""
			} else {
				in.Bcs[i].Error = pb.XChainErrorEnum_SUCCESS
				in.Bcs[i].Balance = bi
			}
		}
	}
	return in, nil
}

// GetBlock get block info according to blockID
func (s *server) GetBlock(ctx context.Context, in *pb.BlockID) (*pb.Block, error) {
	s.mg.Speed.Add("GetBlock")
	if in.Header == nil {
		in.Header = global.GHeader()
	}
	s.log.Trace("Start to dealwith GetBlock", "logid", in.Header.Logid, "in", in)
	bc := s.mg.Get(in.Bcname)
	if bc == nil {
		out := pb.Block{Header: &pb.Header{}}
		out.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE // 拒绝
		return &out, nil
	}
	out := bc.GetBlock(in)

	block := out.GetBlock()
	transactions := block.GetTransactions()
	transactionsFilter := []*pb.Transaction{}
	for _, transaction := range transactions {
		txid := transaction.GetTxid()
		if bc.QueryTxFromForbidden(txid) {
			continue
		}
		transactionsFilter = append(transactionsFilter, transaction)
	}
	if transactions != nil {
		out.Block.Transactions = transactionsFilter
	}
	s.log.Trace("Start to dealwith GetBlock result", "logid", in.Header.Logid,
		"blockid", out.Blockid, "height", out.GetBlock().GetHeight())
	return out, nil
}

// GetBlockChainStatus get systemstatus
func (s *server) GetBlockChainStatus(ctx context.Context, in *pb.BCStatus) (*pb.BCStatus, error) {
	s.mg.Speed.Add("GetBlockChainStatus")
	if in.Header == nil {
		in.Header = global.GHeader()
	}
	bc := s.mg.Get(in.Bcname)
	if bc == nil {
		out := pb.BCStatus{Header: &pb.Header{}}
		out.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE // 拒绝
		return &out, nil
	}
	return bc.GetBlockChainStatus(in), nil
}

// ConfirmBlockChainStatus confirm is_trunk
func (s *server) ConfirmBlockChainStatus(ctx context.Context, in *pb.BCStatus) (*pb.BCTipStatus, error) {
	s.mg.Speed.Add("ConfirmBlockChainStatus")
	if in.Header == nil {
		in.Header = global.GHeader()
	}
	bc := s.mg.Get(in.Bcname)
	if bc == nil {
		out := pb.BCTipStatus{Header: &pb.Header{}}
		out.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE // 拒绝
		return &out, nil
	}
	return bc.ConfirmTipBlockChainStatus(in), nil
}

// GetBlockChains get BlockChains
func (s *server) GetBlockChains(ctx context.Context, in *pb.CommonIn) (*pb.BlockChains, error) {
	s.mg.Speed.Add("GetBlockChains")
	if in.Header == nil {
		in.Header = global.GHeader()
	}
	out := &pb.BlockChains{Header: &pb.Header{}}
	out.Blockchains = s.mg.GetAll()
	return out, nil
}

// GetSystemStatus get systemstatus
func (s *server) GetSystemStatus(ctx context.Context, in *pb.CommonIn) (*pb.SystemsStatusReply, error) {
	s.mg.Speed.Add("GetSystemStatus")
	if in.Header == nil {
		in.Header = global.GHeader()
	}
	out := &pb.SystemsStatusReply{Header: &pb.Header{}}
	systemsStatus := &pb.SystemsStatus{
		Header: in.Header,
		Speeds: &pb.Speeds{
			SumSpeeds: make(map[string]float64),
			BcSpeeds:  make(map[string]*pb.BCSpeeds),
		},
	}
	bcs := s.mg.GetAll()
	for _, v := range bcs {
		bc := s.mg.Get(v)
		tmpBcs := &pb.BCStatus{Header: in.Header, Bcname: v}
		bcst := bc.GetBlockChainStatus(tmpBcs)
		if _, ok := systemsStatus.Speeds.BcSpeeds[v]; !ok {
			systemsStatus.Speeds.BcSpeeds[v] = &pb.BCSpeeds{}
			systemsStatus.Speeds.BcSpeeds[v].BcSpeed = bc.Speed.GetMaxSpeed()
		}
		systemsStatus.BcsStatus = append(systemsStatus.BcsStatus, bcst)
	}
	systemsStatus.Speeds.SumSpeeds = s.mg.Speed.GetMaxSpeed()
	systemsStatus.PeerUrls = s.mg.P2pv2.GetPeerUrls()
	out.SystemsStatus = systemsStatus
	return out, nil
}

// GetNetURL get net url in p2pv2
func (s *server) GetNetURL(ctx context.Context, in *pb.CommonIn) (*pb.RawUrl, error) {
	if in.Header == nil {
		in.Header = global.GHeader()
	}
	out := &pb.RawUrl{Header: &pb.Header{Logid: in.Header.Logid}}
	out.Header.Error = pb.XChainErrorEnum_SUCCESS
	netURL := s.mg.P2pv2.GetNetURL()
	out.RawUrl = netURL
	return out, nil
}

// SelectUTXO select utxo inputs
func (s *server) SelectUTXO(ctx context.Context, in *pb.UtxoInput) (*pb.UtxoOutput, error) {
	if in.Header == nil {
		in.Header = global.GHeader()
	}
	out := &pb.UtxoOutput{Header: &pb.Header{Logid: in.Header.Logid}}
	bc := s.mg.Get(in.GetBcname())
	if bc == nil {
		out.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE // 拒绝
		s.log.Warn("failed to select utxo, bcname not exists", "logid", in.Header.Logid)
		return out, nil
	}

	totalNeed, ok := new(big.Int).SetString(in.TotalNeed, 10)
	if !ok {
		out.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE // 拒绝
		return out, nil
	}
	out.Header.Error = pb.XChainErrorEnum_SUCCESS
	utxos, _, totalSelected, err := bc.Utxovm.SelectUtxos(in.GetAddress(), in.GetPublickey(), totalNeed, in.GetNeedLock(), false)
	if err != nil {
		out.Header.Error = xchaincore.HandlerUtxoError(err)
		s.log.Warn("failed to select utxo", "logid", in.Header.Logid, "error", err.Error())
		return out, nil
	}
	utxoList := []*pb.Utxo{}
	for _, v := range utxos {
		utxo := &pb.Utxo{}
		utxo.RefTxid = v.RefTxid
		utxo.Amount = v.Amount
		utxo.RefOffset = v.RefOffset
		utxo.ToAddr = v.FromAddr
		utxoList = append(utxoList, utxo)
		s.log.Trace("Select utxo list", "refTxid", fmt.Sprintf("%x", v.RefTxid), "refOffset", v.RefOffset, "amount", new(big.Int).SetBytes(v.Amount).String())
	}
	totalSelectedStr := totalSelected.String()
	s.log.Trace("Select utxo totalSelect", "totalSelect", totalSelectedStr)
	out.UtxoList = utxoList
	out.TotalSelected = totalSelectedStr
	return out, nil
}

// DeployNativeCode deploy native contract
func (s *server) DeployNativeCode(ctx context.Context, request *pb.DeployNativeCodeRequest) (*pb.DeployNativeCodeResponse, error) {
	if request.Header == nil {
		request.Header = global.GHeader()
	}
	if !s.mg.Cfg.Native.Enable {
		return nil, errors.New("native module is disabled")
	}

	cfg := s.mg.Cfg.Native.Deploy
	if cfg.WhiteList.Enable {
		found := false
		for _, addr := range cfg.WhiteList.Addresses {
			if addr == request.Address {
				found = true
				break
			}
		}
		if !found {
			return nil, errors.New("permission denied")
		}
	}

	desc := request.GetDesc()
	digest := hash.DoubleSha256(request.Code)
	if !bytes.Equal(digest, desc.Digest) {
		return nil, errors.New("digest not equal")
	}

	// should get blockchain firstly, so we can use CryptoClient
	bc := s.mg.Get(request.GetBcname())
	response := &pb.DeployNativeCodeResponse{Header: &pb.Header{Logid: request.Header.Logid}}
	if bc == nil {
		response.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE // 拒绝
		s.log.Warn("failed to get blockchain before deploy", "logid", request.Header.Logid)
		return response, nil
	}

	pubkey, err := bc.CryptoClient.GetEcdsaPublicKeyFromJSON(request.Pubkey)
	if err != nil {
		return nil, err
	}
	ok, _ := bc.CryptoClient.VerifyAddressUsingPublicKey(request.Address, pubkey)
	if !ok {
		return nil, errors.New("address and public key not match")
	}

	descbuf, _ := proto.Marshal(desc)
	deschash := hash.DoubleSha256(descbuf)
	ok, err = bc.CryptoClient.VerifyECDSA(pubkey, request.Sign, deschash)
	if err != nil || !ok {
		return nil, errors.New("verify sign error")
	}

	err = bc.NativeCodeMgr.Deploy(request.GetDesc(), request.GetCode())
	if err != nil {
		return nil, err
	}

	return response, nil
}

// NativeCodeStatus get native contract status
func (s *server) NativeCodeStatus(ctx context.Context, request *pb.NativeCodeStatusRequest) (*pb.NativeCodeStatusResponse, error) {
	if !s.mg.Cfg.Native.Enable {
		return nil, errors.New("native module is disabled")
	}

	bc := s.mg.Get(request.GetBcname())
	if request.Header == nil {
		request.Header = global.GHeader()
	}
	response := &pb.NativeCodeStatusResponse{Header: &pb.Header{Logid: request.Header.Logid}}
	if bc == nil {
		response.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE // 拒绝
		s.log.Warn("failed to get blockchain before deploy", "logid", request.Header.Logid)
		return response, nil
	}
	response.Status = bc.NativeCodeMgr.Status()
	return response, nil
}

// DposCandidates get dpos candidates
func (s *server) DposCandidates(ctx context.Context, request *pb.DposCandidatesRequest) (*pb.DposCandidatesResponse, error) {
	bc := s.mg.Get(request.GetBcname())
	if request.Header == nil {
		request.Header = global.GHeader()
	}

	response := &pb.DposCandidatesResponse{Header: &pb.Header{Logid: request.Header.Logid}}
	if bc == nil {
		response.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE // 拒绝
		s.log.Warn("DposCandidates failed to get blockchain", "logid", request.Header.Logid)
		return response, nil
	}
	if bc.GetConsType() != consensus.ConsensusTypeTdpos {
		response.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE // 拒绝
		s.log.Warn("DposCandidates failed to check consensus type", "logid", request.Header.Logid)
		return response, errors.New("The consensus is not tdpos")
	}

	candidates, err := bc.GetDposCandidates()
	if err != nil {
		response.Header.Error = pb.XChainErrorEnum_DPOS_QUERY_ERROR
		s.log.Warn("DposCandidates error", "logid", request.Header.Logid, "error", err)
		return response, err
	}
	response.Header.Error = pb.XChainErrorEnum_SUCCESS
	response.CandidatesInfo = candidates
	return response, nil
}

// DposNominateRecords get dpos 提名者提名记录
func (s *server) DposNominateRecords(ctx context.Context, request *pb.DposNominateRecordsRequest) (*pb.DposNominateRecordsResponse, error) {
	bc := s.mg.Get(request.GetBcname())
	if request.Header == nil {
		request.Header = global.GHeader()
	}
	response := &pb.DposNominateRecordsResponse{Header: &pb.Header{Logid: request.Header.Logid}}

	if bc == nil {
		response.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE // 拒绝
		s.log.Warn("DposNominateRecords failed to get blockchain", "logid", request.Header.Logid)
		return response, nil
	}

	if bc.GetConsType() != consensus.ConsensusTypeTdpos {
		response.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE // 拒绝
		s.log.Warn("DposNominateRecords failed to check consensus type", "logid", request.Header.Logid)
		return response, errors.New("The consensus is not tdpos")
	}
	s.log.Info("DposNominateRecords GetDposNominateRecords")
	nominateRecords, err := bc.GetDposNominateRecords(request.Address)
	if err != nil {
		response.Header.Error = pb.XChainErrorEnum_DPOS_QUERY_ERROR
		s.log.Warn("DposNominateRecords error", "logid", request.Header.Logid, "error", err)
		return response, err
	}
	response.Header.Error = pb.XChainErrorEnum_SUCCESS
	response.NominateRecords = nominateRecords
	return response, nil
}

// DposNomineeRecords 候选人被提名记录
func (s *server) DposNomineeRecords(ctx context.Context, request *pb.DposNomineeRecordsRequest) (*pb.DposNomineeRecordsResponse, error) {
	bc := s.mg.Get(request.GetBcname())
	if request.Header == nil {
		request.Header = global.GHeader()
	}
	response := &pb.DposNomineeRecordsResponse{Header: &pb.Header{Logid: request.Header.Logid}}
	if bc == nil {
		response.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE // 拒绝
		s.log.Warn("DposNominatedRecords failed to get blockchain", "logid", request.Header.Logid)
		return response, nil
	}

	if bc.GetConsType() != consensus.ConsensusTypeTdpos {
		response.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE // 拒绝
		s.log.Warn("DposNominatedRecords failed to check consensus type", "logid", request.Header.Logid)
		return response, errors.New("The consensus is not tdpos")
	}

	txid, err := bc.GetDposNominatedRecords(request.Address)
	if err != nil {
		response.Header.Error = pb.XChainErrorEnum_DPOS_QUERY_ERROR
		s.log.Warn("DposNominatedRecords error", "logid", request.Header.Logid, "error", err)
		return response, err
	}
	response.Header.Error = pb.XChainErrorEnum_SUCCESS
	response.Txid = txid
	return response, nil
}

// DposVoteRecords 选民投票记录
func (s *server) DposVoteRecords(ctx context.Context, request *pb.DposVoteRecordsRequest) (*pb.DposVoteRecordsResponse, error) {
	bc := s.mg.Get(request.GetBcname())
	if request.Header == nil {
		request.Header = global.GHeader()
	}
	response := &pb.DposVoteRecordsResponse{Header: &pb.Header{Logid: request.Header.Logid}}
	if bc == nil {
		response.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE // 拒绝
		s.log.Warn("DposVoteRecords failed to get blockchain", "logid", request.Header.Logid)
		return response, nil
	}

	if bc.GetConsType() != consensus.ConsensusTypeTdpos {
		response.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE // 拒绝
		s.log.Warn("DposVoteRecords failed to check consensus type", "logid", request.Header.Logid)
		return response, errors.New("The consensus is not tdpos")
	}

	voteRecords, err := bc.GetDposVoteRecords(request.Address)
	if err != nil {
		response.Header.Error = pb.XChainErrorEnum_DPOS_QUERY_ERROR
		s.log.Warn("DposVoteRecords error", "logid", request.Header.Logid, "error", err)
		return response, err
	}
	response.Header.Error = pb.XChainErrorEnum_SUCCESS
	response.VoteTxidRecords = voteRecords
	return response, nil
}

// DposVotedRecords 候选人被投票记录
func (s *server) DposVotedRecords(ctx context.Context, request *pb.DposVotedRecordsRequest) (*pb.DposVotedRecordsResponse, error) {
	bc := s.mg.Get(request.GetBcname())
	if request.Header == nil {
		request.Header = global.GHeader()
	}
	response := &pb.DposVotedRecordsResponse{Header: &pb.Header{Logid: request.Header.Logid}}
	if bc == nil {
		response.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE // 拒绝
		s.log.Warn("DposVotedRecords failed to get blockchain", "logid", request.Header.Logid)
		return response, nil
	}

	if bc.GetConsType() != consensus.ConsensusTypeTdpos {
		response.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE // 拒绝
		s.log.Warn("DposVoteRecords failed to check consensus type", "logid", request.Header.Logid)
		return response, errors.New("The consensus is not tdpos")
	}
	votedRecords, err := bc.GetDposVotedRecords(request.Address)
	if err != nil {
		response.Header.Error = pb.XChainErrorEnum_DPOS_QUERY_ERROR
		s.log.Warn("GetDposVotedRecords error", "logid", request.Header.Logid, "error", err)
		return response, err
	}
	response.Header.Error = pb.XChainErrorEnum_SUCCESS
	response.VotedTxidRecords = votedRecords
	return response, nil
}

// DposCheckResults get dpos 检查结果
func (s *server) DposCheckResults(ctx context.Context, request *pb.DposCheckResultsRequest) (*pb.DposCheckResultsResponse, error) {
	bc := s.mg.Get(request.GetBcname())
	if request.Header == nil {
		request.Header = global.GHeader()
	}

	response := &pb.DposCheckResultsResponse{Header: &pb.Header{Logid: request.Header.Logid}}
	if bc == nil {
		response.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE // 拒绝
		s.log.Warn("DposVotedRecords failed to get blockchain", "logid", request.Header.Logid)
		return response, nil
	}

	if bc.GetConsType() != consensus.ConsensusTypeTdpos {
		response.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE // 拒绝
		s.log.Warn("DposVoteRecords failed to check consensus type", "logid", request.Header.Logid)
		return response, errors.New("The consensus is not tdpos")
	}

	checkResult, err := bc.GetCheckResults(request.Term)
	if err != nil {
		response.Header.Error = pb.XChainErrorEnum_DPOS_QUERY_ERROR
		s.log.Warn("DposCheckResults error", "logid", request.Header.Logid, "error", err)
		return response, err
	}
	response.Term = request.Term
	response.CheckResult = checkResult
	return response, nil
}

// DposStatus get dpos current status
func (s *server) DposStatus(ctx context.Context, request *pb.DposStatusRequest) (*pb.DposStatusResponse, error) {
	bc := s.mg.Get(request.GetBcname())
	if request.Header == nil {
		request.Header = global.GHeader()
	}
	response := &pb.DposStatusResponse{Header: &pb.Header{Logid: request.Header.Logid}, Status: &pb.DposStatus{}}
	if bc == nil {
		response.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE
		s.log.Warn("DposStatus failed  to get blockchain", "logid", request.Header.Logid)
		return response, nil
	}

	if bc.GetConsType() != consensus.ConsensusTypeTdpos {
		response.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE
		s.log.Warn("DposStatus failed to check consensus type", "logid", request.Header.Logid)
		return response, errors.New("The consensus is not tdpos")
	}

	status := bc.GetConsStatus()
	response.Status.Term = status.Term
	response.Status.BlockNum = status.BlockNum
	response.Status.Proposer = status.Proposer
	checkResult, err := bc.GetCheckResults(status.Term)
	if err != nil {
		response.Header.Error = pb.XChainErrorEnum_DPOS_QUERY_ERROR
		s.log.Warn("DposStatus error", "logid", request.Header.Logid, "error", err)
		return response, err
	}
	response.Status.CheckResult = checkResult
	response.Status.ProposerNum = int64(len(checkResult))
	return response, nil
}

// PreExecWithSelectUTXO preExec + selectUtxo
func (s *server) PreExecWithSelectUTXO(ctx context.Context, request *pb.PreExecWithSelectUTXORequest) (*pb.PreExecWithSelectUTXOResponse, error) {
	// verify input param
	if request == nil {
		return nil, errors.New("request is invalid")
	}
	if request.Header == nil {
		request.Header = global.GHeader()
	}

	// initialize output
	responses := &pb.PreExecWithSelectUTXOResponse{Header: &pb.Header{Logid: request.Header.Logid}}
	responses.Bcname = request.GetBcname()
	// for PreExec
	preExecRequest := request.GetRequest()
	fee := int64(0)
	if preExecRequest != nil {
		preExecRequest.Header = request.Header
		invokeRPCResponse, preErr := s.PreExec(ctx, preExecRequest)
		if preErr != nil {
			return nil, preErr
		}
		invokeResponse := invokeRPCResponse.GetResponse()
		responses.Response = invokeResponse
		fee = responses.Response.GetGasUsed()
	}

	totalAmount := request.GetTotalAmount() + fee

	if totalAmount > 0 {
		utxoInput := &pb.UtxoInput{
			Bcname:    request.GetBcname(),
			Address:   request.GetAddress(),
			TotalNeed: strconv.FormatInt(totalAmount, 10),
			Publickey: request.GetSignInfo().GetPublicKey(),
			UserSign:  request.GetSignInfo().GetSign(),
			NeedLock:  request.GetNeedLock(),
		}
		if ok := validUtxoAccess(utxoInput, s.mg.Get(utxoInput.GetBcname())); !ok {
			return nil, errors.New("validUtxoAccess failed")
		}
		utxoOutput, selectErr := s.SelectUTXO(ctx, utxoInput)
		if selectErr != nil {
			return nil, selectErr
		}
		if utxoOutput.Header.Error != pb.XChainErrorEnum_SUCCESS {
			return nil, common.ServerError{utxoOutput.Header.Error}
		}
		responses.UtxoOutput = utxoOutput
	}

	return responses, nil
}

// PreExec smart contract preExec process
func (s *server) PreExec(ctx context.Context, request *pb.InvokeRPCRequest) (*pb.InvokeRPCResponse, error) {
	s.log.Trace("Got PreExec req", "req", request)
	bc := s.mg.Get(request.GetBcname())
	if request.Header == nil {
		request.Header = global.GHeader()
	}
	rsps := &pb.InvokeRPCResponse{Header: &pb.Header{Logid: request.Header.Logid}}
	if bc == nil {
		rsps.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE // 拒绝
		s.log.Warn("failed to get blockchain before query", "logid", request.Header.Logid)
		return rsps, nil
	}
	hd := &global.XContext{Timer: global.NewXTimer()}
	vmResponse, err := bc.PreExec(request, hd)
	if err != nil {
		return nil, err
	}
	txInputs := vmResponse.GetInputs()
	for _, txInput := range txInputs {
		if bc.QueryTxFromForbidden(txInput.GetRefTxid()) {
			return rsps, errors.New("RefTxid has been forbidden")
		}
	}
	rsps.Response = vmResponse
	s.log.Info("PreExec", "logid", request.Header.Logid, "cost", hd.Timer.Print())
	return rsps, nil
}

// GetBlockByHeight  get trunk block by height
func (s *server) GetBlockByHeight(ctx context.Context, in *pb.BlockHeight) (*pb.Block, error) {
	if in.Header == nil {
		in.Header = global.GHeader()
	}
	s.log.Trace("Start to get dealwith GetBlockByHeight", "logid", in.Header.Logid, "bcname", in.Bcname, "height", in.Height)
	bc := s.mg.Get(in.Bcname)
	if bc == nil {
		out := pb.Block{Header: &pb.Header{}}
		out.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE // 拒绝
		return &out, nil
	}
	out := bc.GetBlockByHeight(in)

	block := out.GetBlock()
	transactions := block.GetTransactions()
	transactionsFilter := []*pb.Transaction{}
	for _, transaction := range transactions {
		txid := transaction.GetTxid()
		if bc.QueryTxFromForbidden(txid) {
			continue
		}
		transactionsFilter = append(transactionsFilter, transaction)
	}
	if transactions != nil {
		out.Block.Transactions = transactionsFilter
	}

	s.log.Trace("GetBlockByHeight result", "logid", in.Header.Logid, "bcname", in.Bcname, "height", in.Height,
		"blockid", out.GetBlockid())
	return out, nil
}

func (s *server) GetAccountByAK(ctx context.Context, request *pb.AK2AccountRequest) (*pb.AK2AccountResponse, error) {
	if request.Header == nil {
		request.Header = global.GHeader()
	}
	bc := s.mg.Get(request.Bcname)
	if bc == nil {
		out := pb.AK2AccountResponse{Header: &pb.Header{}}
		out.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE // 拒绝
		return &out, nil
	}
	out := &pb.AK2AccountResponse{
		Bcname: request.Bcname,
		Header: global.GHeader(),
	}
	accounts, err := bc.QueryAccountContainAK(request.GetAddress())
	if err != nil || accounts == nil {
		return out, err
	}
	out.Account = accounts
	return out, err
}

func startTCPServer(xchainmg *xchaincore.XChainMG) error {
	var (
		cfg   = xchainmg.Cfg
		log   = xchainmg.Log
		isTLS = cfg.TCPServer.TLS
		svr   = server{log: log, mg: xchainmg, dedupCache: common.NewLRUCache(cfg.DedupCacheSize), dedupTimeLimit: cfg.DedupTimeLimit}

		rpcOptions []grpc.ServerOption
	)

	if cfg.TCPServer.MetricPort != "" {
		// add prometheus support
		rpcOptions = append(rpcOptions,
			grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
			grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
		)

	}

	rpcOptions = append(rpcOptions,
		grpc.MaxMsgSize(cfg.TCPServer.MaxMsgSize),
		grpc.ReadBufferSize(cfg.TCPServer.ReadBufferSize),
		grpc.InitialWindowSize(cfg.TCPServer.InitialWindowSize),
		grpc.InitialConnWindowSize(cfg.TCPServer.InitialConnWindowSize),
		grpc.WriteBufferSize(cfg.TCPServer.WriteBufferSize),
	)

	if isTLS {
		log.Trace("start tls rpc server")
		tlsPath := cfg.TCPServer.TLSPath
		bs, err := ioutil.ReadFile(tlsPath + "/cert.crt")
		if err != nil {
			return err
		}
		certPool := x509.NewCertPool()
		ok := certPool.AppendCertsFromPEM(bs)
		if !ok {
			return err
		}

		certificate, err := tls.LoadX509KeyPair(tlsPath+"/key.pem", tlsPath+"/private.key")

		creds := credentials.NewTLS(
			&tls.Config{
				ServerName:   cfg.TCPServer.MServerName,
				Certificates: []tls.Certificate{certificate},
				RootCAs:      certPool,
				ClientCAs:    certPool,
				ClientAuth:   tls.RequireAndVerifyClientCert,
			})

		l, err := net.Listen("tcp", cfg.TCPServer.HTTPSPort)
		if err != nil {
			return err
		}
		// Copy rpc options
		options := append([]grpc.ServerOption{}, rpcOptions...)
		options = append(options, grpc.Creds(creds))
		s := grpc.NewServer(options...)
		pb.RegisterXchainServer(s, &svr)
		reflection.Register(s)
		log.Trace("start tls rpc server")
		go func() {
			err := s.Serve(l)
			if err != nil {
				panic(err)
			}
		}()
	}

	log.Trace("start rpc server")
	s := grpc.NewServer(rpcOptions...)
	pb.RegisterXchainServer(s, &svr)
	if cfg.TCPServer.MetricPort != "" {
		grpc_prometheus.EnableHandlingTimeHistogram(
			grpc_prometheus.WithHistogramBuckets([]float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}),
		)
		// 考虑到tls rpc server跟普通server共用一个XchainService，所以只需要注册一次prometheus就行
		// 因为两者的ServerOption都包含了prometheus的拦截器，所以都可以监控到
		// Must be called after RegisterXchainServer
		grpc_prometheus.Register(s)
		http.Handle("/metrics", promhttp.Handler())
		go func() {
			panic(http.ListenAndServe(cfg.TCPServer.MetricPort, nil))
		}()
	}

	lis, err := net.Listen("tcp", cfg.TCPServer.Port)
	if err != nil {
		log.Error("failed to listen: ", "error", err.Error())
		return err
	}
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Error("failed to serve: ", "error", err.Error())
		return err
	}
	return nil
}

// SerRun xchain server start entrance
func SerRun(xchainmg *xchaincore.XChainMG) {
	err := startTCPServer(xchainmg)
	if err != nil {
		close(xchainmg.Quit)
	}
}
