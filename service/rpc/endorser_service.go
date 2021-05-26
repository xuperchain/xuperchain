package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"sync"
	"time"

	"google.golang.org/grpc"

	scom "github.com/xuperchain/xuperchain/service/common"
	sconf "github.com/xuperchain/xuperchain/service/config"
	"github.com/xuperchain/xuperchain/service/pb"
	"github.com/xuperchain/xupercore/bcs/ledger/xledger/state/utxo/txhash"
	sctx "github.com/xuperchain/xupercore/example/xchain/common/context"
	ecom "github.com/xuperchain/xupercore/kernel/engines/xuperos/common"
	crypto_client "github.com/xuperchain/xupercore/lib/crypto/client"
	"github.com/xuperchain/xupercore/lib/crypto/hash"
)

const (
	EndorserModuleDefault = "default"
	EndorserModuleProxy   = "proxy"
)

type XEndorser interface {
	EndorserCall(gctx context.Context, req *pb.EndorserRequest) (*pb.EndorserResponse, error)
}

type ProxyXEndorser struct {
	engine      ecom.Engine
	clientCache sync.Map
	mutex       sync.Mutex
	conf        *sconf.ServConf
}

func newEndorserService(cfg *sconf.ServConf, engine ecom.Engine, svr XEndorserServer) (XEndorser, error) {
	switch cfg.EndorserModule {
	case EndorserModuleDefault:
		dxe := NewDefaultXEndorser(svr)
		return dxe, nil
	case EndorserModuleProxy:
		return &ProxyXEndorser{
			engine: engine,
			conf:   cfg,
		}, nil
	default:
		return nil, fmt.Errorf("unknown endorser module")
	}

}

func (pxe *ProxyXEndorser) EndorserCall(gctx context.Context, req *pb.EndorserRequest) (*pb.EndorserResponse, error) {
	resp := &pb.EndorserResponse{}
	rctx := sctx.ValueReqCtx(gctx)
	endc, err := pxe.getClient(pxe.getHost())
	if err != nil {
		return resp, err
	}
	res, err := endc.EndorserCall(gctx, req)
	if err != nil {
		return resp, err
	}
	resp.EndorserAddress = res.EndorserAddress
	resp.ResponseName = res.ResponseName
	resp.ResponseData = res.ResponseData
	resp.EndorserSign = res.EndorserSign
	rctx.GetLog().SetInfoField("bc_name", req.GetBcName())
	rctx.GetLog().SetInfoField("request_name", req.GetBcName())
	return resp, nil
}

func (pxe *ProxyXEndorser) getHost() string {
	host := ""
	hostCnt := len(pxe.conf.EndorserHosts)
	if hostCnt > 0 {
		rand.Seed(time.Now().Unix())
		index := rand.Intn(hostCnt)
		host = pxe.conf.EndorserHosts[index]
	}
	return host
}

func (pxe *ProxyXEndorser) getClient(host string) (pb.XendorserClient, error) {
	if host == "" {
		return nil, fmt.Errorf("empty host")
	}
	if c, ok := pxe.clientCache.Load(host); ok {
		return c.(pb.XendorserClient), nil
	}

	pxe.mutex.Lock()
	defer pxe.mutex.Unlock()
	if c, ok := pxe.clientCache.Load(host); ok {
		return c.(pb.XendorserClient), nil
	}
	conn, err := grpc.Dial(host, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	c := pb.NewXendorserClient(conn)
	pxe.clientCache.Store(host, c)
	return c, nil
}

type XEndorserServer interface {
	// PostTx post Transaction to a node
	PostTx(context.Context, *pb.TxStatus) (*pb.CommonReply, error)
	// QueryTx query Transaction by TxStatus,
	// Bcname and Txid are required for this
	QueryTx(context.Context, *pb.TxStatus) (*pb.TxStatus, error)
	// PreExecWithSelectUTXO preExec & selectUtxo
	PreExecWithSelectUTXO(context.Context, *pb.PreExecWithSelectUTXORequest) (*pb.PreExecWithSelectUTXOResponse, error)
	// 预执行合约
	PreExec(context.Context, *pb.InvokeRPCRequest) (*pb.InvokeRPCResponse, error)
}

type DefaultXEndorser struct {
	svr         XEndorserServer
	requestType map[string]bool
}

var _ XEndorser = (*DefaultXEndorser)(nil)

const (
	// DefaultKeyPath is the default key path
	DefaultKeyPath = "./data/endorser/keys/"
)

func NewDefaultXEndorser(svr XEndorserServer) *DefaultXEndorser {
	return &DefaultXEndorser{
		requestType: map[string]bool{
			"PreExecWithFee":    true,
			"ComplianceCheck":   true,
			"CrossQueryPreExec": true,
			"TxQuery":           true,
		},
		svr: svr,
	}
}

// EndorserCall process endorser call
func (dxe *DefaultXEndorser) EndorserCall(ctx context.Context, req *pb.EndorserRequest) (*pb.EndorserResponse, error) {
	// make response header
	resHeader := &pb.Header{
		Error: pb.XChainErrorEnum_SUCCESS,
	}
	if req.GetHeader() == nil {
		resHeader.Logid = req.GetHeader().GetLogid()
	}

	// check param
	if _, ok := dxe.requestType[req.GetRequestName()]; !ok {
		resHeader.Error = pb.XChainErrorEnum_SERVICE_REFUSED_ERROR
		return dxe.generateErrorResponse(req, resHeader, errors.New("request name not supported"))
	}

	switch req.GetRequestName() {
	case "ComplianceCheck":
		success, errcode, err := dxe.processFee(ctx, req)
		if err != nil || !success {
			resHeader.Error = errcode
			return dxe.generateErrorResponse(req, resHeader, err)
		}
		addr, sign, err := dxe.generateTxSign(ctx, req)
		if err != nil {
			resHeader.Error = pb.XChainErrorEnum_SERVICE_REFUSED_ERROR
			return dxe.generateErrorResponse(req, resHeader, err)
		}

		reply := &pb.CommonReply{
			Header: &pb.Header{
				Error: pb.XChainErrorEnum_SUCCESS,
			},
		}
		resData, err := json.Marshal(reply)
		if err != nil {
			resHeader.Error = pb.XChainErrorEnum_SERVICE_REFUSED_ERROR
			return dxe.generateErrorResponse(req, resHeader, err)
		}
		return dxe.generateSuccessResponse(req, resData, addr, sign, resHeader)

	case "PreExecWithFee":
		resData, errcode, err := dxe.getPreExecResult(ctx, req)
		if err != nil {
			resHeader.Error = errcode
			return dxe.generateErrorResponse(req, resHeader, err)
		}
		return dxe.generateSuccessResponse(req, resData, nil, nil, resHeader)

	case "CrossQueryPreExec":
		resData, errcode, err := dxe.getCrossQueryResult(ctx, req)
		if err != nil {
			resHeader.Error = errcode
			return dxe.generateErrorResponse(req, resHeader, err)
		}
		data := append(req.RequestData[:], resData[:]...)
		digest := hash.UsingSha256(data)
		addr, sign, err := dxe.signData(ctx, digest, DefaultKeyPath)
		if err != nil {
			resHeader.Error = errcode
			return dxe.generateErrorResponse(req, resHeader, err)
		}
		return dxe.generateSuccessResponse(req, resData, addr, sign, resHeader)
	case "TxQuery":
		resData, errcode, err := dxe.getTxResult(ctx, req)
		if err != nil {
			resHeader.Error = errcode
			return dxe.generateErrorResponse(req, resHeader, err)
		}
		data := append(req.RequestData[:], resData[:]...)
		digest := hash.UsingSha256(data)
		addr, sign, err := dxe.signData(ctx, digest, DefaultKeyPath)
		if err != nil {
			resHeader.Error = errcode
			return dxe.generateErrorResponse(req, resHeader, err)
		}
		return dxe.generateSuccessResponse(req, resData, addr, sign, resHeader)
	}

	return nil, nil
}

func (dxe *DefaultXEndorser) getPreExecResult(ctx context.Context, req *pb.EndorserRequest) ([]byte, pb.XChainErrorEnum, error) {
	request := &pb.PreExecWithSelectUTXORequest{}
	err := json.Unmarshal(req.GetRequestData(), request)
	if err != nil {
		return nil, pb.XChainErrorEnum_SERVICE_REFUSED_ERROR, err
	}

	res, err := dxe.svr.PreExecWithSelectUTXO(ctx, request)
	if err != nil {
		return nil, res.GetHeader().GetError(), err
	}

	sData, err := json.Marshal(res)
	if err != nil {
		return nil, pb.XChainErrorEnum_SERVICE_REFUSED_ERROR, err
	}
	return sData, pb.XChainErrorEnum_SUCCESS, nil
}

func (dxe *DefaultXEndorser) getCrossQueryResult(ctx context.Context, req *pb.EndorserRequest) ([]byte, pb.XChainErrorEnum, error) {
	cqReq := &pb.CrossQueryRequest{}
	err := json.Unmarshal(req.GetRequestData(), cqReq)
	if err != nil {
		return nil, pb.XChainErrorEnum_SERVICE_REFUSED_ERROR, err
	}

	preExecReq := &pb.InvokeRPCRequest{
		Header:      req.GetHeader(),
		Bcname:      cqReq.GetBcname(),
		Initiator:   cqReq.GetInitiator(),
		AuthRequire: cqReq.GetAuthRequire(),
	}
	preExecReq.Requests = append(preExecReq.Requests, cqReq.GetRequest())

	preExecRes, err := dxe.svr.PreExec(ctx, preExecReq)
	if err != nil {
		return nil, preExecRes.GetHeader().GetError(), err
	}

	if preExecRes.GetHeader().GetError() != pb.XChainErrorEnum_SUCCESS {
		return nil, preExecRes.GetHeader().GetError(), errors.New("PreExec not success")
	}

	res := &pb.CrossQueryResponse{}
	contractRes := preExecRes.GetResponse().GetResponses()
	if len(contractRes) > 0 {
		res.Response = contractRes[len(contractRes)-1]
	}

	sData, err := json.Marshal(res)
	if err != nil {
		return nil, pb.XChainErrorEnum_SERVICE_REFUSED_ERROR, err
	}

	return sData, pb.XChainErrorEnum_SUCCESS, nil
}

func (dxe *DefaultXEndorser) getTxResult(ctx context.Context, req *pb.EndorserRequest) ([]byte, pb.XChainErrorEnum, error) {
	request := &pb.TxStatus{}
	err := json.Unmarshal(req.GetRequestData(), request)
	if err != nil {
		return nil, pb.XChainErrorEnum_SERVICE_REFUSED_ERROR, err
	}

	reply, err := dxe.svr.QueryTx(ctx, request)
	if err != nil {
		return nil, reply.GetHeader().GetError(), err
	}

	if reply.GetHeader().GetError() != pb.XChainErrorEnum_SUCCESS {
		return nil, reply.GetHeader().GetError(), errors.New("QueryTx not success")
	}

	if reply.Tx == nil {
		return nil, reply.GetHeader().GetError(), errors.New("tx not found")
	}

	sData, err := json.Marshal(reply.Tx)
	if err != nil {
		return nil, pb.XChainErrorEnum_SERVICE_REFUSED_ERROR, err
	}

	return sData, pb.XChainErrorEnum_SUCCESS, nil
}

func (dxe *DefaultXEndorser) processFee(ctx context.Context, req *pb.EndorserRequest) (bool, pb.XChainErrorEnum, error) {
	if req.GetFee() == nil {
		// no fee provided, default to true
		return true, pb.XChainErrorEnum_SUCCESS, nil
	}

	txStatus := &pb.TxStatus{
		Txid:   req.GetFee().GetTxid(),
		Bcname: req.GetBcName(),
		Tx:     req.GetFee(),
	}

	res, err := dxe.svr.PostTx(ctx, txStatus)
	if err != nil {
		return false, res.GetHeader().GetError(), err
	} else if res.GetHeader().GetError() != pb.XChainErrorEnum_SUCCESS {
		return false, res.GetHeader().GetError(), errors.New("Fee post to chain failed")
	}

	return true, pb.XChainErrorEnum_SUCCESS, nil
}

func (dxe *DefaultXEndorser) generateTxSign(ctx context.Context, req *pb.EndorserRequest) ([]byte, *pb.SignatureInfo, error) {
	if req.GetRequestData() == nil {
		return nil, nil, errors.New("request data is empty")
	}

	txStatus := &pb.TxStatus{}
	err := json.Unmarshal(req.GetRequestData(), txStatus)
	if err != nil {
		return nil, nil, err
	}

	tx := scom.TxToXledger(txStatus.GetTx())
	digest, err := txhash.MakeTxDigestHash(tx)
	if err != nil {
		return nil, nil, err
	}

	return dxe.signData(ctx, digest, DefaultKeyPath)
}

func (dxe *DefaultXEndorser) signData(ctx context.Context, data []byte, keypath string) ([]byte, *pb.SignatureInfo, error) {
	addr, jsonSKey, jsonAKey, err := dxe.getEndorserKey(keypath)
	if err != nil {
		return nil, nil, err
	}

	cryptoClient, err := crypto_client.CreateCryptoClientFromJSONPrivateKey(jsonSKey)
	if err != nil {
		return nil, nil, err
	}

	privateKey, err := cryptoClient.GetEcdsaPrivateKeyFromJsonStr(string(jsonSKey))
	if err != nil {
		return nil, nil, err
	}

	sign, err := cryptoClient.SignECDSA(privateKey, data)
	if err != nil {
		return nil, nil, err
	}

	signInfo := &pb.SignatureInfo{
		PublicKey: string(jsonAKey),
		Sign:      sign,
	}
	return addr, signInfo, nil
}

func (dxe *DefaultXEndorser) generateErrorResponse(req *pb.EndorserRequest, header *pb.Header,
	err error) (*pb.EndorserResponse, error) {
	res := &pb.EndorserResponse{
		Header:       header,
		ResponseName: req.GetRequestName(),
	}
	return res, err
}

func (dxe *DefaultXEndorser) generateSuccessResponse(req *pb.EndorserRequest, resData []byte,
	addr []byte, sign *pb.SignatureInfo, header *pb.Header) (*pb.EndorserResponse, error) {
	res := &pb.EndorserResponse{
		Header:          header,
		ResponseName:    req.GetRequestName(),
		ResponseData:    resData,
		EndorserAddress: string(addr),
		EndorserSign:    sign,
	}
	return res, nil
}

func (dxe *DefaultXEndorser) getEndorserKey(keypath string) ([]byte, []byte, []byte, error) {
	sk, err := ioutil.ReadFile(keypath + "private.key")
	if err != nil {
		return nil, nil, nil, err
	}

	ak, err := ioutil.ReadFile(keypath + "public.key")
	if err != nil {
		return nil, nil, nil, err
	}

	addr, err := ioutil.ReadFile(keypath + "address")
	return addr, sk, ak, err
}
