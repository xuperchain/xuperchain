package server

import (
	"encoding/json"
	"errors"
	"io/ioutil"

	context "golang.org/x/net/context"

	crypto_client "github.com/xuperchain/xuperchain/core/crypto/client"
	"github.com/xuperchain/xuperchain/core/crypto/hash"
	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/utxo/txhash"
)

// XEndorser is the interface for endorser service
// Endorser protocol provide standard interface for endorser operations.
// In many cases, a full node could be used for Endorser service.
// For example, an endorser service can provide delegated computing and
// compliance check for transactions, and set an endorser fee for each request.
// endorser service provider could decide how much fee needed for each operation.
type XEndorser interface {
	EndorserCall(ctx context.Context, req *pb.EndorserRequest) (*pb.EndorserResponse, error)
}

// DefaultXEndorser default implementation of XEndorser
// Endorser service can implement the interface in their own way and follow
// the protocol defined in xendorser.proto.
type DefaultXEndorser struct {
	svr         *server
	requestType map[string]bool
}

var (
	// DefaultKeyPath is the default key path
	DefaultKeyPath = "./data/endorser/keys/"

	// NodeKeyPath is the key path of xchain node
	NodeKeyPath = "./data/keys/"
)

// NewDefaultXEndorser create instance of DefaultXEndorser
func NewDefaultXEndorser(svr *server) *DefaultXEndorser {
	return &DefaultXEndorser{
		svr: svr,
		requestType: map[string]bool{
			"PreExecWithFee":    true,
			"ComplianceCheck":   true,
			"CrossQueryPreExec": true,
		},
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

	digest, err := txhash.MakeTxDigestHash(txStatus.GetTx())
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

	privateKey, err := cryptoClient.GetEcdsaPrivateKeyFromJSON(jsonSKey)
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
